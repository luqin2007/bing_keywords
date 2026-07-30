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

var sourceForgeOS = []string{"windows", "linux", "mac", "android", "ios", "bsd"}

var sourceForgeCategories = []string{
	"development", "communications", "games", "multimedia", "security",
	"system-administration", "education", "business", "internet",
	"science-engineering", "mobile", "office", "social-networking",
	"storage", "text-editors", "voip", "web-browsers", "clustering",
	"database", "networking", "printing", "religion", "screensavers",
}

func FetchSourceForge() ([]string, error) {
	os := sourceForgeOS[rand.Intn(len(sourceForgeOS))]
	cat := sourceForgeCategories[rand.Intn(len(sourceForgeCategories))]
	url := fmt.Sprintf("https://sourceforge.net/directory/os:%s/category:%s/", os, cat)

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("sourceforge: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Accept", "text/html")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sourceforge: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1 << 20))
	if err != nil {
		return nil, fmt.Errorf("sourceforge: %w", err)
	}

	return extractSourceForgeAppNames(string(body)), nil
}

func extractSourceForgeAppNames(htmlContent string) []string {
	var names []string
	seen := make(map[string]bool)
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return names
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			var isProjectLink bool
			var href string
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					href = attr.Val
				}
			}
			if strings.HasPrefix(href, "/projects/") && strings.Count(href, "/") == 2 {
				isProjectLink = true
			}
			if isProjectLink {
				title := extractText(n)
				title = strings.TrimSpace(title)
				if title != "" && !seen[title] {
					names = append(names, title)
					seen[title] = true
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
				var href string
				for _, attr := range n.Attr {
					if attr.Key == "href" {
						href = attr.Val
					}
				}
				if strings.Contains(href, "/projects/") {
					title := extractText(n)
					title = strings.TrimSpace(title)
					if title != "" && !seen[title] {
						names = append(names, title)
						seen[title] = true
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

func extractText(n *html.Node) string {
	var text string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			text += n.Data
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return text
}