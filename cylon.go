package cylon

type AI interface {
	Think(s *RobotState) *RobotCommands
}

type RobotState struct {
	Position     Point   `json:"position"`
	Heading      float64 `json:"heading"`
	GunHeading   float64 `json:"gunHeading"`
	RadarHeading float64 `json:"radarHeading"`
	Velocity     float64 `json:"velocity"`
	Heat         float64 `json:"heat"`
	Energy       float64 `json:"energy"`
	Alive        bool    `json:"alive"`
}

type RobotCommands struct {
	Turn       float64 `json:"turn"`
	TurnGun    float64 `json:"turnGun"`
	TurnRadar  float64 `json:"turnRadar"`
	Accelerate float64 `json:"accelerate"`
	Fire       float64 `json:"fire"`
}

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}
