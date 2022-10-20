package helpers

type UpdateTask struct {
	TriesLeft int      `json:"triesLeft" bson:"triesLeft"`
	Type      string   `json:"type" bson:"type"`
	Realm     string   `json:"realm" bson:"realm"`
	PlayerIDs []string `json:"playerIDs" bson:"playerIDs"`
}
