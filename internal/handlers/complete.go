package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"grow/internal/db"
	"grow/internal/models"
	"grow/internal/service"
)

type CompleteRequest struct {
	Note        string `json:"note"`
	CompletedAt string `json:"completed_at"`
}

func RegisterCompleteRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/activities/{id}/complete", CompleteActivity)
}

func CompleteActivity(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	activityID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req CompleteRequest
	if r.Body != nil {
		json.NewDecoder(r.Body).Decode(&req)
	}

	completedAt := time.Now()
	if req.CompletedAt != "" {
		parsed, err := time.Parse(time.RFC3339, req.CompletedAt)
		if err != nil {
			parsed, err = time.Parse("2006-01-02 15:04:05", req.CompletedAt)
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid completed_at format, use ISO 8601")
				return
			}
		}
		completedAt = parsed
	}

	// Run everything in a transaction
	tx, err := db.DB.Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback()

	// Load activity
	activity, err := models.GetActivityTx(tx, activityID)
	if err != nil {
		writeError(w, http.StatusNotFound, "activity not found")
		return
	}

	// Load effects
	effects, err := models.GetEffectsByActivityTx(tx, activityID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var snapshots []service.CompleteActivitySnapshot
	var logSnapshots []models.LogAbilitySnapshot

	for _, effect := range effects {
		ability, err := models.GetAbilityTx(tx, effect.AbilityID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "ability not found: "+err.Error())
			return
		}

		oldValue := ability.CurrentValue
		newValue := service.CalculateGrowth(oldValue, effect.BoostPercentage, ability.GrowthRate)

		if err := models.UpdateAbilityValueTx(tx, ability.ID, newValue, completedAt); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		snapshots = append(snapshots, service.CompleteActivitySnapshot{
			AbilityID:       ability.ID,
			AbilityName:     ability.Name,
			OldValue:        oldValue,
			NewValue:        newValue,
			BoostPercentage: effect.BoostPercentage,
		})

		logSnapshots = append(logSnapshots, models.LogAbilitySnapshot{
			AbilityID:   ability.ID,
			AbilityName: ability.Name,
			OldValue:    oldValue,
			NewValue:    newValue,
		})
	}

	// Create log entry
	logEntry := &models.ActivityLog{
		ActivityID:  activityID,
		CompletedAt: completedAt,
		Note:        req.Note,
		Snapshots:   logSnapshots,
	}
	if err := models.CreateLogTx(tx, logEntry); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := service.CompleteActivityResponse{
		LogID: logEntry.ID,
		Activity: service.CompleteActivityInfo{
			ID:   activity.ID,
			Name: activity.Name,
		},
		CompletedAt: completedAt,
		Note:        req.Note,
		Snapshots:   snapshots,
	}

	writeJSON(w, http.StatusOK, response)
}
