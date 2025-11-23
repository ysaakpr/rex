package jobs

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const (
	TypeTenantInitialization = "tenant:initialize"
	TypeUserInvitation       = "user:invitation"
	TypeSystemUserExpiry     = "system_user:expiry"

	QueueCritical = "critical"
	QueueDefault  = "default"
	QueueLow      = "low"
)

type Client interface {
	EnqueueTenantInitialization(tenantID uuid.UUID) error
	EnqueueUserInvitation(invitationID uuid.UUID) error
	Close() error
}

type client struct {
	asynqClient *asynq.Client
}

func NewClient(redisAddr string, redisPassword string) (Client, error) {
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: redisPassword,
	})

	return &client{
		asynqClient: asynqClient,
	}, nil
}

func (c *client) EnqueueTenantInitialization(tenantID uuid.UUID) error {
	payload, err := json.Marshal(map[string]interface{}{
		"tenant_id": tenantID.String(),
	})
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TypeTenantInitialization, payload)

	info, err := c.asynqClient.Enqueue(
		task,
		asynq.Queue(QueueCritical),
		asynq.MaxRetry(5),
	)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	fmt.Printf("Enqueued tenant initialization task: id=%s, queue=%s\n", info.ID, info.Queue)
	return nil
}

func (c *client) EnqueueUserInvitation(invitationID uuid.UUID) error {
	payload, err := json.Marshal(map[string]interface{}{
		"invitation_id": invitationID.String(),
	})
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TypeUserInvitation, payload)

	info, err := c.asynqClient.Enqueue(
		task,
		asynq.Queue(QueueDefault),
		asynq.MaxRetry(3),
	)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	fmt.Printf("Enqueued user invitation task: id=%s, queue=%s\n", info.ID, info.Queue)
	return nil
}

func (c *client) Close() error {
	return c.asynqClient.Close()
}
