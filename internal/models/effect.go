package models

import "database/sql"

type ActivityEffect struct {
	ID              int64   `json:"id"`
	ActivityID      int64   `json:"activity_id"`
	AbilityID       int64   `json:"ability_id"`
	AbilityName     string  `json:"ability_name,omitempty"`
	BoostPercentage float64 `json:"boost_percentage"`
}

func GetEffectsByActivity(db *sql.DB, activityID int64) ([]ActivityEffect, error) {
	rows, err := db.Query(`
		SELECT ae.id, ae.activity_id, ae.ability_id, a.name, ae.boost_percentage
		FROM activity_effects ae
		JOIN abilities a ON a.id = ae.ability_id
		WHERE ae.activity_id = ?
	`, activityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var effects []ActivityEffect
	for rows.Next() {
		var e ActivityEffect
		if err := rows.Scan(&e.ID, &e.ActivityID, &e.AbilityID, &e.AbilityName, &e.BoostPercentage); err != nil {
			return nil, err
		}
		effects = append(effects, e)
	}
	return effects, rows.Err()
}

func ReplaceEffects(db *sql.DB, activityID int64, effects []ActivityEffect) error {
	_, err := db.Exec(`DELETE FROM activity_effects WHERE activity_id = ?`, activityID)
	if err != nil {
		return err
	}
	for _, e := range effects {
		_, err := db.Exec(`INSERT INTO activity_effects (activity_id, ability_id, boost_percentage) VALUES (?, ?, ?)`,
			activityID, e.AbilityID, e.BoostPercentage)
		if err != nil {
			return err
		}
	}
	return nil
}

// Tx versions

func GetEffectsByActivityTx(tx *sql.Tx, activityID int64) ([]ActivityEffect, error) {
	rows, err := tx.Query(`
		SELECT ae.id, ae.activity_id, ae.ability_id, a.name, ae.boost_percentage
		FROM activity_effects ae
		JOIN abilities a ON a.id = ae.ability_id
		WHERE ae.activity_id = ?
	`, activityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var effects []ActivityEffect
	for rows.Next() {
		var e ActivityEffect
		if err := rows.Scan(&e.ID, &e.ActivityID, &e.AbilityID, &e.AbilityName, &e.BoostPercentage); err != nil {
			return nil, err
		}
		effects = append(effects, e)
	}
	return effects, rows.Err()
}
