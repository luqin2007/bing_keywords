package cleaner

import (
	"log"
	"time"

	"everyday_keywords/internal/db"
)

type Cleaner struct {
	db     *db.DB
	stopCh chan struct{}
}

func New(database *db.DB) *Cleaner {
	return &Cleaner{db: database, stopCh: make(chan struct{})}
}

func (c *Cleaner) Start(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.clean()
			case <-c.stopCh:
				return
			}
		}
	}()
}

func (c *Cleaner) Stop() {
	close(c.stopCh)
}

func (c *Cleaner) clean() {
	deleted, err := c.db.DeleteOlderThan(30)
	if err != nil {
		log.Printf("[cleaner] delete error: %v", err)
		return
	}
	if deleted > 0 {
		total, _ := c.db.Count()
		log.Printf("[cleaner] deleted %d keywords older than 30 days, remaining: %d", deleted, total)
	}
}