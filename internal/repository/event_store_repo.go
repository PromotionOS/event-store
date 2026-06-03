package repository

import (
	"github.com/promotionos/event-store/internal/model"
	"gorm.io/gorm"
)

type EventStoreRepository interface {
	Append(event *model.DomainEvent) error
	FindByAggregate(aggregateID string, tenantID string) ([]*model.DomainEvent, error)
	FindByContext(context string, tenantID string, from string, to string) ([]*model.DomainEvent, error)
}

type EventStoreRepositoryImpl struct {
	db *gorm.DB
}

func NewEventStoreRepositoryImpl(db *gorm.DB) *EventStoreRepositoryImpl {
	return &EventStoreRepositoryImpl{db: db}
}

func (r *EventStoreRepositoryImpl) Append(event *model.DomainEvent) error {
	// Append-only — never UPDATE, never DELETE
	return r.db.Create(event).Error
}

func (r *EventStoreRepositoryImpl) FindByAggregate(aggregateID, tenantID string) ([]*model.DomainEvent, error) {
	var events []*model.DomainEvent
	err := r.db.Where("aggregate_id = ? AND tenant_id = ?", aggregateID, tenantID).
		Order("occurred_at ASC").
		Find(&events).Error
	return events, err
}

func (r *EventStoreRepositoryImpl) FindByContext(context, tenantID, from, to string) ([]*model.DomainEvent, error) {
	var events []*model.DomainEvent
	query := r.db.Where("context = ? AND tenant_id = ?", context, tenantID)
	if from != "" {
		query = query.Where("occurred_at >= ?", from)
	}
	if to != "" {
		query = query.Where("occurred_at <= ?", to)
	}
	err := query.Order("occurred_at ASC").Find(&events).Error
	return events, err
}
