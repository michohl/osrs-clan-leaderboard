package types

// Hiscores is the struct we will unmarshall our response from the API into
type Hiscores struct {
	Name       string            `json:"name"`
	Skills     []SkillHiscore    `json:"skills"`
	Activities []ActivityHiscore `json:"activities"`
}

// SkillHiscore is a representation for a "skilling" activity
type SkillHiscore struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Rank  int    `json:"rank"`
	Level int    `json:"level"`
	XP    int    `json:"xp"`
}

// ActivityHiscore is a representation for a non-skilling activity
// such as clues or PVM
type ActivityHiscore struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Rank  int    `json:"rank"`
	Score int    `json:"score"`
}
