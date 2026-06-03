package model

import "time"

type DomainEvent struct {
	ID            string    `gorm:"primaryKey"`
	TenantID      string    `gorm:"index"`
	Context       string    `gorm:"index"`
	AggregateID   string    `gorm:"index"`
	AggregateType string
	EventType     string    `gorm:"index"`
	SchemaVersion int
	Payload       string    `gorm:"type:jsonb"`
	OccurredAt    time.Time `gorm:"index"`
	RecordedAt    time.Time `gorm:"autoCreateTime"`
}

func (DomainEvent) TableName() string { return "domain_events" }
