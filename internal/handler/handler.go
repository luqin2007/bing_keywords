package handler

import (
	"encoding/json"
	"log"
	"math"
	"net"
	"net/http"
	"strconv"
	"time"

	"everyday_keywords/internal/collector"
	"everyday_keywords/internal/db"
	"everyday_keywords/internal/logger"
	"everyday_keywords/internal/notifier"
)

type Handler struct {
	db        *db.DB
	collector *collector.Collector
	notifier  *notifier.Notifier
	logger    *logger.Logger
}

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data []TitleItem `json:"data"`
}

type TitleItem struct {
	Title string `json:"title"`
}

func New(database *db.DB, coll *collector.Collector, n *notifier.Notifier, l *logger.Logger) *Handler {
	return &Handler{
		db:        database,
		collector: coll,
		notifier:  n,
		logger:    l,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, "仅支持 GET 请求", nil)
		return
	}

	countStr := r.URL.Query().Get("count")
	count := 5
	if countStr != "" {
		n, err := strconv.Atoi(countStr)
		if err != nil || n <= 0 || n > 100 {
			writeJSON(w, http.StatusBadRequest, "count 参数无效，范围 1-100", nil)
			return
		}
		count = n
	}

	available, err := h.db.CountAvailable()
	if err != nil {
		log.Printf("[handler] count available error: %v", err)
		writeJSON(w, http.StatusInternalServerError, "服务器内部错误", nil)
		return
	}

	lastCntStr, _ := h.db.GetMeta("last_request_count")
	lastCnt := count
	if lastCntStr != "" {
		lastCnt, _ = strconv.Atoi(lastCntStr)
	}

	threshold := int(math.Max(float64(count), float64(lastCnt))) * 3

	if available < threshold {
		log.Printf("[handler] pool low: available=%d, threshold=%d, collecting...", available, threshold)
		_, errs := h.collector.CollectUntil(threshold)

		if len(errs) > 0 {
			newAvailable, _ := h.db.CountAvailable()
			if newAvailable < count {
				sourceNames := []string{"github_repos", "github_devs", "sourceforge", "openrouter", "steam", "itch"}
				var errMsgs []string
				for _, e := range errs {
					errMsgs = append(errMsgs, e.Error())
				}
				h.notifier.SendAlert(sourceNames, errMsgs, newAvailable, threshold, lastCnt)
			}
		}
	}

	keywords, err := h.db.PickRandom(count)
	if err != nil {
		log.Printf("[handler] pick random error: %v", err)
		writeJSON(w, http.StatusInternalServerError, "服务器内部错误", nil)
		return
	}

	if len(keywords) < count {
		writeJSON(w, http.StatusOK, "关键字不足，已返回全部可用关键字", toTitleItems(keywords))
		h.logRequest(r, count, keywords, start, "partial", "")
		return
	}

	var ids []int64
	for _, k := range keywords {
		ids = append(ids, k.ID)
	}
	if err := h.db.MarkUsed(ids); err != nil {
		log.Printf("[handler] mark used error: %v", err)
	}

	h.db.SetMeta("last_request_count", strconv.Itoa(count))

	writeJSON(w, http.StatusOK, "请求成功", toTitleItems(keywords))
	h.logRequest(r, count, keywords, start, "success", "")
}

func (h *Handler) logRequest(r *http.Request, count int, keywords []db.Keyword, start time.Time, status, errStr string) {
	var titles []string
	for _, k := range keywords {
		titles = append(titles, k.Word)
	}
	ip := r.RemoteAddr
	if h, _, err := net.SplitHostPort(ip); err == nil {
		ip = h
	}
	h.logger.Write(logger.LogEntry{
		Type:      "request",
		Method:    r.Method,
		Path:      r.URL.String(),
		IP:        ip,
		UserAgent: r.UserAgent(),
		Count:    count,
		Keywords: titles,
		DurMs:    time.Since(start).Milliseconds(),
		Status:   status,
		Error:    errStr,
	})
}

func toTitleItems(keywords []db.Keyword) []TitleItem {
	items := make([]TitleItem, len(keywords))
	for i, k := range keywords {
		items[i] = TitleItem{Title: k.Word}
	}
	return items
}

func writeJSON(w http.ResponseWriter, code int, msg string, data []TitleItem) {
	resp := Response{Code: code, Msg: msg, Data: data}
	if data == nil {
		resp.Data = []TitleItem{}
	}
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(resp)
}