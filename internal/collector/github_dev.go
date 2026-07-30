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

func FetchGitHubDevelopers() ([]string, error) {
	lang := languages[rand.Intn(len(languages))]
	url := fmt.Sprintf("https://github.com/trending/developers?since=weekly&spoken_language=%s", lang)

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("github devs: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Accept", "text/html")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github devs: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("github devs: %w", err)
	}

	return extractGitHubDeveloperNames(string(body)), nil
}

func extractGitHubDeveloperNames(htmlContent string) []string {
	var names []string
	seen := make(map[string]bool)
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return names
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "h2" {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && c.Data == "a" {
					for _, attr := range c.Attr {
						if attr.Key == "href" && strings.HasPrefix(attr.Val, "/") {
							username := strings.TrimPrefix(attr.Val, "/")
							username = strings.TrimSpace(username)
							if username != "" && !seen[username] &&
								!strings.Contains(username, "/") {
								names = append(names, username)
								seen[username] = true
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

	if len(names) == 0 {
		var f2 func(*html.Node)
		f2 = func(n *html.Node) {
			if n.Type == html.ElementNode && n.Data == "a" {
				for _, attr := range n.Attr {
					if attr.Key == "href" && strings.HasPrefix(attr.Val, "/") {
						username := strings.TrimPrefix(attr.Val, "/")
						username = strings.TrimSpace(username)
						if username != "" && !seen[username] &&
							!strings.Contains(username, "/") &&
							!strings.Contains(username, "?") {
							names = append(names, username)
							seen[username] = true
						}
					}
				}
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f2(c)
			}
		}
		f2(doc)
	}

	return names
}