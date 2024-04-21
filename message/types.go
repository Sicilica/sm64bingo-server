package message

const (
	TYPE_HELLO = iota + 1
	TYPE_PLAYER_CONNECTED
	TYPE_PLAYER_DISCONNECTED
	TYPE_BOARD
	TYPE_PLAYER_LOCATION
	TYPE_PLAYER_COMPLETION
)

type Hello struct {
	Name           string `json:"name"`
	PreferredColor int    `json:"color"`
}

type PlayerConnected struct {
	Slot  int    `json:"p"`
	Name  string `json:"name"`
	Color uint32 `json:"color"`
}

type PlayerDisconnected struct {
	Slot int `json:"p"`
}

type Board struct {
	Seed   int32  `json:"s"`
	Config string `json:"cfg"`
}

type PlayerLocation struct {
	Slot     int   `json:"p"`
	Location int16 `json:"loc"`
}

type PlayerCompletion struct {
	Slot       int    `json:"p"`
	Completion uint32 `json:"com"`
}
