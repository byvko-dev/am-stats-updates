package wotinspector

type VehicleInfo struct {
	Name    string `json:"en"`
	Tier    int    `json:"tier"`
	Type    int    `json:"type"`
	Premium int    `json:"premium"`
}
