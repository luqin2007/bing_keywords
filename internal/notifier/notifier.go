package notifier

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Notifier struct {
	webhookURL string
	hmacSecret string
	client     *http.Client
}

func New(webhookURL, hmacSecret string) *Notifier {
	return &Notifier{
		webhookURL: webhookURL,
		hmacSecret: hmacSecret,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (n *Notifier) SendAlert(failedSources []string, errs []string, available, threshold, lastCnt int) {
	if n.webhookURL == "" {
		return
	}

	payload := map[string]interface{}{
		"event":              "pool_low_and_collect_failed",
		"time":               time.Now().Format(time.RFC3339),
		"available_pool":     available,
		"threshold":          threshold,
		"last_request_count": lastCnt,
		"failed_sources":     failedSources,
		"errors":             errs,
	}

	n.send(payload)
}

func (n *Notifier) SendError(event, detail string) {
	if n.webhookURL == "" {
		return
	}

	payload := map[string]string{
		"event":  event,
		"time":   time.Now().Format(time.RFC3339),
		"detail": detail,
	}

	n.send(payload)
}

func (n *Notifier) send(payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[notifier] marshal error: %v", err)
		return
	}

	req, err := http.NewRequest("POST", n.webhookURL, bytes.NewReader(data))
	if err != nil {
		log.Printf("[notifier] create request error: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	if n.hmacSecret != "" {
		mac := hmac.New(sha256.New, []byte(n.hmacSecret))
		mac.Write(data)
		sig := hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-Signature-256", fmt.Sprintf("sha256=%s", sig))
	}

	resp, err := n.client.Do(req)
	if err != nil {
		log.Printf("[notifier] failed to send webhook: %v", err)
		return
	}
	resp.Body.Close()
	if resp.StatusCode >= 300 {
		log.Printf("[notifier] webhook returned non-2xx: %d", resp.StatusCode)
	}
}