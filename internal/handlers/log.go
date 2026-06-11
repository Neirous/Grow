package handlers

import (
	"net/http"
	"strconv"

	"grow/internal/db"
	"grow/internal/models"
)

func RegisterLogRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/logs", ListLogs)
	mux.HandleFunc("DELETE /api/logs/{id}", DeleteLog)
}

func ListLogs(w http.ResponseWriter, r *http.Request) {
	limit := 20
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	var activityID *int64
	var abilityID *int64

	if a := r.URL.Query().Get("activity_id"); a != "" {
		v, err := strconv.ParseInt(a, 10, 64)
		if err == nil {
			activityID = &v
		}
	}
	if a := r.URL.Query().Get("ability_id"); a != "" {
		v, err := strconv.ParseInt(a, 10, 64)
		if err == nil {
			abilityID = &v
		}
	}

	logs, err := models.ListLogs(db.DB, limit, offset, activityID, abilityID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if logs == nil {
		logs = []models.ActivityLog{}
	}

	writeJSON(w, http.StatusOK, logs)
}

func DeleteLog(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := models.DeleteLog(db.DB, id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
