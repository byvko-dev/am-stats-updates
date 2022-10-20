package helpers

type UpdateResult struct {
	AccountID string `json:"accountID" bson:"accountID"`

	WillRetry bool   `json:"willRetry" bson:"willRetry"`
	Success   bool   `json:"success" bson:"success"`
	Error     string `json:"error" bson:"error,omitempty"`
}
