package sheets

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"

	"github.com/guumaster/netcon-metadata/game"
	"github.com/puzpuzpuz/xsync/v3"
)

const (
	spreadsheetID = "1OeJnf9Eq5EuFn23s0n553jiGcGuXuoafREjz32T9i6I"
	gamesRange    = "Partidas!A:Z"
	calendarRange = "Calendario!B:T"
	credentials   = "key.json"
)

func SaveToGoogleSheet(ctx context.Context, list *xsync.MapOf[string, game.Detail]) error {
	svc, err := makeGoogleService(ctx)
	if err != nil {
		return err
	}

	//Append value to the sheet.
	rows := &sheets.ValueRange{
		Values: [][]any{
			{
				"game_id", "title", "system", "description", "master_name", "master_description",
				"start_date", "duration", "end_date", "security", "sensible_content", "platform",
				"channel", "streamed", "initiation_game", "max_players", "registered_players", "completed",
			},
		},
	}

	keys := sortGameIndex(list)
	for _, key := range keys {
		gameData, _ := list.Load(key)
		rows.Values = append(rows.Values, []any{
			gameData.Id, gameData.Title, gameData.System, gameData.Description, gameData.MasterName, gameData.MasterDescription,
			gameData.StartDate, gameData.Duration, gameData.EndDate, gameData.Security, gameData.SensibleContent, gameData.Platform,
			gameData.Channel, gameData.Streamed, gameData.InitiationGame, gameData.MaxPlayers, gameData.RegisteredPlayers, gameData.Completed,
		})
	}

	rsp, err := svc.Spreadsheets.Values.
		Update(spreadsheetID, gamesRange, rows).
		ValueInputOption("USER_ENTERED").
		//InsertDataOption("OVERWRITE").
		Context(ctx).Do()
	if err != nil || rsp.HTTPStatusCode != 200 {
		log.Fatalf("Unable to write Google Sheets: %v", err)
		return err
	}
	return nil
}

func SaveToGoogleCalendarSheet(ctx context.Context, list *xsync.MapOf[string, game.Detail]) error {
	svc, err := makeGoogleService(ctx)
	if err != nil {
		return err
	}

	//Append value to the sheet.
	rows := &sheets.ValueRange{
		Values: [][]any{
			{"Update", "Title", "Start", "End", "Start Time", "End Time",
				"Repeat", "Interval", "Count", "Until", "By Day",
				"Description", "Location", "Timezone",
			},
		},
	}

	keys := sortGameIndex(list)
	for _, key := range keys {
		streamed := "No"
		streamMark := ""
		game, _ := list.Load(key)

		if game.Streamed {
			streamed = "Si"
			streamMark = "🎥"
		}
		hasOpenSeats := "🔒"
		freeSeats := game.MaxPlayers - game.RegisteredPlayers
		if freeSeats > 0 {
			hasOpenSeats = "✨"
		}
		title := fmt.Sprintf("%s [%d/%d] %s %s", hasOpenSeats, freeSeats, game.MaxPlayers, game.Title, streamMark)
		start := game.StartDate.Format("2006-01-02")
		end := game.EndDate.Format("2006-01-02")
		startTime := game.StartDate.Format("15:04:05")
		endTime := game.EndDate.Format("15:04:05")

		description := fmt.Sprintf(`🔗 Enlace: %s
👥 Plazas libres: %d/%d
🧙🏻 Organizadora: %s
🎲 Sistema: %s
🎥 Emitida: %s
---			
📝 Descripción: 
%s

`, game.Link, freeSeats, game.MaxPlayers, game.MasterName, game.System, streamed, game.Description)

		rows.Values = append(rows.Values, []any{
			"TRUE", title, start, end, startTime, endTime,
			"", "", "", "", "",
			description, "", "Europe/Madrid",
		})

	}

	rsp, err := svc.Spreadsheets.Values.
		Update(spreadsheetID, calendarRange, rows).
		ValueInputOption("USER_ENTERED").
		//InsertDataOption("OVERWRITE").
		Context(ctx).Do()
	if err != nil || rsp.HTTPStatusCode != 200 {
		log.Fatalf("Unable to write Google Sheets: %v", err)
		return err
	}

	return nil
}

func sortGameIndex(gameList *xsync.MapOf[string, game.Detail]) []string {
	// Extract keys (index strings)
	var keys []string
	gameList.Range(func(key string, value game.Detail) bool {
		keys = append(keys, key)
		return true
	})

	// Sort keys
	sort.Slice(keys, func(i, j int) bool {
		index1, _ := strconv.Atoi(keys[i])
		index2, _ := strconv.Atoi(keys[j])
		return index1 < index2
	})
	return keys
}

func makeGoogleService(ctx context.Context) (*sheets.Service, error) {
	// Load the Google Sheets API credentials from your JSON file.
	creds, err := os.ReadFile(credentials)
	if err != nil {
		log.Fatalf("Unable to read credentials file: %v", err)
		return nil, err
	}

	config, err := google.JWTConfigFromJSON(creds, sheets.SpreadsheetsScope)
	if err != nil {
		log.Fatalf("Unable to create JWT config: %v", err)
		return nil, err
	}

	client := config.Client(ctx)
	sheetsService, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to create Google Sheets service: %v", err)
		return nil, err
	}
	return sheetsService, nil
}
