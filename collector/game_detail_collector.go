package collector

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/guumaster/netcon-metadata/game"
	"github.com/puzpuzpuz/xsync/v3"
	"log"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func MakeGameDetailCollector(detailList *xsync.MapOf[string, game.Detail]) (*colly.Collector, error) {
	c := colly.NewCollector(
		// MaxDepth is 2, so only the links on the scraped page
		// and links on those pages are visited
		colly.AllowedDomains("netconplay.com", "app.netconplay.com", "www.netconplay.com", "*.netconplay.com"),
		//colly.MaxDepth(5),
		colly.Async(true),
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"),
	)
	_ = c.Limit(&colly.LimitRule{Parallelism: 10})

	// Find and visit all links
	c.OnHTML("div.panel-body", func(e *colly.HTMLElement) {
		u := e.Request.URL
		gameId := path.Base(u.Path)
		if gameId == "games" {
			return
		}

		link := e.Request.URL.String()
		title := e.DOM.Find("h3").First().Text()

		content := e.DOM.Find("div").First()

		masterName := findWithoutTitle("p:nth-child(1)", content)
		description := content.Find("p:nth-child(2)").First().Text()
		security := findWithoutTitle("p:nth-child(3)", content)
		sensibleContent := findWithoutTitle("p:nth-child(4)", content)
		system := findWithoutTitle("p:nth-child(5)", content)
		platform := findWithoutTitle("p:nth-child(6)", content)
		startDateStr := findWithoutTitle("p:nth-child(7)", content)
		startDate, _ := parseDate(startDateStr)
		durationStr := findWithoutTitle("p:nth-child(8)", content)
		durationHours, _ := sanitizeNumber(durationStr)
		streamed := findWithoutTitle("p:nth-child(9)", content)
		initiation := findWithoutTitle("p:nth-child(10)", content)

		masterDesc := ""
		channel := ""
		maxPlayers := 0
		regPlayers := 0

		for i := 11; i <= 15; i++ {
			header := strings.ToLower(content.Find(fmt.Sprintf("p:nth-child(%d) strong", i)).Text())

			switch header {
			case "sobre la directora":
				masterDesc = content.Find(fmt.Sprintf("p:nth-child(%d)", i+1)).Text()
			case "canal de emision":
				channel = findWithoutTitle(fmt.Sprintf("p:nth-child(%d)", i), content)
			case "número máximo de jugadoras":
				maxPlayersStr := findWithoutTitle(fmt.Sprintf("p:nth-child(%d)", i), content)
				maxPlayers, _ = sanitizeNumber(maxPlayersStr)
			case "número de jugadoras registradas":
				regPlayersStr := findWithoutTitle(fmt.Sprintf("p:nth-child(%d)", i), content)
				regPlayers, _ = sanitizeNumber(regPlayersStr)
			}

		}

		duration := time.Duration(durationHours) * time.Hour

		gameDetail := game.Detail{
			Id:                strings.TrimSpace(gameId),
			Link:              strings.TrimSpace(link),
			BackgroundImage:   "",
			Title:             strings.TrimSpace(title),
			System:            strings.TrimSpace(system),
			Description:       strings.TrimSpace(description),
			MasterName:        strings.TrimSpace(masterName),
			MasterDescription: strings.TrimSpace(masterDesc),
			StartDate:         startDate,
			Duration:          strconv.Itoa(durationHours),
			EndDate:           startDate.Add(duration),
			Security:          strings.TrimSpace(security),
			SensibleContent:   strings.TrimSpace(sensibleContent),
			Platform:          strings.TrimSpace(platform),
			Streamed:          strings.TrimSpace(streamed) == "Si",
			Channel:           strings.TrimSpace(channel),
			InitiationGame:    strings.TrimSpace(initiation) == "Si",
			MaxPlayers:        maxPlayers,
			RegisteredPlayers: regPlayers,
			Completed:         maxPlayers == regPlayers,
		}
		detailList.Store(gameId, gameDetail)

	})

	c.OnError(func(r *colly.Response, err error) {
		log.Println("Something went wrong:", err)
		log.Printf("resp: %#v ", r)
	})

	c.OnResponse(func(r *colly.Response) {
		fmt.Println("Visited", r.Request.URL)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	err := c.Visit(game.ListPage)
	if err != nil {
		panic(err)
	}

	return c, nil
}

func findWithoutTitle(selector string, e *goquery.Selection) string {
	node := e.Find(selector).First()
	node.Find("strong").Remove()
	text := strings.TrimPrefix(node.Text(), ":")
	return text
}

func sanitizeNumber(input string) (int, error) {
	// Define a regular expression to match all non-digit characters
	regex := regexp.MustCompile(`\D+`)

	// Replace all non-digit characters with an empty string
	sanitizedStr := regex.ReplaceAllString(input, "")

	// Convert the sanitized string to an integer
	sanitizedInt, err := strconv.Atoi(sanitizedStr)
	if err != nil {
		return 0, err // Return error if conversion fails
	}

	return sanitizedInt, nil
}

func shortDuration(d time.Duration) string {
	s := d.String()
	if strings.HasSuffix(s, "m0s") {
		s = s[:len(s)-2]
	}
	if strings.HasSuffix(s, "h0m") {
		s = s[:len(s)-2]
	}
	return s
}

// TODO: improve instead of using fixed to "march" instead of parsing spanish dates
func parseDate(dateStr string) (time.Time, error) {
	loc, _ := time.LoadLocation("Europe/Madrid")
	dateStr = strings.ReplaceAll(dateStr, "\n", "")
	dateStr = strings.ReplaceAll(dateStr, "marzo", "Mar")
	dateStr = strings.ReplaceAll(dateStr, "abril", "Apr")
	dateStr = strings.TrimSpace(dateStr)

	dateStr = removeFirstWord(dateStr)

	layout := "2 Jan 2006 15:04"
	// Parsing the date string into a time.Time object
	return time.ParseInLocation(layout, dateStr, loc)
}

func removeFirstWord(s string) string {
	// Find the index of the first whitespace
	index := strings.Index(s, " ")
	if index == -1 {
		// If there's no whitespace, return the original string
		return s
	}
	// Extract the substring starting from the position after the first whitespace
	return s[index+1:]
}
