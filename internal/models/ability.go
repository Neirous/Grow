package models

import (
	"database/sql"
	"time"
)

type Ability struct {
	ID                   int64      `json:"id"`
	Name                 string     `json:"name"`
	Description          string     `json:"description"`
	BaseValue            float64    `json:"base_value"`
	CurrentValue         float64    `json:"current_value"`
	GrowthRate           float64    `json:"growth_rate"`
	DecayRate            float64    `json:"decay_rate"`
	LastActivityAt       *time.Time `json:"last_activity_at"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	// Computed fields (not stored)
	EffectiveValue        float64 `json:"effective_value"`
	DaysSinceLastActivity float64 `json:"days_since_last_activity"`
}

// scanAbility scans a row into an Ability, handling SQLite text datetime fields.
func scanAbility(scanner interface {
	Scan(dest ...any) error
}) (*Ability, error) {
	var a Ability
	var lastAct sql.NullString
	var createdAtStr, updatedAtStr string

	err := scanner.Scan(&a.ID, &a.Name, &a.Description, &a.BaseValue,
		&a.CurrentValue, &a.GrowthRate, &a.DecayRate, &lastAct, &createdAtStr, &updatedAtStr)
	if err != nil {
		return nil, err
	}

	a.CreatedAt = ParseTime(createdAtStr)
	a.UpdatedAt = ParseTime(updatedAtStr)
	if lastAct.Valid {
		t := ParseTime(lastAct.String)
		a.LastActivityAt = &t
	}

	return &a, nil
}

func ListAbilities(db *sql.DB) ([]Ability, error) {
	rows, err := db.Query(`
		SELECT id, name, description, base_value, current_value,
		       growth_rate, decay_rate, last_activity_at, created_at, updated_at
		FROM abilities ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var abilities []Ability
	for rows.Next() {
		a, err := scanAbility(rows)
		if err != nil {
			return nil, err
		}
		abilities = append(abilities, *a)
	}
	return abilities, rows.Err()
}

func GetAbility(db *sql.DB, id int64) (*Ability, error) {
	return scanAbility(db.QueryRow(`
		SELECT id, name, description, base_value, current_value,
		       growth_rate, decay_rate, last_activity_at, created_at, updated_at
		FROM abilities WHERE id = ?
	`, id))
}

func CreateAbility(db *sql.DB, a *Ability) error {
	result, err := db.Exec(`
		INSERT INTO abilities (name, description, base_value, current_value, growth_rate, decay_rate)
		VALUES (?, ?, ?, ?, ?, ?)
	`, a.Name, a.Description, a.BaseValue, a.CurrentValue, a.GrowthRate, a.DecayRate)
	if err != nil {
		return err
	}
	a.ID, _ = result.LastInsertId()
	return nil
}

func UpdateAbility(db *sql.DB, a *Ability) error {
	_, err := db.Exec(`
		UPDATE abilities SET name=?, description=?, base_value=?, current_value=?,
		       growth_rate=?, decay_rate=?, updated_at=datetime('now')
		WHERE id=?
	`, a.Name, a.Description, a.BaseValue, a.CurrentValue,
		a.GrowthRate, a.DecayRate, a.ID)
	return err
}

func DeleteAbility(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM abilities WHERE id = ?`, id)
	return err
}

func UpdateAbilityValue(db *sql.DB, id int64, newValue float64, activityTime time.Time) error {
	_, err := db.Exec(`
		UPDATE abilities SET current_value=?, last_activity_at=?, updated_at=datetime('now')
		WHERE id=?
	`, newValue, FormatTime(activityTime), id)
	return err
}

// Tx versions for use inside transactions

func GetAbilityTx(tx *sql.Tx, id int64) (*Ability, error) {
	return scanAbility(tx.QueryRow(`
		SELECT id, name, description, base_value, current_value,
		       growth_rate, decay_rate, last_activity_at, created_at, updated_at
		FROM abilities WHERE id = ?
	`, id))
}

func UpdateAbilityValueTx(tx *sql.Tx, id int64, newValue float64, activityTime time.Time) error {
	_, err := tx.Exec(`
		UPDATE abilities SET current_value=?, last_activity_at=?, updated_at=datetime('now')
		WHERE id=?
	`, newValue, FormatTime(activityTime), id)
	return err
}
