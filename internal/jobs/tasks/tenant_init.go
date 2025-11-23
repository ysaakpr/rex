package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/vyshakhp/utm-backend/internal/config"
	"github.com/vyshakhp/utm-backend/internal/models"
	"gorm.io/gorm"
)

type TenantInitHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewTenantInitHandler(db *gorm.DB, cfg *config.Config) *TenantInitHandler {
	return &TenantInitHandler{
		db:  db,
		cfg: cfg,
	}
}

type TenantInitPayload struct {
	TenantID string `json:"tenant_id"`
}

func (h *TenantInitHandler) HandleTenantInitialization(ctx context.Context, task *asynq.Task) error {
	var payload TenantInitPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	tenantID, err := uuid.Parse(payload.TenantID)
	if err != nil {
		return fmt.Errorf("invalid tenant ID: %w", err)
	}

	fmt.Printf("Starting tenant initialization for tenant: %s\n", tenantID)

	// Get tenant from database
	var tenant models.Tenant
	if err := h.db.Where("id = ?", tenantID).First(&tenant).Error; err != nil {
		return fmt.Errorf("failed to get tenant: %w", err)
	}

	// Check if already initialized
	if tenant.Status == models.TenantStatusActive {
		fmt.Printf("Tenant %s is already active\n", tenantID)
		return nil
	}

	// Initialize tenant in each backend service
	services := h.cfg.TenantInit.Services
	if len(services) == 0 {
		fmt.Println("No services configured for tenant initialization")
		// Still mark as active
		return h.markTenantActive(tenantID)
	}

	// Call each service sequentially for reliable initialization
	for _, serviceURL := range services {
		if err := h.initializeTenantInService(ctx, serviceURL, &tenant); err != nil {
			// Log error but continue with other services
			// In a production system, you might want to implement partial failure handling
			fmt.Printf("Failed to initialize tenant in service %s: %v\n", serviceURL, err)
			// Return error to trigger retry
			return fmt.Errorf("failed to initialize in service %s: %w", serviceURL, err)
		}
		fmt.Printf("Successfully initialized tenant in service: %s\n", serviceURL)
	}

	// Mark tenant as active
	return h.markTenantActive(tenantID)
}

func (h *TenantInitHandler) initializeTenantInService(ctx context.Context, serviceURL string, tenant *models.Tenant) error {
	// Prepare initialization request
	initData := map[string]interface{}{
		"tenant_id":   tenant.ID,
		"tenant_name": tenant.Name,
		"tenant_slug": tenant.Slug,
		"metadata":    tenant.Metadata,
		"created_at":  tenant.CreatedAt,
	}

	jsonData, err := json.Marshal(initData)
	if err != nil {
		return fmt.Errorf("failed to marshal init data: %w", err)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Construct the initialization endpoint
	// Assuming each service has a POST /api/v1/tenants/initialize endpoint
	url := fmt.Sprintf("%s/api/v1/tenants/initialize", serviceURL)

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add any authentication headers if needed
	// req.Header.Set("Authorization", "Bearer " + token)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("service returned error status %d: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("Service response: %s\n", string(jsonData))
	return nil
}

func (h *TenantInitHandler) markTenantActive(tenantID uuid.UUID) error {
	return h.db.Model(&models.Tenant{}).
		Where("id = ?", tenantID).
		Update("status", models.TenantStatusActive).Error
}
