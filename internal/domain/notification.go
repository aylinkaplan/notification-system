package domain

import (
	"time"

	"github.com/google/uuid"
)

type Channel string

const (
	ChannelSMS   Channel = "sms"
	ChannelEmail Channel = "email"
	ChannelPush  Channel = "push"
)

type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityNormal Priority = "normal"
	PriorityLow    Priority = "low"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusDelivered  Status = "delivered"
	StatusFailed     Status = "failed"
	StatusCancelled  Status = "cancelled"
)

type Notification struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	BatchID         *uuid.UUID `json:"batch_id,omitempty" db:"batch_id"`
	Recipient       string     `json:"recipient" db:"recipient"`
	Channel         Channel    `json:"channel" db:"channel"`
	Content         string     `json:"content" db:"content"`
	Priority        Priority   `json:"priority" db:"priority"`
	Status          Status     `json:"status" db:"status"`
	IdempotencyKey  *string    `json:"idempotency_key,omitempty" db:"idempotency_key"`
	ProviderMsgID   *string    `json:"provider_msg_id,omitempty" db:"provider_msg_id"`
	RetryCount      int        `json:"retry_count" db:"retry_count"`
	MaxRetries      int        `json:"max_retries" db:"max_retries"`
	ScheduledAt     *time.Time `json:"scheduled_at,omitempty" db:"scheduled_at"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
	ErrorMessage    *string    `json:"error_message,omitempty" db:"error_message"`
}

type CreateNotificationRequest struct {
	Recipient      string   `json:"recipient"`
	Channel        Channel  `json:"channel"`
	Content        string   `json:"content"`
	Priority       Priority `json:"priority"`
	IdempotencyKey *string  `json:"idempotency_key,omitempty"`
	ScheduledAt    *time.Time `json:"scheduled_at,omitempty"`
}

type ListFilter struct {
	Status   *Status  `json:"status,omitempty"`
	Channel  *Channel `json:"channel,omitempty"`
	From     *time.Time `json:"from,omitempty"`
	To       *time.Time `json:"to,omitempty"`
	Limit    int       `json:"limit"`
	Offset   int       `json:"offset"`
}
