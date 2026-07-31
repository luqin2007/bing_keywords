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

var steamCategories = map[int]string{
	1:  "action",
	2:  "adventure",
	3:  "casual",
	4:  "early_access",
	5:  "free_to_play",
	6:  "indie",
	7:  "massively_multiplayer",
	8:  "racing",
	9:  "rpg",
	10: "simulation",
	11: "sports",
	12: "strategy",
}

var steamCategoryIDs []int

func init() {
	for id := range steamCategories {
		steamCategoryIDs = append(steamCategoryIDs, id)
	}
}

func FetchSteamGames() ([]string, string, error) {
	catID := steamCategoryIDs[rand.Intn(len(steamCategoryIDs))]
	catName := steamCategories[catID]
	url := fmt.Sprintf("https://store.steampowered.com/search/?category1=%d&count=50&l=english", catID)
	source := fmt.Sprintf("steam(category:%s)", catName)

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, source, fmt.Errorf("steam: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "text/html")

	resp, err := client.Do(req)
	if err != nil {
		return nil, source, fmt.Errorf("steam: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, source, fmt.Errorf("steam: %w", err)
	}

	return extractSteamGameNames(string(body)), source, nil
}

func extractSteamGameNames(htmlContent string) []string {
	var names []string
	seen := make(map[string]bool)
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return names
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "span" {
			isTitle := false
			for _, attr := range n.Attr {
				if attr.Key == "class" && strings.Contains(attr.Val, "title") {
					isTitle = true
					break
				}
			}
			if isTitle {
				title := extractText(n)
				title = strings.TrimSpace(title)
				if title != "" && !seen[title] && len(title) <= 60 {
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
				for _, attr := range n.Attr {
					if attr.Key == "class" && strings.Contains(attr.Val, "search_result_row") {
						for c := n.FirstChild; c != nil; c = c.NextSibling {
							if c.Type == html.ElementNode && c.Data == "div" {
								for _, ca := range c.Attr {
									if ca.Key == "class" && strings.Contains(ca.Val, "responsive_search_name_combined") {
										for c2 := c.FirstChild; c2 != nil; c2 = c2.NextSibling {
											if c2.Type == html.ElementNode && c2.Data == "span" {
												for _, ca2 := range c2.Attr {
													if ca2.Key == "class" && strings.Contains(ca2.Val, "title") {
														title := extractText(c2)
														title = strings.TrimSpace(title)
														if title != "" && !seen[title] && len(title) <= 60 {
															names = append(names, title)
															seen[title] = true
														}
													}
												}
											}
										}
									}
								}
							}
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