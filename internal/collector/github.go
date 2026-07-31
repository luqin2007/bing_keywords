package collector

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

var languages = []string{
	"python", "javascript", "typescript", "go", "rust", "c", "c++", "java",
	"kotlin", "swift", "ruby", "php", "scala", "haskell", "elixir", "lua",
	"dart", "r", "julia", "perl", "lisp", "clojure", "erlang", "nim",
	"zig", "crystal", "ocaml", "solidity", "fortran", "COBOL",
}

func FetchGitHubRepos() ([]string, string, error) {
	lang := languages[rand.Intn(len(languages))]
	url := fmt.Sprintf("https://github.com/trending?since=weekly&spoken_language=%s", lang)
	source := fmt.Sprintf("github_repo(language:%s)", lang)

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, source, fmt.Errorf("github repos: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Accept", "text/html")

	resp, err := client.Do(req)
	if err != nil {
		return nil, source, fmt.Errorf("github repos: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, source, fmt.Errorf("github repos: %w", err)
	}

	return extractGitHubRepoNames(string(body)), source, nil
}

func extractGitHubRepoNames(htmlContent string) []string {
	var names []string
	seen := make(map[string]bool)
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return names
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					href := attr.Val
					if i := strings.IndexByte(href, '?'); i >= 0 {
						href = href[:i]
					}
					if strings.Count(href, "/") == 2 && strings.HasPrefix(href, "/") {
						parts := strings.Split(strings.TrimPrefix(href, "/"), "/")
						if len(parts) == 2 {
							repoName := strings.TrimSpace(parts[1])
							if repoName != "" && !seen[repoName] {
								names = append(names, repoName)
								seen[repoName] = true
							}
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return names
}