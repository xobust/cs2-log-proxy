package receiver

// CS2ServerHeaders represents all custom server-specific headers from CS2 log POSTs
// Add more fields as needed if new headers appear in the future.
type CS2ServerHeaders struct {
	GameMap             string `header:"X-Game-Map"`
	GameScoreCT         int    `header:"X-Game-Scorect"`
	GameScoreT          int    `header:"X-Game-Scoret"`
	GameState           string `header:"X-Game-State"`
	GameTeamCT          string `header:"X-Game-Teamct"`
	GameTeamT           string `header:"X-Game-Teamt"`
	LogBytesBeginOffset int    `header:"X-Logbytes-Beginoffset"`
	LogBytesEndOffset   int    `header:"X-Logbytes-Endoffset"`
	ServerAddr          string `header:"X-Server-Addr"`
	ServerInstanceToken string `header:"X-Server-Instance-Token"`
	SteamID             string `header:"X-Steamid"`
	TickEnd             int    `header:"X-Tick-End"`
	TickStart           int    `header:"X-Tick-Start"`
	Timestamp           string `header:"X-Timestamp"`
}
