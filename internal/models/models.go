package models

type MessageRequest struct {
	UserID   string                 `json:"user_id" binding:"required"`
	Action   string                 `json:"action" binding:"required"`
	Payload  interface{}            `json:"payload" binding:"required"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type MessageResponse struct {
	Status    string      `json:"status"`
	MessageID string      `json:"message_id,omitempty"`
	Error     string      `json:"error,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

type QueueMessage struct {
	ID        string      `json:"id"`
	UserID    string      `json:"user_id"`
	Action    string      `json:"action"`
	Payload   interface{} `json:"payload"`
	Timestamp int64       `json:"timestamp"`
	Metadata  interface{} `json:"metadata"`
}
