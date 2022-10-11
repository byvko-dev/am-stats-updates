package helpers

type UpdateResult struct {
	AccountID string `json:"accountID" bson:"accountID"`
	Success   bool   `json:"success" bson:"success"`

	WillRetry bool   `json:"willRetry" bson:"willRetry"`
	Error     string `json:"error" bson:"error,omitempty"`
}
