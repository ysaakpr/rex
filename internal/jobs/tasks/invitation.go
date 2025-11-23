package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/smtp"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/vyshakhp/utm-backend/internal/config"
	"github.com/vyshakhp/utm-backend/internal/models"
	"gorm.io/gorm"
)

type InvitationHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewInvitationHandler(db *gorm.DB, cfg *config.Config) *InvitationHandler {
	return &InvitationHandler{
		db:  db,
		cfg: cfg,
	}
}

type InvitationPayload struct {
	InvitationID string `json:"invitation_id"`
}

func (h *InvitationHandler) HandleUserInvitation(ctx context.Context, task *asynq.Task) error {
	var payload InvitationPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	invitationID, err := uuid.Parse(payload.InvitationID)
	if err != nil {
		return fmt.Errorf("invalid invitation ID: %w", err)
	}

	fmt.Printf("Processing invitation email for invitation: %s\n", invitationID)

	// Get invitation from database
	var invitation models.UserInvitation
	if err := h.db.Preload("Tenant").Preload("Relation").
		Where("id = ?", invitationID).First(&invitation).Error; err != nil {
		return fmt.Errorf("failed to get invitation: %w", err)
	}

	// Check if invitation is still pending
	if invitation.Status != models.InvitationStatusPending {
		fmt.Printf("Invitation %s is not pending, skipping email\n", invitationID)
		return nil
	}

	// Send invitation email
	if err := h.sendInvitationEmail(&invitation); err != nil {
		return fmt.Errorf("failed to send invitation email: %w", err)
	}

	fmt.Printf("Successfully sent invitation email to: %s\n", invitation.Email)
	return nil
}

func (h *InvitationHandler) sendInvitationEmail(invitation *models.UserInvitation) error {
	// Build invitation URL
	invitationURL := fmt.Sprintf("%s?token=%s", h.cfg.Invitation.BaseURL, invitation.Token)

	// Prepare email content
	subject := fmt.Sprintf("You've been invited to join %s", invitation.Tenant.Name)
	body := fmt.Sprintf(`
Hello,

You have been invited to join %s on our platform.

Relation: %s

Please click the link below to accept your invitation:
%s

This invitation will expire on: %s

If you didn't expect this invitation, you can safely ignore this email.

Best regards,
The Team
	`, invitation.Tenant.Name, invitation.Relation.Name, invitationURL, invitation.ExpiresAt.Format("Jan 02, 2006 at 3:04 PM"))

	// Send email based on provider
	switch h.cfg.Email.Provider {
	case "smtp":
		return h.sendSMTPEmail(invitation.Email, subject, body)
	default:
		// For development, just log the email
		fmt.Printf("\n=== INVITATION EMAIL ===\n")
		fmt.Printf("To: %s\n", invitation.Email)
		fmt.Printf("Subject: %s\n", subject)
		fmt.Printf("Body:\n%s\n", body)
		fmt.Printf("========================\n\n")
		return nil
	}
}

func (h *InvitationHandler) sendSMTPEmail(to, subject, body string) error {
	from := h.cfg.Email.FromAddress
	smtpHost := h.cfg.Email.SMTPHost
	smtpPort := h.cfg.Email.SMTPPort
	smtpUser := h.cfg.Email.SMTPUser
	smtpPassword := h.cfg.Email.SMTPPassword

	// Compose message
	message := fmt.Sprintf("From: %s\r\n", from)
	message += fmt.Sprintf("To: %s\r\n", to)
	message += fmt.Sprintf("Subject: %s\r\n", subject)
	message += "\r\n"
	message += body

	// Setup authentication
	var auth smtp.Auth
	if smtpUser != "" && smtpPassword != "" {
		auth = smtp.PlainAuth("", smtpUser, smtpPassword, smtpHost)
	}

	// Send email
	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	err := smtp.SendMail(addr, auth, from, []string{to}, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send SMTP email: %w", err)
	}

	return nil
}
