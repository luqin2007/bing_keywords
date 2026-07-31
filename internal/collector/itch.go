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
	"tag-1-bit",
	"tag-16-bit",
	"tag-1gam",
	"tag-2d",
	"tag-3d",
	"tag-3d-platformer",
	"tag-4x",
	"tag-7dfps",
	"tag-8-bit",
	"tag-abstract",
	"genre-action",
	"tag-action-rpg",
	"tag-action-adventure",
	"tag-adult",
	"genre-adventure",
	"tag-ai-generated",
	"tag-aliens",
	"tag-alternate-history",
	"tag-altgame",
	"tag-amare",
	"tag-amiga",
	"tag-amstrad-cpc",
	"tag-animals",
	"tag-anime",
	"tag-arcade",
	"tag-archery",
	"tag-arena-shooter",
	"tag-artgame",
	"tag-artificial-intelligence",
	"tag-ascii",
	"tag-aseprite",
	"tag-asteroids",
	"tag-astronomy",
	"tag-atari",
	"tag-atmospheric",
	"tag-augmented-reality",
	"tag-auto-battler",
	"tag-automation",
	"tag-backrooms",
	"tag-bara",
	"tag-barfcade",
	"tag-basketball",
	"tag-beat-em-up",
	"tag-beginner-friendly",
	"tag-binaural",
	"tag-bitsy",
	"tag-black-and-white",
	"tag-blocks",
	"tag-board-game",
	"tag-boring",
	"tag-boss-battle",
	"tag-boxing",
	"tag-boys-love",
	"tag-brain-training",
	"tag-breakout",
	"tag-bullet-hell",
	"tag-card-game",
	"tag-cartoon",
	"tag-casual",
	"tag-cats",
	"tag-celtic",
	"tag-character-customization",
	"tag-chess",
	"tag-chicken",
	"tag-chip8",
	"tag-chiptune",
	"tag-christmas",
	"tag-ciphers",
	"tag-city-builder",
	"tag-classes",
	"tag-clicker",
	"tag-co-op",
	"tag-colorful",
	"tag-combos",
	"tag-comedy",
	"tag-coming-of-age",
	"tag-commodore-64",
	"tag-construct-2",
	"tag-controller",
	"tag-cooking",
	"tag-corona-sdk",
	"tag-cozy",
	"tag-crafting",
	"tag-creative",
	"tag-creepy",
	"tag-creepypasta",
	"tag-crime",
	"tag-cult-classic",
	"tag-cute",
	"tag-cyberpunk",
	"tag-dark",
	"tag-dark-fantasy",
	"tag-dark-humor",
	"tag-dating-sim",
	"tag-deathmatch",
	"tag-deck-building",
	"tag-decker",
	"tag-demake",
	"tag-desktop-pet",
	"tag-destruction",
	"tag-detective",
	"tag-dice",
	"tag-difficult",
	"tag-dinosaurs",
	"tag-discord",
	"tag-dogs",
	"tag-doom",
	"tag-doom-wad",
	"tag-dos",
	"tag-dragons",
	"tag-drawing",
	"tag-dreamcast",
	"tag-dreams",
	"tag-dress-up",
	"tag-driving",
	"tag-drm-free",
	"tag-dungeon-crawler",
	"tag-dystopian",
	"tag-economy",
	"tag-educational",
	"tag-emulator",
	"tag-endless",
	"tag-endless-runner",
	"tag-episodic",
	"tag-eroge",
	"tag-erotic",
	"tag-escape-game",
	"tag-experimental",
	"tag-exploration",
	"tag-explosions",
	"tag-fairy-tale",
	"tag-family-friendly",
	"tag-fangame",
	"tag-fantasy",
	"tag-fantasy-console",
	"tag-farming",
	"tag-fast-paced",
	"tag-feel-good",
	"tag-female-protagonist",
	"tag-femdom",
	"tag-fighting",
	"tag-filipino",
	"tag-first-person",
	"tag-fishing",
	"tag-fnaf",
	"tag-flappy-bird",
	"tag-flat-shading",
	"tag-flight",
	"tag-flying",
	"tag-fmv",
	"tag-folklore",
	"tag-food",
	"tag-football",
	"tag-forest",
	"tag-fps",
	"tag-fps-platformer",
	"tag-frog",
	"tag-funny",
	"tag-furry",
	"tag-futuristic",
	"tag-gacha",
	"tag-gameboy",
	"tag-gameboy-advance",
	"tag-gameboy-rom",
	"tag-game-design",
	"tag-gmtkjam",
	"tag-gamemaker",
	"tag-gamepad",
	"tag-gardening",
	"tag-gay",
	"tag-gbjam",
	"tag-gender",
	"tag-generator",
	"tag-ggj15",
	"tag-ghosts",
	"tag-github",
	"tag-glitch",
	"tag-global-game-jam",
	"tag-ggj2020",
	"tag-godot",
	"tag-golf",
	"tag-gore",
	"tag-gorilla-tag",
	"tag-gothic",
	"tag-gravity",
	"tag-grayscale",
	"tag-hack-and-slash",
	"tag-hacking",
	"tag-halloween",
	"tag-hand-drawn",
	"tag-heist",
	"tag-hex-based",
	"tag-hidden-object",
	"tag-high-score",
	"tag-historical",
	"tag-homebrew",
	"tag-horrible",
	"tag-horror",
	"tag-horses",
	"tag-hypercard",
	"tag-idle",
	"tag-immersive",
	"tag-incredibox",
	"tag-incremental",
	"tag-indie",
	"tag-infinite-runner",
	"tag-instrument",
	"tag-interactive-fiction",
	"tag-internet",
	"tag-isekai",
	"tag-isometric",
	"tag-job-system",
	"tag-jrpg",
	"tag-jumping",
	"tag-kickstarter",
	"tag-kinect",
	"tag-kinetic-novel",
	"tag-kung-fu",
	"tag-leap-motion",
	"tag-lesbian",
	"tag-lgbt",
	"tag-lgbtqia",
	"tag-libgdx",
	"tag-life-simulation",
	"tag-liminal-space",
	"tag-live-action",
	"tag-lo-fi",
	"tag-local-co-op",
	"local-multiplayer",
	"tag-lone-survivor",
	"tag-loot",
	"tag-love2d",
	"tag-lovecraft",
	"tag-low-poly",
	"tag-lowrezjam",
	"tag-ludum-dare",
	"tag-ludum-dare-29",
	"tag-ludum-dare-30",
	"tag-ludum-dare-31",
	"tag-ludum-dare-32",
	"tag-ludum-dare-33",
	"tag-ludum-dare-34",
	"tag-ludum-dare-35",
	"tag-ludum-dare-36",
	"tag-ludum-dare-37",
	"tag-ludum-dare-38",
	"tag-ludum-dare-39",
	"tag-ludum-dare-40",
	"tag-ludum-dare-41",
	"tag-ludum-dare-42",
	"tag-ludum-dare-43",
	"tag-ludum-dare-44",
	"tag-ludum-dare-45",
	"tag-ludum-dare-46",
	"tag-ludum-dare-47",
	"tag-ludum-dare-48",
	"tag-ludum-dare-49",
	"tag-ludum-dare-50",
	"tag-ludum-dare-51",
	"tag-ludum-dare-52",
	"tag-ludum-dare-53",
	"tag-ludum-dare-54",
	"tag-ludum-dare-55",
	"tag-ludum-dare-56",
	"tag-ludum-dare-57",
	"tag-ludum-dare-58",
	"tag-ludum-dare-59",
	"tag-magic",
	"tag-magical-realism",
	"tag-male-protagonist",
	"tag-management",
	"tag-manga",
	"tag-mashup",
	"tag-massively-multiplayer",
	"tag-math",
	"tag-maze",
	"tag-meaningful-choices",
	"tag-mechs",
	"tag-medieval",
	"tag-mega-drive",
	"tag-meme",
	"tag-memoir",
	"tag-mental-health",
	"tag-metroidvania",
	"tag-midi",
	"tag-mind-bending",
	"tag-minecraft",
	"tag-minesweeper",
	"tag-minigames",
	"tag-minimalist",
	"tag-mmorpg",
	"tag-moddable",
	"tag-modeling",
	"tag-moe",
	"tag-monster-girls",
	"tag-monsters",
	"tag-mountains",
	"tag-mouse-only",
	"tag-movement-shooter",
	"tag-ms-dos",
	"tag-msx",
	"tag-multiplayer",
	"tag-multiple-endings",
	"tag-music",
	"tag-my-first-game-jam",
	"tag-mystery",
	"tag-mythology",
	"tag-narrative",
	"tag-nature",
	"tag-neon",
	"tag-nes",
	"tag-nes-rom",
	"tag-ninja",
	"tag-nintendo64",
	"tag-no-ai",
	"tag-noir",
	"tag-non-violent",
	"tag-non-eucledian",
	"tag-non-linear",
	"tag-nonogram",
	"tag-norse",
	"tag-norway",
	"tag-oculus-quest",
	"tag-oculus-rift",
	"tag-on-rails-shooter",
	"tag-one-button",
	"tag-one-hit-kill",
	"tag-open-source",
	"tag-open-world",
	"tag-otome",
	"tag-painting",
	"tag-parallax",
	"tag-parkour",
	"tag-parody",
	"tag-party-game",
	"tag-pastel",
	"tag-period-piece",
	"tag-perma-death",
	"tag-perspective",
	"tag-photorealistic",
	"tag-physics",
	"tag-pico-8",
	"tag-picross",
	"tag-pinball",
	"tag-pirates",
	"tag-pixel-art",
	"tag-pizza-tower",
	"genre-platformer",
	"tag-playdate",
	"tag-psp",
	"tag-point-and-click",
	"tag-post-apocalyptic",
	"tag-procedural",
	"tag-procjam",
	"tag-prototype",
	"tag-psx",
	"tag-psychedelic",
	"tag-psychological-horror",
	"genre-puzzle",
	"tag-puzzle-platformer",
	"tag-puzzlescript",
	"tag-pvp",
	"tag-queer",
	"tag-quiz",
	"tag-racing",
	"tag-railroad",
	"tag-real-time-strategy",
	"tag-real-time",
	"tag-relationship",
	"tag-relaxing",
	"tag-remake",
	"tag-renpy",
	"tag-retro",
	"tag-rhythm",
	"tag-roadtrip",
	"tag-robots",
	"tag-roguelike",
	"tag-roguelite",
	"genre-rpg",
	"tag-romance",
	"tag-rotation",
	"tag-rpgmaker",
	"tag-rpg-maker-mv",
	"tag-rpg-maker-mz",
	"tag-runner",
	"tag-sailing",
	"tag-sandbox",
	"tag-satire",
	"tag-scary",
	"tag-science-fiction",
	"tag-score-attack",
	"tag-scratch",
	"tag-screensaver",
	"tag-secrets",
	"tag-sega-genesis",
	"tag-7drl",
	"tag-sfml",
	"tag-shadows",
	"tag-sharecart1000",
	"tag-shoot-em-up",
	"genre-shooter",
	"tag-short",
	"tag-side-scroller",
	"tag-simple",
	"genre-simulation",
	"tag-singleplayer",
	"tag-siren-head",
	"tag-sitting-simulator",
	"tag-skating",
	"tag-skeletons",
	"tag-slasher",
	"tag-slice-of-life",
	"tag-slime",
	"tag-snake",
	"tag-soccer",
	"tag-sokoban",
	"tag-solitaire",
	"tag-souls-like",
	"tag-soundtoy",
	"tag-sourcecode",
	"tag-space",
	"tag-space-sim",
	"tag-speedrun",
	"tag-split-screen",
	"tag-spooky",
	"tag-spoopy",
	"genre-sports",
	"tag-sprunki",
	"tag-stealth",
	"tag-steampunk",
	"tag-stencyl",
	"tag-stop-motion",
	"tag-story-rich",
	"genre-strategy",
	"tag-strategy-rpg",
	"tag-streaming",
	"tag-suika-game",
	"tag-superhero",
	"tag-supernatural",
	"tag-superpowers",
	"tag-surreal",
	"tag-survival",
	"tag-survival-horror",
	"tag-survivor-like",
	"tag-suspense",
	"tag-swords",
	"tag-synthwave",
	"tag-tablet",
	"tag-tactical",
	"tag-tactical-rpg",
	"tag-tanks",
	"tag-tarot",
	"tag-team-based",
	"tag-tennis",
	"tag-tentacles",
	"tag-tetris",
	"tag-text-based",
	"tag-thanksgiving",
	"tag-third-person",
	"tag-third-person-shooter",
	"tag-thriller",
	"tag-tic-80",
	"tag-time-attack",
	"tag-time-travel",
	"tag-top-down-adventure",
	"tag-top-down-shooter",
	"tag-top-down",
	"tag-touch-friendly",
	"tag-touhou",
	"tag-tower-defense",
	"tag-trading",
	"tag-trains",
	"tag-transgender",
	"tag-traps",
	"tag-trashcore",
	"tag-trijam",
	"tag-trivia",
	"tag-turbografx",
	"tag-turn-based",
	"tag-turn-based-combat",
	"tag-turn-based-strategy",
	"tag-tutorial",
	"tag-twin-stick-shooter",
	"tag-twine",
	"tag-two-colors",
	"tag-tycoon",
	"tag-typing",
	"tag-tyranobuilder",
	"tag-undertale",
	"tag-underwater",
	"tag-unicorns",
	"tag-unity",
	"tag-unreal-engine",
	"tag-upgrades",
	"tag-urban",
	"tag-vampire",
	"tag-vector",
	"tag-versus",
	"tag-vic-20",
	"tag-victorian",
	"tag-video",
	"tag-violent",
	"tag-virtual-pet",
	"tag-virtual-reality",
	"genre-visual-novel",
	"tag-visualization",
	"tag-voice-acting",
	"tag-voice-controlled",
	"tag-voxel",
	"tag-walking-simulator",
	"tag-war",
	"tag-watercolor",
	"tag-weird",
	"tag-western",
	"tag-wholesome",
	"tag-wild-west",
	"tag-wizards",
	"tag-wobbly",
	"tag-word-game",
	"tag-wordle",
	"tag-working-simulator",
	"tag-world-war-i",
	"tag-world-war-ii",
	"tag-yandere",
	"tag-yaoi",
	"tag-yuri",
	"tag-zero-gravity",
	"tag-zine",
	"tag-zombies",
	"tag-zx-spectrum",
}

func FetchItchGames() ([]string, string, error) {
	tag := itchTags[rand.Intn(len(itchTags))]
	url := fmt.Sprintf("https://itch.io/games/%s", tag)
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
