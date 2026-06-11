package handlers

import (
	"net/http"

	"grow/internal/db"
	"grow/internal/models"
	"grow/internal/service"
)

func RegisterNotionRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/export/notion", ExportNotion)
}

func ExportNotion(w http.ResponseWriter, r *http.Request) {
	apiKey, _ := models.GetSetting(db.DB, "notion_api_key")
	databaseID, _ := models.GetSetting(db.DB, "notion_database_id")

	if apiKey == "" || databaseID == "" {
		writeError(w, http.StatusBadRequest, "Notion API key or database ID not configured")
		return
	}

	result, err := service.SyncToNotion(db.DB, apiKey, databaseID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Update last sync time
	models.SetSetting(db.DB, "notion_last_sync", result.SyncedAt)

	writeJSON(w, http.StatusOK, result)
}
