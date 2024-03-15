package game

import "time"

const (
	ListPage = "https://app.netconplay.com/games?page=1"
)

type Detail struct {
	Id                string
	Link              string
	BackgroundImage   string
	Title             string
	System            string
	Description       string
	MasterName        string
	MasterDescription string
	StartDate         time.Time
	Duration          string
	EndDate           time.Time
	Security          string
	SensibleContent   string
	Platform          string
	Channel           string
	Streamed          bool
	InitiationGame    bool
	MaxPlayers        int
	RegisteredPlayers int
}
