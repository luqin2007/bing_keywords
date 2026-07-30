package collector

import (
	"math/rand"
	"time"

	"everyday_keywords/internal/db"
)

type SourceResult struct {
	Source string
	Words  []string
	Err    error
}

type SourceFunc func() (words []string, source string, err error)

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

func (c *Collector) Collect(target int) ([]SourceResult, []error) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	indices := rng.Perm(len(c.sources))

	var allErrors []error
	var allResults []SourceResult

	for _, idx := range indices {
		words, source, err := c.sources[idx]()
		if err != nil {
			allErrors = append(allErrors, err)
			continue
		}
		if len(words) == 0 {
			continue
		}
		inserted, err := c.db.Insert(words, source)
		if err != nil {
			allErrors = append(allErrors, err)
			continue
		}
		if inserted > 0 {
			allResults = append(allResults, SourceResult{Source: source, Words: words})
		}
		count, _ := c.db.Count()
		if count >= target {
			break
		}
	}

	return allResults, allErrors
}

func (c *Collector) CollectUntil(target int) ([]SourceResult, []error) {
	var allErrors []error
	var allResults []SourceResult

	for attempts := 0; attempts < 8; attempts++ {
		results, errs := c.Collect(target)
		allResults = append(allResults, results...)
		allErrors = append(allErrors, errs...)

		count, _ := c.db.Count()
		if count >= target {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	c.db.UpdateLastCollectTime()
	return allResults, allErrors
}