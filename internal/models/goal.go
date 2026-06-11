package models

import (
	"database/sql"
	"time"
)

type Goal struct {
	ID          int64      `json:"id"`
	AbilityID   int64      `json:"ability_id"`
	AbilityName string     `json:"ability_name,omitempty"`
	TargetValue float64    `json:"target_value"`
	Deadline    *time.Time `json:"deadline"`
	IsAchieved  bool       `json:"is_achieved"`
	CreatedAt   time.Time  `json:"created_at"`
	// Computed
	CurrentValue float64 `json:"current_value"`
	Progress     float64 `json:"progress"` // 0-100
}

func scanGoal(scanner interface{ Scan(dest ...any) error }) (*Goal, error) {
	var g Goal
	var deadline sql.NullString
	var createdAtStr string
	var isAchieved int
	err := scanner.Scan(&g.ID, &g.AbilityID, &g.AbilityName, &g.TargetValue,
		&deadline, &isAchieved, &createdAtStr)
	if err != nil {
		return nil, err
	}
	g.IsAchieved = isAchieved == 1
	g.CreatedAt = ParseTime(createdAtStr)
	if deadline.Valid {
		t, _ := time.Parse("2006-01-02", deadline.String)
		g.Deadline = &t
	}
	return &g, nil
}

func ListGoals(db *sql.DB) ([]Goal, error) {
	rows, err := db.Query(`
		SELECT g.id, g.ability_id, a.name, g.target_value, g.deadline, g.is_achieved, g.created_at
		FROM goals g
		JOIN abilities a ON a.id = g.ability_id
		ORDER BY g.is_achieved ASC, g.deadline ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var goals []Goal
	for rows.Next() {
		g, err := scanGoal(rows)
		if err != nil {
			return nil, err
		}
		// Get current ability value for progress calculation
		var currentValue float64
		db.QueryRow(`SELECT current_value FROM abilities WHERE id = ?`, g.AbilityID).Scan(&currentValue)
		g.CurrentValue = currentValue
		if g.TargetValue > 0 {
			g.Progress = (currentValue / g.TargetValue) * 100
			if g.Progress > 100 {
				g.Progress = 100
			}
		}
		goals = append(goals, *g)
	}
	return goals, rows.Err()
}

func CreateGoal(db *sql.DB, g *Goal) error {
	var deadlineStr *string
	if g.Deadline != nil {
		s := g.Deadline.Format("2006-01-02")
		deadlineStr = &s
	}
	result, err := db.Exec(`INSERT INTO goals (ability_id, target_value, deadline) VALUES (?, ?, ?)`,
		g.AbilityID, g.TargetValue, deadlineStr)
	if err != nil {
		return err
	}
	g.ID, _ = result.LastInsertId()
	return nil
}

func UpdateGoal(db *sql.DB, g *Goal) error {
	var deadlineStr *string
	if g.Deadline != nil {
		s := g.Deadline.Format("2006-01-02")
		deadlineStr = &s
	}
	isAchieved := 0
	if g.IsAchieved {
		isAchieved = 1
	}
	_, err := db.Exec(`UPDATE goals SET target_value=?, deadline=?, is_achieved=? WHERE id=?`,
		g.TargetValue, deadlineStr, isAchieved, g.ID)
	return err
}

func DeleteGoal(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM goals WHERE id = ?`, id)
	return err
}
