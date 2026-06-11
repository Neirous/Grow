package service

import (
	"database/sql"
	"log"
	"time"

	"grow/internal/models"
)

func StartScheduler(db *sql.DB) {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			now := time.Now()

			// Check Feishu reminder
			checkFeishuReminder(db, now)

			// Check Notion daily sync (at 23:00)
			checkNotionSync(db, now)
		}
	}()
	log.Println("Scheduler started")
}

func checkFeishuReminder(db *sql.DB, now time.Time) {
	enabled, _ := models.GetSetting(db, "reminder_enabled")
	if enabled != "true" {
		return
	}

	reminderTime, _ := models.GetSetting(db, "reminder_time")
	if reminderTime == "" {
		reminderTime = "20:00"
	}

	currentTime := now.Format("15:04")
	if currentTime != reminderTime {
		return
	}

	// Check if already sent today
	lastSent, _ := models.GetSetting(db, "reminder_last_sent")
	today := now.Format("2006-01-02")
	if lastSent == today {
		return
	}

	webhookURL, _ := models.GetSetting(db, "feishu_webhook_url")
	if webhookURL == "" {
		return
	}

	if err := SendFeishuReminder(db, webhookURL); err != nil {
		log.Printf("Feishu reminder failed: %v", err)
		return
	}

	models.SetSetting(db, "reminder_last_sent", today)
}

func checkNotionSync(db *sql.DB, now time.Time) {
	enabled, _ := models.GetSetting(db, "notion_sync_enabled")
	if enabled != "true" {
		return
	}

	// Run at 23:00
	if now.Format("15:04") != "23:00" {
		return
	}

	// Check if already synced today
	lastSync, _ := models.GetSetting(db, "notion_last_sync")
	today := now.Format("2006-01-02")
	if lastSync == today {
		return
	}

	apiKey, _ := models.GetSetting(db, "notion_api_key")
	databaseID, _ := models.GetSetting(db, "notion_database_id")
	if apiKey == "" || databaseID == "" {
		return
	}

	result, err := SyncToNotion(db, apiKey, databaseID)
	if err != nil {
		log.Printf("Notion auto sync failed: %v", err)
		return
	}

	models.SetSetting(db, "notion_last_sync", result.SyncedAt)
	log.Printf("Notion auto sync: %d updated, %d created", result.Updated, result.Created)
}
