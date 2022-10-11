package helpers

type Payload struct {
	TriesLeft int      `json:"triesLeft"`
	Type      string   `json:"type"`
	Realm     string   `json:"realm"`
	PlayerIDs []string `json:"playerIDs"`
}
