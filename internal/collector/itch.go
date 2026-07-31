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

var itchTags = []string{
	"puzzle", "platformer", "horror", "adventure", "simulation", "strategy",
	"rpg", "action", "shooter", "visual-novel", "point-and-click", "racing",
	"sports", "fighting", "rhythm", "stealth", "survival", "battle-royale",
	"tower-defense", "card-game", "board-game", "trivia", "educational",
	"retro", "pixel-art", "low-poly", "3d", "2d", "first-person",
	"third-person", "top-down", "side-scroller", "open-world", "sandbox",
	"procedural-generation", "roguelike", "metroidvania", "souls-like",
	"bullet-hell", "beat-em-up", "hack-and-slash", "turn-based",
	"real-time", "multiplayer", "co-op", "singleplayer", "local-multiplayer",
	"online-multiplayer", "split-screen", "moddable", "controller",
	"keyboard-only", "mouse-only", "touch-friendly", "vr", "ar",
}

func FetchItchGames() ([]string, string, error) {
	tag := itchTags[rand.Intn(len(itchTags))]
	url := fmt.Sprintf("https://itch.io/games/tag-%s/sort-popular", tag)
	source := fmt.Sprintf("itch(tag:%s)", tag)

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, source, fmt.Errorf("itch: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "text/html")

	resp, err := client.Do(req)
	if err != nil {
		return nil, source, fmt.Errorf("itch: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, source, fmt.Errorf("itch: %w", err)
	}

	return extractItchGameNames(string(body)), source, nil
}

func extractItchGameNames(htmlContent string) []string {
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
				if attr.Key == "class" && strings.Contains(attr.Val, "game_link") {
					title := extractText(n)
					title = strings.TrimSpace(title)
					if title != "" && !seen[title] && len(title) <= 60 {
						names = append(names, title)
						seen[title] = true
					}
					return
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
			if n.Type == html.ElementNode && n.Data == "div" {
				for _, attr := range n.Attr {
					if attr.Key == "class" && strings.Contains(attr.Val, "game_cell") {
						for c := n.FirstChild; c != nil; c = c.NextSibling {
							if c.Type == html.ElementNode && c.Data == "a" {
								for _, ca := range c.Attr {
									if ca.Key == "class" && strings.Contains(ca.Val, "game_link") {
										title := extractText(c)
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
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f2(c)
			}
		}
		f2(doc)
	}

	if len(names) == 0 {
		var f3 func(*html.Node)
		f3 = func(n *html.Node) {
			if n.Type == html.ElementNode && n.Data == "a" {
				hasGameLink := false
				for _, attr := range n.Attr {
					if attr.Key == "class" && strings.Contains(attr.Val, "game_link") {
						hasGameLink = true
						break
					}
				}
				if !hasGameLink {
					for _, attr := range n.Attr {
						if attr.Key == "href" && strings.Contains(attr.Val, "/games/") {
							hasGameLink = true
							break
						}
					}
				}
				if hasGameLink {
					title := extractText(n)
					title = strings.TrimSpace(title)
					if title != "" && !seen[title] && len(title) <= 60 {
						names = append(names, title)
						seen[title] = true
					}
				}
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f3(c)
			}
		}
		f3(doc)
	}

	return names
}