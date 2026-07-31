package handler

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"everyday_keywords/internal/collector"
	"everyday_keywords/internal/db"
	"everyday_keywords/internal/logger"
)

//go:embed admin.html
var adminHTML string

type AdminHandler struct {
	db         *db.DB
	collector  *collector.Collector
	logger     *logger.Logger
	configPath string
	config     *AdminConfig
}

type AdminConfig struct {
	Port        int    `json:"port"`
	DBPath      string `json:"db_path"`
	LogFile     string `json:"log_file"`
	WebhookURL  string `json:"webhook_url"`
	WebhookHMAC string `json:"webhook_hmac"`
	AdminToken  string `json:"admin_token"`
}

func NewAdmin(database *db.DB, coll *collector.Collector, l *logger.Logger, configPath string) *AdminHandler {
	a := &AdminHandler{
		db:         database,
		collector:  coll,
		logger:     l,
		configPath: configPath,
		config:     &AdminConfig{},
	}
	a.loadConfigFromFile()
	return a
}

func (a *AdminHandler) loadConfigFromFile() {
	data, err := os.ReadFile(a.configPath)
	if err != nil {
		return
	}
	var cfg AdminConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return
	}
	a.config = &cfg
}

func (a *AdminHandler) ServeAdmin(w http.ResponseWriter, r *http.Request) {
	token := a.config.AdminToken
	if token != "" {
		queryToken := r.URL.Query().Get("token")
		cookieToken, _ := r.Cookie("admin_token")
		ok := queryToken == token || (cookieToken != nil && cookieToken.Value == token)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(adminHTML))
}

func (a *AdminHandler) HandleStats(w http.ResponseWriter, r *http.Request) {
	stats, err := a.db.GetStats()
	if err != nil {
		writeAdminJSON(w, 500, "获取统计失败: "+err.Error(), nil)
		return
	}
	writeAdminJSON(w, 200, "ok", stats)
}

func (a *AdminHandler) HandleKeywords(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	if size < 1 || size > 200 {
		size = 50
	}
	source := r.URL.Query().Get("source")
	status := r.URL.Query().Get("status")
	search := r.URL.Query().Get("search")

	keywords, total, err := a.db.ListKeywords(page, size, source, status, search)
	if err != nil {
		writeAdminJSON(w, 500, "查询失败: "+err.Error(), nil)
		return
	}

	result := map[string]interface{}{
		"keywords": keywords,
		"total":    total,
		"page":     page,
		"size":     size,
	}
	writeAdminJSON(w, 200, "ok", result)
}

func (a *AdminHandler) HandleDeleteKeywords(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		IDs []int64 `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAdminJSON(w, 400, "无效的请求: "+err.Error(), nil)
		return
	}
	if len(req.IDs) == 0 {
		writeAdminJSON(w, 400, "请选择要删除的关键字", nil)
		return
	}

	deleted, err := a.db.DeleteByIDs(req.IDs)
	if err != nil {
		writeAdminJSON(w, 500, "删除失败: "+err.Error(), nil)
		return
	}

	writeAdminJSON(w, 200, "删除成功", map[string]int{"deleted": deleted})
}

func (a *AdminHandler) HandleLogs(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	if size < 1 || size > 200 {
		size = 50
	}
	status := r.URL.Query().Get("status")

	logPage, err := a.logger.ReadLogs(page, size, status)
	if err != nil {
		writeAdminJSON(w, 500, "读取日志失败: "+err.Error(), nil)
		return
	}

	writeAdminJSON(w, 200, "ok", logPage)
}

func (a *AdminHandler) HandleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		data, err := os.ReadFile(a.configPath)
		if err != nil {
			writeAdminJSON(w, 500, "读取配置失败: "+err.Error(), nil)
			return
		}
		writeAdminJSON(w, 200, "ok", json.RawMessage(data))
	case http.MethodPut:
		var newCfg map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&newCfg); err != nil {
			writeAdminJSON(w, 400, "无效的 JSON: "+err.Error(), nil)
			return
		}
		data, err := json.MarshalIndent(newCfg, "", "  ")
		if err != nil {
			writeAdminJSON(w, 500, "序列化失败: "+err.Error(), nil)
			return
		}
		if err := os.WriteFile(a.configPath, data, 0644); err != nil {
			writeAdminJSON(w, 500, "保存配置失败: "+err.Error(), nil)
			return
		}
		a.loadConfigFromFile()
		writeAdminJSON(w, 200, "配置已保存", nil)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *AdminHandler) HandleCollect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	count, _ := strconv.Atoi(r.URL.Query().Get("count"))
	if count < 1 {
		count = 100
	}

	results, errs := a.collector.CollectUntil(count)

	result := map[string]interface{}{
		"sources": len(results),
		"errors":  len(errs),
	}
	if len(errs) > 0 {
		var msgs []string
		for _, e := range errs {
			msgs = append(msgs, e.Error())
		}
		result["error_details"] = msgs
	}

	writeAdminJSON(w, 200, "采集完成", result)
}

type adminResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func writeAdminJSON(w http.ResponseWriter, code int, msg string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(adminResponse{Code: code, Msg: msg, Data: data})
}