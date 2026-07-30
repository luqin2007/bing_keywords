package notifier

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type Notifier struct {
	webhookURL string
	client     *http.Client
}

func New(webhookURL string) *Notifier {
	return &Notifier{
		webhookURL: webhookURL,
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

	data, _ := json.Marshal(payload)
	resp, err := n.client.Post(n.webhookURL, "application/json", bytes.NewReader(data))
	if err != nil {
		log.Printf("[notifier] failed to send webhook: %v", err)
		return
	}
	resp.Body.Close()
	if resp.StatusCode >= 300 {
		log.Printf("[notifier] webhook returned non-2xx: %d", resp.StatusCode)
	}
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

	data, _ := json.Marshal(payload)
	resp, err := n.client.Post(n.webhookURL, "application/json", bytes.NewReader(data))
	if err != nil {
		log.Printf("[notifier] failed to send webhook: %v", err)
		return
	}
	resp.Body.Close()
	if resp.StatusCode >= 300 {
		log.Printf("[notifier] webhook returned non-2xx: %d", resp.StatusCode)
	}
}