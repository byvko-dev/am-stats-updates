package helpers

import "go.mongodb.org/mongo-driver/bson/primitive"

type Payload struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Timestamp    int                `json:"timestamp" bson:"timestamp"`
	TriesLeft    int                `json:"triesLeft" bson:"triesLeft"`
	IsProcessing bool               `json:"isProcessing" bson:"isProcessing"`

	Type      string   `json:"type" bson:"type"`
	Realm     string   `json:"realm" bson:"realm"`
	PlayerIDs []string `json:"playerIDs" bson:"playerIDs"`
}
