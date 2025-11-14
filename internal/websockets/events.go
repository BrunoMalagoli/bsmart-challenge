package websockets

import (
	"encoding/json"
	"log"
)

// Event types
const (
	EventProductCreated  = "product:created"
	EventProductUpdated  = "product:updated"
	EventProductDeleted  = "product:deleted"
	EventCategoryCreated = "category:created"
	EventCategoryUpdated = "category:updated"
	EventCategoryDeleted = "category:deleted"
)

type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func BroadcastEvent(hub *Hub, eventType string, data interface{}) {
	if hub == nil {
		return
	}

	event := Event{
		Type: eventType,
		Data: data,
	}

	message, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal WebSocket event: %v", err)
		return
	}

	// Broadcast to all clients
	hub.Broadcast(message)

	log.Printf("Broadcasted event '%s' to %d clients", eventType, hub.ClientCount())
}
