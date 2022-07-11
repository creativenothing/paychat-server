package controllers

import (
	db "github.com/creativenothing/paychat-server/database"
	"github.com/creativenothing/paychat-server/models"
)

func NewChatMessage(t string, text string, userID string) *models.ChatMessage {
	var chat = models.ChatMessage{
		Type:   t,
		Text:   text,
		UserID: userID,
	}

	db.Instance.Create(&chat)

	return &chat
}
