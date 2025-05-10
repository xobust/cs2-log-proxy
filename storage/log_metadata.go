package storage

// LogMetadata stores per-log metadata (one file per ServerInstanceToken)
type LogMetadata struct {
	ServerInstanceToken string `json:"server_instance_token"`
	GameMap             string `json:"game_map"`
	SteamID             string `json:"steam_id"`
	ServerAddr          string `json:"server_addr"`
}
