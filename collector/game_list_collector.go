package collector

import (
	"fmt"
	"github.com/gocolly/colly"
	"github.com/guumaster/netcon-metadata/game"
	"github.com/puzpuzpuz/xsync/v3"
	"log"
	"net/url"
	"path"
)

func MakeGameListCollector(gameList *xsync.MapOf[string, game.Detail]) (*colly.Collector, error) {
	c := colly.NewCollector(
		// MaxDepth is 2, so only the links on the scraped page
		// and links on those pages are visited
		colly.AllowedDomains("netconplay.com", "app.netconplay.com", "www.netconplay.com", "*.netconplay.com"),
		//colly.MaxDepth(5),
		colly.Async(true),
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"),
	)

	// Limit the maximum parallelism to 2
	// This is necessary if the goroutines are dynamically
	// created to control the limit of simultaneous requests.
	//
	// Parallelism can be controlled also by spawning fixed
	// number of go routines.
	_ = c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 5})

	// Find and visit all links
	c.OnHTML("div.games a", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		u, err := url.Parse(link)
		if err != nil {
			fmt.Println("Error parsing URL:", err)
			return
		}
		gameId := path.Base(u.Path)
		if gameId == "games" {
			return
		}

		title := e.DOM.Find("p.game_title").First().Text()
		system := e.DOM.Find("p:nth-child(2)").First().Text()

		gameList.Store(gameId, game.Detail{
			Id:     gameId,
			Title:  title,
			System: system,
			Link:   link,
		})
	})

	c.OnHTML("ul.pagination li a[href]", func(e *colly.HTMLElement) {
		_ = e.Request.Visit(e.Attr("href"))
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

	return c, nil
}
