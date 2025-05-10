package storage

// LogMeta holds metadata about a single log session
// LogID is unique per log: ServerInstanceToken + LogStartTime
// LogStartTime is the timestamp of the first chunk with BeginOffset == 0
// GameMap and ServerAddr are set from the first chunk

type LogMeta struct {
	LogID          string `json:"log_id"`
	LogStartTime   string `json:"log_start_time"`
	GameMap        string `json:"game_map"`
	ServerAddr     string `json:"server_addr"`
	LastActivity   string `json:"last_activity"`
	LastByteOffset int    `json:"last_byte_offset"`
}

type ServerMeta struct {
	ServerInstanceToken string    `json:"server_instance_token"`
	SteamID             string    `json:"steam_id"`
	Logs                []LogMeta `json:"logs"`
}

// Legacy: LogMetadata stores per-log metadata (one file per ServerInstanceToken)
type LogMetadata struct {
	ServerInstanceToken string `json:"server_instance_token"`
	GameMap             string `json:"game_map"`
	SteamID             string `json:"steam_id"`
	ServerAddr          string `json:"server_addr"`
	LastActivity        string `json:"last_activity"`
}
