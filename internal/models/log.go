package models

import (
	"database/sql"
	"time"
)

type ActivityLog struct {
	ID           int64                `json:"id"`
	ActivityID   int64                `json:"activity_id"`
	ActivityName string               `json:"activity_name,omitempty"`
	CompletedAt  time.Time            `json:"completed_at"`
	Note         string               `json:"note"`
	Snapshots    []LogAbilitySnapshot `json:"snapshots,omitempty"`
}

type LogAbilitySnapshot struct {
	ID          int64   `json:"id"`
	LogID       int64   `json:"log_id"`
	AbilityID   int64   `json:"ability_id"`
	AbilityName string  `json:"ability_name,omitempty"`
	OldValue    float64 `json:"old_value"`
	NewValue    float64 `json:"new_value"`
}

type HistoryPoint struct {
	Date         time.Time `json:"date"`
	Value        float64   `json:"value"`
	ActivityName string    `json:"activity_name"`
}

func scanLog(scanner interface{ Scan(dest ...any) error }) (*ActivityLog, error) {
	var l ActivityLog
	var completedAtStr string
	err := scanner.Scan(&l.ID, &l.ActivityID, &l.ActivityName, &completedAtStr, &l.Note)
	if err != nil {
		return nil, err
	}
	l.CompletedAt = ParseTime(completedAtStr)
	return &l, nil
}

func ListLogs(db *sql.DB, limit, offset int, activityID, abilityID *int64) ([]ActivityLog, error) {
	query := `
		SELECT al.id, al.activity_id, COALESCE(a.name, '(deleted)'), al.completed_at, al.note
		FROM activity_logs al
		LEFT JOIN activities a ON a.id = al.activity_id
	`
	var args []any
	var conditions []string

	if activityID != nil {
		conditions = append(conditions, "al.activity_id = ?")
		args = append(args, *activityID)
	}
	if abilityID != nil {
		conditions = append(conditions, `al.id IN (
			SELECT DISTINCT las.log_id FROM log_ability_snapshots las WHERE las.ability_id = ?
		)`)
		args = append(args, *abilityID)
	}

	if len(conditions) > 0 {
		query += " WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += " AND " + conditions[i]
		}
	}

	query += " ORDER BY al.completed_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []ActivityLog
	for rows.Next() {
		l, err := scanLog(rows)
		if err != nil {
			return nil, err
		}
		logs = append(logs, *l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Load snapshots for each log
	for i := range logs {
		snapshots, err := GetSnapshotsByLog(db, logs[i].ID)
		if err != nil {
			return nil, err
		}
		logs[i].Snapshots = snapshots
	}

	return logs, nil
}

func GetSnapshotsByLog(db *sql.DB, logID int64) ([]LogAbilitySnapshot, error) {
	rows, err := db.Query(`
		SELECT las.id, las.log_id, las.ability_id, a.name, las.old_value, las.new_value
		FROM log_ability_snapshots las
		JOIN abilities a ON a.id = las.ability_id
		WHERE las.log_id = ?
	`, logID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []LogAbilitySnapshot
	for rows.Next() {
		var s LogAbilitySnapshot
		if err := rows.Scan(&s.ID, &s.LogID, &s.AbilityID, &s.AbilityName, &s.OldValue, &s.NewValue); err != nil {
			return nil, err
		}
		snapshots = append(snapshots, s)
	}
	return snapshots, rows.Err()
}

func CreateLog(db *sql.DB, log *ActivityLog) error {
	result, err := db.Exec(`INSERT INTO activity_logs (activity_id, completed_at, note) VALUES (?, ?, ?)`,
		log.ActivityID, FormatTime(log.CompletedAt), log.Note)
	if err != nil {
		return err
	}
	log.ID, _ = result.LastInsertId()

	for i := range log.Snapshots {
		log.Snapshots[i].LogID = log.ID
		_, err := db.Exec(`INSERT INTO log_ability_snapshots (log_id, ability_id, old_value, new_value) VALUES (?, ?, ?, ?)`,
			log.Snapshots[i].LogID, log.Snapshots[i].AbilityID, log.Snapshots[i].OldValue, log.Snapshots[i].NewValue)
		if err != nil {
			return err
		}
	}
	return nil
}

// Tx versions

func CreateLogTx(tx *sql.Tx, log *ActivityLog) error {
	result, err := tx.Exec(`INSERT INTO activity_logs (activity_id, completed_at, note) VALUES (?, ?, ?)`,
		log.ActivityID, FormatTime(log.CompletedAt), log.Note)
	if err != nil {
		return err
	}
	log.ID, _ = result.LastInsertId()

	for i := range log.Snapshots {
		log.Snapshots[i].LogID = log.ID
		_, err := tx.Exec(`INSERT INTO log_ability_snapshots (log_id, ability_id, old_value, new_value) VALUES (?, ?, ?, ?)`,
			log.Snapshots[i].LogID, log.Snapshots[i].AbilityID, log.Snapshots[i].OldValue, log.Snapshots[i].NewValue)
		if err != nil {
			return err
		}
	}
	return nil
}

func DeleteLog(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM activity_logs WHERE id = ?`, id)
	return err
}

func GetAbilityHistory(db *sql.DB, abilityID int64) ([]HistoryPoint, error) {
	rows, err := db.Query(`
		SELECT al.completed_at, las.new_value, COALESCE(a.name, 'Initial')
		FROM log_ability_snapshots las
		JOIN activity_logs al ON las.log_id = al.id
		LEFT JOIN activities a ON al.activity_id = a.id
		WHERE las.ability_id = ?
		ORDER BY al.completed_at ASC
	`, abilityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []HistoryPoint
	for rows.Next() {
		var p HistoryPoint
		var dateStr string
		if err := rows.Scan(&dateStr, &p.Value, &p.ActivityName); err != nil {
			return nil, err
		}
		p.Date = ParseTime(dateStr)
		points = append(points, p)
	}
	return points, rows.Err()
}
