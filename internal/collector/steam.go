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
	492:     "Indie",
	19:      "Action",
	21:      "Adventure",
	597:     "Casual",
	4182:    "Singleplayer",
	599:     "Simulation",
	122:     "RPG",
	9:       "Strategy",
	3871:    "2D",
	493:     "Early Access",
	4191:    "3D",
	113:     "Free to Play",
	4166:    "Atmospheric",
	1742:    "Story Rich",
	4305:    "Colorful",
	3834:    "Exploration",
	1684:    "Fantasy",
	3859:    "Multiplayer",
	4726:    "Cute",
	3964:    "Pixel Graphics",
	3993:    "Combat",
	3839:    "First-Person",
	1664:    "Puzzle",
	1654:    "Relaxing",
	4106:    "Action-Adventure",
	4252:    "Stylized",
	4136:    "Funny",
	1773:    "Arcade",
	7481:    "Controller",
	4085:    "Anime",
	1667:    "Horror",
	6730:    "PvE",
	3942:    "Sci-fi",
	1685:    "Co-op",
	701:     "Sports",
	128:     "Massively Multiplayer",
	1697:    "Third Person",
	6426:    "Choices Matter",
	4667:    "Violent",
	4004:    "Retro",
	4791:    "Top-Down",
	5350:    "Family Friendly",
	1774:    "Shooter",
	7208:    "Female Protagonist",
	12095:   "Sexual Content",
	4342:    "Dark",
	1775:    "PvP",
	4175:    "Realistic",
	699:     "Racing",
	5716:    "Mystery",
	7250:    "Linear",
	6971:    "Multiple Endings",
	1695:    "Open World",
	6650:    "Nudity",
	3843:    "Online Co-Op",
	1662:    "Survival",
	4747:    "Character Customization",
	1719:    "Comedy",
	4195:    "Cartoony",
	3799:    "Visual Novel",
	3968:    "Physics",
	1721:    "Psychological Horror",
	4345:    "Gore",
	1625:    "Platformer",
	4057:    "Magic",
	1716:    "Roguelike",
	5379:    "2D Platformer",
	3959:    "Roguelite",
	3810:    "Sandbox",
	12472:   "Management",
	8945:    "Resource Management",
	3916:    "Old School",
	1708:    "Tactical",
	1663:    "FPS",
	4172:    "Medieval",
	6815:    "Hand-drawn",
	4231:    "Action RPG",
	9204:    "Immersive Sim",
	4325:    "Turn-Based Combat",
	4094:    "Minimalist",
	1741:    "Turn-Based Strategy",
	1702:    "Crafting",
	4295:    "Futuristic",
	1643:    "Building",
	1698:    "Point &amp; Click",
	5608:    "Emotional",
	4604:    "Dark Fantasy",
	42804:   "Action Roguelike",
	4562:    "Cartoon",
	5125:    "Procedural Generation",
	1755:    "Space",
	4026:    "Difficult",
	5395:    "3D Platformer",
	11014:   "Interactive Fiction",
	4947:    "Romance",
	4486:    "Choose Your Own Adventure",
	30358:   "Nature",
	6129:    "Logic",
	3978:    "Survival Horror",
	14139:   "Turn-Based Tactics",
	87:      "Utilities",
	7368:    "Local Multiplayer",
	9130:    "Hentai",
	6691:    "1990's",
	7332:    "Base Building",
	1646:    "Hack and Slash",
	1710:    "Surreal",
	560542:  "Incremental",
	9551:    "Dating Sim",
	1738:    "Hidden Object",
	21978:   "VR",
	4885:    "Bullet Hell",
	3798:    "Side Scroller",
	17305:   "Strategy RPG",
	3835:    "Post-apocalyptic",
	5537:    "Puzzle Platformer",
	1036:    "Education",
	5900:    "Walking Simulator",
	1720:    "Dungeon Crawler",
	3854:    "Lore-Rich",
	4145:    "Cinematic",
	17389:   "Tabletop",
	10235:   "Life Sim",
	5154:    "Score Attack",
	42152:   "Dialogue Heavy",
	6276:    "Inventory Management",
	1666:    "Card Game",
	1678:    "War",
	4255:    "Shoot 'Em Up",
	31275:   "Text-Based",
	4695:    "Economy",
	5186:    "Psychological",
	4434:    "JRPG",
	1687:    "Stealth",
	1756:    "Great Soundtrack",
	1659:    "Zombies",
	615955:  "Idler",
	7743:    "1980s",
	84:      "Design &amp; Illustration",
	8369:    "Investigation",
	3841:    "Local Co-Op",
	4975:    "2.5D",
	3987:    "Historical",
	4064:    "Thriller",
	10695:   "Party-Based RPG",
	10808:   "Supernatural",
	12057:   "Tutorial",
	5851:    "Isometric",
	5923:    "Dark Humor",
	32322:   "Deckbuilding",
	6869:    "Nonlinear",
	16689:   "Time Management",
	1677:    "Turn-Based",
	4168:    "Military",
	7926:    "Artificial Intelligence",
	4637:    "Top-Down Shooter",
	4711:    "Replay Value",
	3814:    "Third-Person Shooter",
	9541:    "Demons",
	5711:    "Team-Based",
	4236:    "Loot",
	1673:    "Aliens",
	21725:   "Tactical RPG",
	4115:    "Cyberpunk",
	5652:    "Collectathon",
	5613:    "Detective",
	5752:    "Robots",
	5673:    "Modern",
	8013:    "Software",
	5030:    "Dystopian",
	3813:    "Real Time Tactics",
	4400:    "Abstract",
	1759:    "Perma Death",
	1645:    "Tower Defense",
	1644:    "Driving",
	4474:    "CRPG",
	1770:    "Board Game",
	1676:    "RTS",
	3877:    "Precision Platformer",
	5547:    "Arena Shooter",
	791774:  "Card Battler",
	29482:   "Souls-like",
	1751:    "Comic Book",
	97376:   "Cozy",
	1714:    "Psychedelic",
	4328:    "City Builder",
	255534:  "Automation",
	4508:    "Co-op Campaign",
	10397:   "Memes",
	17894:   "Cats",
	4845:    "Capitalism",
	16094:   "Mythology",
	4684:    "Wargame",
	4598:    "Alternate History",
	4046:    "Dragons",
	4840:    "4 Player Local",
	13906:   "Game Development",
	916648:  "Creature Collector",
	7569:    "Grid-Based Movement",
	6378:    "Crime",
	4234:    "Short",
	8122:    "Level Editor",
	5363:    "Destruction",
	4155:    "Class-Based",
	4036:    "Parkour",
	1734:    "Fast-Paced",
	15045:   "Flight",
	4158:    "Beat 'em up",
	1669:    "Moddable",
	1628:    "Metroidvania",
	872:     "Animation &amp; Modeling",
	15277:   "Philosophical",
	8666:    "Runner",
	1621:    "Music",
	19995:   "Dark Comedy",
	4202:    "Trading",
	4736:    "2D Fighter",
	87918:   "Farming Sim",
	3920:    "Cooking",
	1100687: "Automobile Sim",
	7948:    "Soundtrack",
	5765:    "Gun Customization",
	7178:    "Party Game",
	6506:    "3D Fighter",
	1084988: "Auto Battler",
	3878:    "Competitive",
	1752:    "Rhythm",
	1743:    "Fighting",
	11104:   "Vehicular Combat",
	5055:    "eSports",
	1754:    "MMORPG",
	6052:    "Noir",
	5372:    "Conspiracy",
	4559:    "Quick-Time Events",
	7432:    "Lovecraftian",
	4608:    "Swordplay",
	5794:    "Science",
	220585:  "Colony Sim",
	24003:   "Word Game",
	4758:    "Twin Stick Shooter",
	1651:    "Satire",
	16598:   "Space Sim",
	3952:    "Gothic",
	4878:    "Parody",
	4364:    "Grand Strategy",
	5981:    "Mining",
	9592:    "Dynamic Narration",
	552282:  "Wholesome",
	13782:   "Experimental",
	784:     "Video Production",
	353880:  "Looter Shooter",
	21006:   "Underground",
	1693:    "Classic",
	176981:  "Battle Royale",
	7702:    "Narrative",
	198631:  "Mystery Dungeon",
	1027:    "Audio Production",
	22602:   "Agriculture",
	4835:    "6DOF",
	5796:    "Bullet Time",
	10816:   "Split Screen",
	4150:    "World War II",
	15564:   "Fishing",
	6625:    "Time Manipulation",
	1091588: "Roguelike Deckbuilder",
	6915:    "Martial Arts",
	4777:    "Spectacle fighter",
	16250:   "Gambling",
	5411:    "Beautiful",
	4821:    "Mechs",
	4102:    "Combat Racing",
	1665:    "Match 3",
	620519:  "Hero Shooter",
	1637:    "Dogs",
	17770:   "Asynchronous Multiplayer",
	10383:   "Transportation",
	3934:    "Immersive",
	91114:   "Shop Keeper",
	18594:   "FMV",
	1723:    "Action RTS",
	1732:    "Voxel",
	1688:    "Ninja",
	1100689: "Open World Survival Craft",
	10679:   "Time Travel",
	12686:   "Vampires",
	9271:    "Trading Card Game",
	5300:    "God Game",
	13070:   "Solitaire",
	31579:   "Otome",
	1777:    "Steampunk",
	1681:    "Pirates",
	9157:    "Underwater",
	1023537: "Boomer Shooter",
	1717:    "Hex Grid",
	1445:    "Software Training",
	9564:    "Hunting",
	5502:    "Hacking",
	26921:   "Political Sim",
	1616:    "Trains",
	180368:  "Faith",
	13276:   "Tanks",
	1674:    "Typing",
	1670:    "4X",
	1718:    "MOBA",
	1730:    "Sokoban",
	5432:    "Programming",
	97070:   "Assassins",
	5708:    "Remake",
	1671:    "Superhero",
	7108:    "Party",
	6310:    "Diplomacy",
	9626:    "Animals",
	5160:    "Dinosaurs",
	3955:    "Character Action Game",
	1647:    "Western",
	723991:  "Bullet Heaven",
	8093:    "Minigames",
	809:     "Photo Editing",
	1680:    "Heist",
	11123:   "Mouse Only",
	5179:    "Cold War",
	454187:  "Traditional Roguelike",
	35079:   "Job Simulator",
	6910:    "Naval",
	9803:    "Snow",
	4137:    "Transhumanism",
	4994:    "Naval Combat",
	13577:   "Sailing",
	769306:  "Escape Room",
	13382:   "Archery",
	4190:    "Addictive",
	6041:    "Horses",
	4161:    "Real-Time",
	14720:   "Nostalgia",
	4520:    "Farming",
	4242:    "Episodic",
	8253:    "Music-Based Procedural Generation",
	1254546: "Football (Soccer)",
	17015:   "Werewolves",
	3965:    "Epic",
	10437:   "Trivia",
	11333:   "Villain Protagonist",
	7622:    "Offroad",
	5390:    "Time Attack",
	7423:    "Sniper",
	7107:    "Real-Time with Pause",
	56690:   "On-Rails Shooter",
	5230:    "Sequel",
	71389:   "Spelling",
	6702:    "Mars",
	1100686: "Outbreak Sim",
	5382:    "World War I",
	4535:    "Dwarves",
	12190:   "Boxing",
	4184:    "Chess",
	1320952: "Desktop Companion",
	4291:    "Spaceships",
	25085:   "Touch-Friendly",
	5348:    "Mod",
	1199779: "Extraction Shooter",
	1100688: "Medical Sim",
	1746:    "Basketball",
	7038:    "Golf",
	19780:   "Submarine",
	745697:  "Social Deduction",
	198913:  "Motorbike",
	42089:   "Jump Scare",
	5727:    "Baseball",
	150626:  "Gaming",
	7556:    "Dice",
	6948:    "Rome",
	776177:  "360 Video",
	6621:    "Pinball",
	123332:  "Bikes",
	61357:   "Electronic Music",
	6054:    "Elves",
	11095:   "Boss Rush",
	856791:  "Asymmetric VR",
	889937:  "Decorating",
	1239876: "Organizing",
	47827:   "Wrestling",
	1753:    "Skateboarding",
	15954:   "Silent Protagonist",
	189941:  "Instrumental Music",
	1254552: "Football (American)",
	22955:   "Mini Golf",
	4852:    "Billiards",
	11634:   "Vikings",
	23491:   "Cleaning",
	96359:   "Skating",
	25959:   "Wuxia",
	337964:  "Rock Music",
	760247:  "Xianxia",
	19568:   "Cycling",
	52406:   "Cult",
	6214:    "Birds",
	5914:    "Tennis",
	8075:    "TrackIR",
	1776:    "Espionage",
	15868:   "Motocross",
	14906:   "Intentionally Awkward Controls",
	33572:   "Mahjong",
	10617:   "Samurai",
	324176:  "Hockey",
	507423:  "Foxes",
	7328:    "Bowling",
	6835:    "Poker",
	3796:    "Based On A Novel",
	27758:   "Voice Control",
	129761:  "ATV",
	117648:  "8-bit Music",
	17337:   "Lemmings",
	28444:   "Snowboarding",
	7309:    "Skiing",
	603297:  "Hardware",
	37376:   "Falling Blocks",
	252854:  "BMX",
	323922:  "Musou",
	21635:   "Language Learning",
	1220528: "Hobby Sim",
	5407:    "Benchmark",
	1352486: "Capybaras",
	20486:   "Wolves",
	46348:   "Zoo",
	847164:  "Volleyball",
	158638:  "Cricket",
	49213:   "Rugby",
	363767:  "Snooker",
	5941:    "Reboot",
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
	url := fmt.Sprintf("https://store.steampowered.com/search/?hwtype=0&tags=%d&supportedlang=english&ndl=1", catID)
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
