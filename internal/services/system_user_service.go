package services

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/supertokens/supertokens-golang/recipe/emailpassword"
	"github.com/supertokens/supertokens-golang/recipe/session"
	"github.com/supertokens/supertokens-golang/recipe/usermetadata"
	"gorm.io/gorm"

	"github.com/ysaakpr/rex/internal/models"
	"github.com/ysaakpr/rex/internal/repository"
)

type SystemUserService interface {
	CreateSystemUser(input *models.CreateSystemUserInput, createdBy string) (*models.SystemUserCreateResponse, error)
	GetSystemUser(id uuid.UUID) (*models.SystemUser, error)
	GetSystemUserByUserID(userID string) (*models.SystemUser, error)
	GetSystemUsersByApplication(applicationName string) ([]*models.SystemUser, error)
	ListSystemUsers(activeOnly bool) ([]*models.SystemUser, error)
	UpdateSystemUser(id uuid.UUID, input *models.UpdateSystemUserInput) (*models.SystemUser, error)
	RegeneratePassword(id uuid.UUID) (string, error)
	RotateWithGracePeriod(id uuid.UUID, gracePeriodDays int) (*models.SystemUserCreateResponse, error)
	RevokeNonPrimary(applicationName string) (int, error)
	DeactivateExpired() (int64, error)
	DeactivateSystemUser(id uuid.UUID) error
	UpdateLastUsed(userID string) error
}

type systemUserService struct {
	repo repository.SystemUserRepository
}

func NewSystemUserService(repo repository.SystemUserRepository) SystemUserService {
	return &systemUserService{repo: repo}
}

// generatePassword creates a cryptographically secure random password
func generatePassword() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "sysuser_" + base64.URLEncoding.EncodeToString(bytes)[:40], nil
}

// generateEmail creates a system email address
func generateEmail(name string) string {
	return fmt.Sprintf("%s@system.internal", name)
}

func (s *systemUserService) CreateSystemUser(input *models.CreateSystemUserInput, createdBy string) (*models.SystemUserCreateResponse, error) {
	// Check if name already exists
	existing, _ := s.repo.GetByName(input.Name)
	if existing != nil {
		return nil, errors.New("system user with this name already exists")
	}

	// Generate email and password
	email := generateEmail(input.Name)
	password, err := generatePassword()
	if err != nil {
		return nil, fmt.Errorf("failed to generate password: %w", err)
	}

	// Create user in SuperTokens
	signUpResponse, err := emailpassword.SignUp("public", email, password)
	if err != nil {
		return nil, fmt.Errorf("failed to create user in SuperTokens: %w", err)
	}

	// Check if signup was successful
	if signUpResponse.OK == nil {
		if signUpResponse.EmailAlreadyExistsError != nil {
			return nil, errors.New("email already exists in SuperTokens")
		}
		return nil, errors.New("failed to create user in SuperTokens")
	}

	supertokensUserID := signUpResponse.OK.User.ID

	// Set SuperTokens metadata
	metadata := map[string]interface{}{
		"is_system_user": true,
		"service_name":   input.Name,
		"service_type":   input.ServiceType,
		"created_by":     createdBy,
	}

	_, err = usermetadata.UpdateUserMetadata(supertokensUserID, metadata)
	if err != nil {
		// Rollback: Try to delete the SuperTokens user
		// Note: SuperTokens doesn't have a direct delete method in Go SDK
		// You might need to handle this through SuperTokens dashboard or API
		return nil, fmt.Errorf("failed to set SuperTokens metadata: %w", err)
	}

	// Create system user in database
	systemUser := &models.SystemUser{
		Name:            input.Name,
		ApplicationName: input.Name, // Initially, application name = name
		Email:           email,
		UserID:          supertokensUserID,
		Description:     input.Description,
		ServiceType:     input.ServiceType,
		IsActive:        true,
		IsPrimary:       true, // First credential is always primary
		ExpiresAt:       nil,  // Initial credential never expires
		CreatedBy:       createdBy,
		Metadata:        input.Metadata,
	}

	if err := s.repo.Create(systemUser); err != nil {
		// Rollback SuperTokens user creation would happen here
		return nil, fmt.Errorf("failed to create system user in database: %w", err)
	}

	// Return response with password (only shown once)
	return &models.SystemUserCreateResponse{
		ID:              systemUser.ID,
		Name:            systemUser.Name,
		ApplicationName: systemUser.ApplicationName,
		Email:           systemUser.Email,
		Password:        password,
		UserID:          systemUser.UserID,
		Description:     systemUser.Description,
		ServiceType:     systemUser.ServiceType,
		IsActive:        systemUser.IsActive,
		IsPrimary:       systemUser.IsPrimary,
		ExpiresAt:       systemUser.ExpiresAt,
		CreatedBy:       systemUser.CreatedBy,
		Metadata:        input.Metadata,
		CreatedAt:       systemUser.CreatedAt,
		Message:         "Save this password securely. It will not be shown again.",
	}, nil
}

