package service

import (
	"database/sql"
	"time"

	"grow/internal/models"
)

// AbilityWithHistory is the response shape for ability history endpoint.
type AbilityWithHistory struct {
	Ability models.Ability          `json:"ability"`
	Points  []models.HistoryPoint   `json:"points"`
}

// GetAbilityWithHistory returns an ability with its growth history.
// It prepends a synthetic "Initial" data point at the ability's creation time.
func GetAbilityWithHistory(db *sql.DB, abilityID int64) (*AbilityWithHistory, error) {
	a, err := models.GetAbility(db, abilityID)
	if err != nil {
		return nil, err
	}

	// Apply decay to the current value
	ApplyDecay(a)

	// Get history points
	points, err := models.GetAbilityHistory(db, abilityID)
	if err != nil {
		return nil, err
	}

	// Prepend initial data point
	initialPoint := models.HistoryPoint{
		Date:         a.CreatedAt,
		Value:        a.BaseValue,
		ActivityName: "初始值",
	}
	allPoints := make([]models.HistoryPoint, 0, len(points)+1)
	allPoints = append(allPoints, initialPoint)
	allPoints = append(allPoints, points...)

	return &AbilityWithHistory{
		Ability: *a,
		Points:  allPoints,
	}, nil
}

// CompleteActivityResponse is the response after completing an activity.
type CompleteActivityResponse struct {
	LogID       int64                        `json:"log_id"`
	Activity    CompleteActivityInfo         `json:"activity"`
	CompletedAt time.Time                    `json:"completed_at"`
	Note        string                       `json:"note"`
	Snapshots   []CompleteActivitySnapshot   `json:"snapshots"`
}

type CompleteActivityInfo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type CompleteActivitySnapshot struct {
	AbilityID       int64   `json:"ability_id"`
	AbilityName     string  `json:"ability_name"`
	OldValue        float64 `json:"old_value"`
	NewValue        float64 `json:"new_value"`
	BoostPercentage float64 `json:"boost_percentage"`
}
