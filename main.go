package main

import (
	"context"
	"fmt"
	"github.com/puzpuzpuz/xsync/v3"
	"log"

	"github.com/guumaster/netcon-metadata/collector"
	"github.com/guumaster/netcon-metadata/game"
	"github.com/guumaster/netcon-metadata/sheets"
)

func main() {

	ctx := context.Background()

	gameList := xsync.NewMapOf[string, game.Detail]()

	c, err := collector.MakeGameListCollector(gameList)
	if err != nil {
		log.Fatal(err)
		return
	}
	err = c.Visit(game.ListPage)
	if err != nil {
		log.Fatal(err)
		return
	}
	c.Wait()

	c, err = collector.MakeGameDetailCollector(gameList)
	if err != nil {
		log.Fatal(err)
	}

	gameList.Range(func(key string, game game.Detail) bool {
		_ = c.Visit(game.Link)
		return true
	})
	c.Wait()

	fmt.Println("Total games: ", gameList.Size())

	err = sheets.SaveToGoogleSheet(ctx, gameList)
	if err != nil {
		log.Fatal(err)
	}

	err = sheets.SaveToGoogleCalendarSheet(ctx, gameList)
	if err != nil {
		log.Fatal(err)
	}
}