func (s *systemUserService) GetSystemUser(id uuid.UUID) (*models.SystemUser, error) {
	return s.repo.GetByID(id)
}

func (s *systemUserService) GetSystemUserByUserID(userID string) (*models.SystemUser, error) {
	return s.repo.GetByUserID(userID)
}

func (s *systemUserService) ListSystemUsers(activeOnly bool) ([]*models.SystemUser, error) {
	return s.repo.List(activeOnly)
}

func (s *systemUserService) UpdateSystemUser(id uuid.UUID, input *models.UpdateSystemUserInput) (*models.SystemUser, error) {
	systemUser, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if input.Description != nil {
		systemUser.Description = *input.Description
	}
	if input.IsActive != nil {
		systemUser.IsActive = *input.IsActive
	}
	if input.Metadata != nil {
		systemUser.Metadata = input.Metadata
	}

	if err := s.repo.Update(systemUser); err != nil {
		return nil, err
	}

	return systemUser, nil
}

func (s *systemUserService) RegeneratePassword(id uuid.UUID) (string, error) {
	systemUser, err := s.repo.GetByID(id)
	if err != nil {
		return "", err
	}

	// Generate new password
	newPassword, err := generatePassword()
	if err != nil {
		return "", fmt.Errorf("failed to generate password: %w", err)
	}

	// Update password in SuperTokens
	// Using the correct function signature for UpdateEmailOrPassword
	updateResponse, err := emailpassword.UpdateEmailOrPassword(
		systemUser.UserID,
		nil,          // email (not changing)
		&newPassword, // new password
		nil,          // applyPasswordPolicy - set to nil to use default
		nil,          // tenantIdForPasswordPolicy
	)
	if err != nil {
		return "", fmt.Errorf("failed to update password in SuperTokens: %w", err)
	}

	if updateResponse.OK == nil {
		// Check for other error types
		if updateResponse.EmailAlreadyExistsError != nil {
			return "", errors.New("email already exists")
		}
		if updateResponse.UnknownUserIdError != nil {
			return "", errors.New("unknown user ID")
		}
		return "", errors.New("failed to update password in SuperTokens")
	}

	// Revoke all existing sessions for this user
	// This ensures old tokens become invalid
	_, err = session.RevokeAllSessionsForUser(systemUser.UserID, nil, nil)
	if err != nil {
		// Log error but don't fail the request
		// The password is still updated
	}

	return newPassword, nil
}

func (s *systemUserService) DeactivateSystemUser(id uuid.UUID) error {
	systemUser, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// Deactivate in database
	systemUser.IsActive = false
	if err := s.repo.Update(systemUser); err != nil {
		return err
	}

	// Revoke all sessions
	_, err = session.RevokeAllSessionsForUser(systemUser.UserID, nil, nil)
	if err != nil {
		// Log error but don't fail
		// The user is already deactivated in the database
	}

	return nil
}

func (s *systemUserService) UpdateLastUsed(userID string) error {
	systemUser, err := s.repo.GetByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Not a system user, ignore
			return nil
		}
		return err
	}

	return s.repo.UpdateLastUsedAt(systemUser.ID)
}

func (s *systemUserService) GetSystemUsersByApplication(applicationName string) ([]*models.SystemUser, error) {
	return s.repo.GetByApplicationName(applicationName)
}

