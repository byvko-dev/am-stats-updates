package blitzstars

type TankAverages struct {
	TankID int `json:"tank_id" bson:"tankId"`
	All    struct {
		Battles              float64 `json:"battles,omitempty" bson:"battles,omitempty"`
		DroppedCapturePoints float64 `json:"dropped_capture_points,omitempty" bson:"droppedCapturePoints,omitempty"`
	} `json:",omitempty" bson:"all"`
	Special struct {
		Winrate         float64 `json:"winrate,omitempty" bson:"winrate,omitempty"`
		KillsPerBattle  float64 `json:"killsPerBattle,omitempty" bson:"killsPerBattle,omitempty"`
		SpotsPerBattle  float64 `json:"spotsPerBattle,omitempty" bson:"spotsPerBattle,omitempty"`
		DamagePerBattle float64 `json:"damagePerBattle,omitempty" bson:"damagePerBattle,omitempty"`
	} `json:"special,omitempty" bson:"special"`
}
