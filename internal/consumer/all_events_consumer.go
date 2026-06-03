package consumer

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/promotionos/event-store/internal/model"
	"github.com/promotionos/event-store/internal/repository"
)

var allChannels = []string{
	"promotionos.campaign.published",
	"promotionos.campaign.paused",
	"promotionos.catalog.item.excluded",
	"promotionos.customer.segment.updated",
	"promotionos.redemption.redeemed",
	"promotionos.redemption.claim.submitted",
	"promotionos.analytics.budget.exhausted",
	"promotionos.campaign.budget.updated",
}

type AllEventsConsumer struct {
	repo   repository.EventStoreRepository
	client *redis.Client
}

func NewAllEventsConsumer(repo repository.EventStoreRepository, client *redis.Client) *AllEventsConsumer {
	return &AllEventsConsumer{repo: repo, client: client}
}

func (c *AllEventsConsumer) Start() {
	pubsub := c.client.Subscribe(context.Background(), allChannels...)
	log.Println("Event Store: subscribed to all channels")

	go func() {
		for msg := range pubsub.Channel() {
			c.handle(msg)
		}
	}()
}

func (c *AllEventsConsumer) handle(msg *redis.Message) {
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(msg.Payload), &raw); err != nil {
		log.Printf("Event Store: failed to parse event on %s: %v", msg.Channel, err)
		return
	}

	event := &model.DomainEvent{
		ID:            uuid.New().String(),
		TenantID:      str(raw["tenantId"]),
		Context:       extractContext(msg.Channel),
		AggregateID:   str(raw["campaignId"]),
		AggregateType: extractAggregateType(msg.Channel),
		EventType:     extractEventType(msg.Channel),
		SchemaVersion: int(num(raw["schemaVersion"])),
		Payload:       msg.Payload,
		OccurredAt:    parseTime(str(raw["occurredAt"])),
	}

	if err := c.repo.Append(event); err != nil {
		log.Printf("Event Store: failed to append event: %v", err)
	}
}

func extractContext(channel string) string {
	switch {
	case strings.Contains(channel, "campaign"):
		return "campaign"
	case strings.Contains(channel, "catalog"), strings.Contains(channel, "customer"):
		return "catalog"
	case strings.Contains(channel, "redemption"):
		return "redemption"
	case strings.Contains(channel, "analytics"):
		return "analytics"
	default:
		return "unknown"
	}
}

func extractEventType(channel string) string {
	types := map[string]string{
		"promotionos.campaign.published":         "CampaignPublished",
		"promotionos.campaign.paused":            "CampaignPaused",
		"promotionos.catalog.item.excluded":      "CatalogItemExcluded",
		"promotionos.customer.segment.updated":   "SegmentUpdated",
		"promotionos.redemption.redeemed":        "OfferRedeemed",
		"promotionos.redemption.claim.submitted": "ClaimSubmitted",
		"promotionos.analytics.budget.exhausted": "BudgetExhausted",
		"promotionos.campaign.budget.updated":    "BudgetUpdated",
	}
	if t, ok := types[channel]; ok {
		return t
	}
	return "Unknown"
}

func extractAggregateType(channel string) string {
	switch extractContext(channel) {
	case "campaign":
		return "Campaign"
	case "catalog":
		return "CatalogItem"
	case "redemption":
		return "Redemption"
	case "analytics":
		return "CampaignMetrics"
	default:
		return "Unknown"
	}
}

func str(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func num(v interface{}) float64 {
	if v == nil {
		return 0
	}
	if n, ok := v.(float64); ok {
		return n
	}
	return 0
}

func parseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Now().UTC()
	}
	return t
}
