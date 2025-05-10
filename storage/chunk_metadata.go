package storage

// ChunkMeta holds metadata about a received log chunk
// Used for idempotency and audit
// The JSON file will contain a list of these for each ServerInstanceToken
// (e.g. logs/{ServerInstanceToken}_chunks.json)
type ChunkMeta struct {
	BeginOffset         int    `json:"begin_offset"`
	EndOffset           int    `json:"end_offset"`
	GameScoreCT         int    `json:"game_score_ct"`
	GameScoreT          int    `json:"game_score_t"`
	GameState           string `json:"game_state"`
	GameTeamCT          string `json:"game_team_ct"`
	GameTeamT           string `json:"game_team_t"`
	TickEnd             int    `json:"tick_end"`
	TickStart           int    `json:"tick_start"`
	Timestamp           string `json:"timestamp"`
}