// RotateWithGracePeriod creates a new credential while keeping the old one active for a grace period
func (s *systemUserService) RotateWithGracePeriod(id uuid.UUID, gracePeriodDays int) (*models.SystemUserCreateResponse, error) {
	// Get the current system user
	currentUser, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get system user: %w", err)
	}

	// Default grace period
	if gracePeriodDays <= 0 {
		gracePeriodDays = 7
	}

	// Calculate expiry time
	expiresAt := time.Now().Add(time.Duration(gracePeriodDays) * 24 * time.Hour)

	// Step 1: Mark current user as non-primary and set expiry
	currentUser.IsPrimary = false
	currentUser.ExpiresAt = &expiresAt
	if err := s.repo.Update(currentUser); err != nil {
		return nil, fmt.Errorf("failed to update current user: %w", err)
	}

	// Step 2: Get all system users for this application to generate next version number
	allCreds, err := s.repo.GetByApplicationName(currentUser.ApplicationName)
	if err != nil {
		return nil, fmt.Errorf("failed to get application credentials: %w", err)
	}

	// Generate next version number
	nextVersion := len(allCreds) + 1
	newName := fmt.Sprintf("%s-v%d", currentUser.ApplicationName, nextVersion)
	newEmail := generateEmail(newName)

	// Step 3: Generate new password
	newPassword, err := generatePassword()
	if err != nil {
		return nil, fmt.Errorf("failed to generate password: %w", err)
	}

	// Step 4: Create new SuperTokens user
	signUpResponse, err := emailpassword.SignUp("public", newEmail, newPassword)
	if err != nil {
		// Rollback: restore current user as primary
		currentUser.IsPrimary = true
		currentUser.ExpiresAt = nil
		s.repo.Update(currentUser)
		return nil, fmt.Errorf("failed to create user in SuperTokens: %w", err)
	}

	if signUpResponse.OK == nil {
		// Rollback
		currentUser.IsPrimary = true
		currentUser.ExpiresAt = nil
		s.repo.Update(currentUser)

		if signUpResponse.EmailAlreadyExistsError != nil {
			return nil, errors.New("email already exists in SuperTokens")
		}
		return nil, errors.New("failed to create user in SuperTokens")
	}

	newUserID := signUpResponse.OK.User.ID

	// Step 5: Set SuperTokens metadata for new user
	metadata := map[string]interface{}{
		"is_system_user": true,
		"service_name":   currentUser.ApplicationName,
		"service_type":   currentUser.ServiceType,
		"created_by":     currentUser.CreatedBy,
		"version":        nextVersion,
	}

	_, err = usermetadata.UpdateUserMetadata(newUserID, metadata)
	if err != nil {
		// Log error but continue - metadata is not critical
	}

	// Step 6: Create new system user record
	newSystemUser := &models.SystemUser{
		Name:            newName,
		ApplicationName: currentUser.ApplicationName,
		Email:           newEmail,
		UserID:          newUserID,
		Description:     currentUser.Description,
		ServiceType:     currentUser.ServiceType,
		IsActive:        true,
		IsPrimary:       true,
		ExpiresAt:       nil,
		CreatedBy:       currentUser.CreatedBy,
		Metadata:        currentUser.Metadata,
	}

	if err := s.repo.Create(newSystemUser); err != nil {
		return nil, fmt.Errorf("failed to create new system user: %w", err)
	}

	// Step 7: Build response with old credentials info
	oldCredentials := []models.OldCredentialInfo{
		{
			Email:     currentUser.Email,
			ExpiresAt: &expiresAt,
			Message:   fmt.Sprintf("This credential will stop working on %s", expiresAt.Format("2006-01-02 15:04:05")),
		},
	}

	return &models.SystemUserCreateResponse{
		ID:              newSystemUser.ID,
		Name:            newSystemUser.Name,
		ApplicationName: newSystemUser.ApplicationName,
		Email:           newSystemUser.Email,
		Password:        newPassword,
		UserID:          newSystemUser.UserID,
		Description:     newSystemUser.Description,
		ServiceType:     newSystemUser.ServiceType,
		IsActive:        newSystemUser.IsActive,
		IsPrimary:       newSystemUser.IsPrimary,
		ExpiresAt:       newSystemUser.ExpiresAt,
		CreatedBy:       newSystemUser.CreatedBy,
		Metadata:        currentUser.Metadata,
		CreatedAt:       newSystemUser.CreatedAt,
		Message:         fmt.Sprintf("New credential created. Old credentials will expire on %s. Both work during grace period (%d days).", expiresAt.Format("2006-01-02"), gracePeriodDays),
		OldCredentials:  oldCredentials,
	}, nil
}

// RevokeNonPrimary immediately deactivates all non-primary credentials for an application
func (s *systemUserService) RevokeNonPrimary(applicationName string) (int, error) {
	// Get all credentials for the application
	credentials, err := s.repo.GetByApplicationName(applicationName)
	if err != nil {
		return 0, fmt.Errorf("failed to get application credentials: %w", err)
	}

	revokedCount := 0
	for _, cred := range credentials {
		if !cred.IsPrimary && cred.IsActive {
			// Deactivate
			cred.IsActive = false
			if err := s.repo.Update(cred); err != nil {
				// Log error but continue
				continue
			}

			// Revoke sessions
			_, err = session.RevokeAllSessionsForUser(cred.UserID, nil, nil)
			if err != nil {
				// Log error but continue
			}

			revokedCount++
		}
	}

	return revokedCount, nil
}

// DeactivateExpired deactivates all expired credentials (background job)
func (s *systemUserService) DeactivateExpired() (int64, error) {
	// Get all expired credentials
	count, err := s.repo.DeactivateExpired()
	if err != nil {
		return 0, fmt.Errorf("failed to deactivate expired credentials: %w", err)
	}

	// Note: We don't revoke sessions here because it could be expensive
	// Sessions will naturally expire based on SuperTokens configuration
	// If immediate revocation is needed, use RevokeNonPrimary manually

	return count, nil
}
