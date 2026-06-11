package models

import (
	"database/sql"
	"time"
)

type Activity struct {
	ID          int64            `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Effects     []ActivityEffect `json:"effects,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
}

func scanActivity(scanner interface{ Scan(dest ...any) error }) (*Activity, error) {
	var a Activity
	var createdAtStr string
	err := scanner.Scan(&a.ID, &a.Name, &a.Description, &createdAtStr)
	if err != nil {
		return nil, err
	}
	a.CreatedAt = ParseTime(createdAtStr)
	return &a, nil
}

func ListActivities(db *sql.DB) ([]Activity, error) {
	rows, err := db.Query(`SELECT id, name, description, created_at FROM activities ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []Activity
	for rows.Next() {
		a, err := scanActivity(rows)
		if err != nil {
			return nil, err
		}
		activities = append(activities, *a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Load effects for each activity
	for i := range activities {
		effects, err := GetEffectsByActivity(db, activities[i].ID)
		if err != nil {
			return nil, err
		}
		activities[i].Effects = effects
	}

	return activities, nil
}

func GetActivity(db *sql.DB, id int64) (*Activity, error) {
	a, err := scanActivity(db.QueryRow(`SELECT id, name, description, created_at FROM activities WHERE id = ?`, id))
	if err != nil {
		return nil, err
	}

	effects, err := GetEffectsByActivity(db, id)
	if err != nil {
		return nil, err
	}
	a.Effects = effects
	return a, nil
}

func CreateActivity(db *sql.DB, a *Activity) error {
	result, err := db.Exec(`INSERT INTO activities (name, description) VALUES (?, ?)`,
		a.Name, a.Description)
	if err != nil {
		return err
	}
	a.ID, _ = result.LastInsertId()

	// Insert effects
	for i := range a.Effects {
		a.Effects[i].ActivityID = a.ID
		_, err := db.Exec(`INSERT INTO activity_effects (activity_id, ability_id, boost_percentage) VALUES (?, ?, ?)`,
			a.Effects[i].ActivityID, a.Effects[i].AbilityID, a.Effects[i].BoostPercentage)
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdateActivity(db *sql.DB, a *Activity) error {
	_, err := db.Exec(`UPDATE activities SET name=?, description=? WHERE id=?`,
		a.Name, a.Description, a.ID)
	if err != nil {
		return err
	}

	// Replace effects: delete old, insert new
	if err := ReplaceEffects(db, a.ID, a.Effects); err != nil {
		return err
	}
	return nil
}

func DeleteActivity(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM activities WHERE id = ?`, id)
	return err
}

// Tx versions

func GetActivityTx(tx *sql.Tx, id int64) (*Activity, error) {
	return scanActivity(tx.QueryRow(`SELECT id, name, description, created_at FROM activities WHERE id = ?`, id))
}
