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
			FetchSteamGames,
			FetchItchGames,
		},
	}
}

func (c *Collector) Collect(targetNew int) ([]SourceResult, []error, int) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	indices := rng.Perm(len(c.sources))

	var allErrors []error
	var allResults []SourceResult
	newWords := 0

	for _, idx := range indices {
		if newWords >= targetNew {
			break
		}

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
			newWords += inserted
		}
	}

	return allResults, allErrors, newWords
}

func (c *Collector) CollectUntil(targetNew int) ([]SourceResult, []error) {
	var allErrors []error
	var allResults []SourceResult
	totalNew := 0

	for attempts := 0; attempts < 8; attempts++ {
		results, errs, newWords := c.Collect(targetNew - totalNew)
		allResults = append(allResults, results...)
		allErrors = append(allErrors, errs...)
		totalNew += newWords

		if totalNew >= targetNew {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	c.db.UpdateLastCollectTime()
	return allResults, allErrors
}