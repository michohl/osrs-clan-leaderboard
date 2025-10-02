package types

import "strings"

// Hiscores is the struct we will unmarshall our response from the API into
type Hiscores struct {
	Name       string            `json:"name"`
	Skills     []SkillHiscore    `json:"skills"`
	Activities []ActivityHiscore `json:"activities"`
}

func (h *Hiscores) GetSkill(name string) *SkillHiscore {
	for _, skill := range h.Skills {
		if strings.EqualFold(skill.Name, strings.Trim(name, " ")) {
			return &skill
		}
	}

	return nil
}

func (h *Hiscores) GetActivity(name string) *ActivityHiscore {
	for _, activity := range h.Activities {
		if strings.EqualFold(activity.Name, strings.Trim(name, " ")) {
			return &activity
		}
	}

	return nil
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

type RankedUser struct {
	User      UsersRow
	LocalRank int
	Rank      int
	XP        int // Used for skills
	Level     int // Used for skills
	Score     int // Used for activities
}

type RankedHiscores struct {
	Activity string
	Rankings []RankedUser
}
