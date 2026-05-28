package db

import (
	"encoding/json"

	"gorm.io/gorm"
)

func CreateAuditEvent(database *gorm.DB, event AuditEvent) (*AuditEvent, error) {
	if err := database.Create(&event).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

func CreateToolAuditEvent(database *gorm.DB, sessionID uint, toolName string, input any, output any, runErr error) (*AuditEvent, error) {
	return CreateToolAuditEventWithType(database, sessionID, "tool.run", toolName, input, output, runErr)
}

func CreateToolAuditEventWithType(database *gorm.DB, sessionID uint, eventType string, toolName string, input any, output any, runErr error) (*AuditEvent, error) {
	inputJSON, _ := json.Marshal(input)
	outputJSON, _ := json.Marshal(output)
	event := AuditEvent{
		SessionID:  sessionID,
		Type:       eventType,
		ToolName:   toolName,
		InputJSON:  string(inputJSON),
		OutputJSON: string(outputJSON),
	}
	if runErr != nil {
		event.Error = runErr.Error()
	}
	return CreateAuditEvent(database, event)
}

func ListAuditEvents(database *gorm.DB, limit int) ([]AuditEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	var events []AuditEvent
	if err := database.Order("created_at desc, id desc").Limit(limit).Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}
