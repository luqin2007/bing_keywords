package collector

import (
	"math/rand"
	"time"

	"everyday_keywords/internal/db"
)

type SourceResult struct {
	Words []string
	Err   error
}

type SourceFunc func() ([]string, error)

type Collector struct {
	db      *db.DB
	sources []SourceFunc
}

func New(database *db.DB) *Collector {
	return &Collector{
		db: database,
		sources: []SourceFunc{
			FetchGitHubRepos,
			FetchGitHubDevelopers,
			FetchSourceForge,
			FetchOpenRouterModels,
		},
	}
}

func (c *Collector) Collect(target int) ([]string, []error) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	indices := rng.Perm(len(c.sources))

	var allErrors []error
	var allWords []string

	for _, idx := range indices {
		words, err := c.sources[idx]()
		if err != nil {
			allErrors = append(allErrors, err)
			continue
		}
		if len(words) == 0 {
			continue
		}
		inserted, err := c.db.Insert(words)
		if err != nil {
			allErrors = append(allErrors, err)
			continue
		}
		if inserted > 0 {
			allWords = append(allWords, words...)
		}
		count, _ := c.db.Count()
		if count >= target {
			break
		}
	}

	return allWords, allErrors
}

func (c *Collector) CollectUntil(target int) ([]string, []error) {
	var allErrors []error
	var allCollected []string

	for attempts := 0; attempts < 8; attempts++ {
		words, errs := c.Collect(target)
		allCollected = append(allCollected, words...)
		allErrors = append(allErrors, errs...)

		count, _ := c.db.Count()
		if count >= target {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	return allCollected, allErrors
}