package repository

import (
	"github.com/google/uuid"
	"github.com/ysaakpr/rex/internal/models"
	"gorm.io/gorm"
)

type MemberRepository interface {
	Create(member *models.TenantMember) error
	GetByID(id uuid.UUID) (*models.TenantMember, error)
	GetByTenantAndUser(tenantID uuid.UUID, userID string) (*models.TenantMember, error)
	GetByTenantID(tenantID uuid.UUID, pagination *models.PaginationParams) ([]*models.TenantMember, int64, error)
	GetByUserID(userID string) ([]*models.TenantMember, error)
	Update(member *models.TenantMember) error
	Delete(id uuid.UUID) error
	AssignRoles(memberID uuid.UUID, roleIDs []uuid.UUID) error
	RemoveRole(memberID uuid.UUID, roleID uuid.UUID) error
	GetMemberWithRoles(memberID uuid.UUID) (*models.TenantMember, error)
}

type memberRepository struct {
	db *gorm.DB
}

func NewMemberRepository(db *gorm.DB) MemberRepository {
	return &memberRepository{db: db}
}

func (r *memberRepository) Create(member *models.TenantMember) error {
	return r.db.Create(member).Error
}

func (r *memberRepository) GetByID(id uuid.UUID) (*models.TenantMember, error) {
	var member models.TenantMember
	err := r.db.
		Preload("Role").
		Where("id = ?", id).
		First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *memberRepository) GetByTenantAndUser(tenantID uuid.UUID, userID string) (*models.TenantMember, error) {
	var member models.TenantMember
	err := r.db.
		Preload("Role").
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *memberRepository) GetByTenantID(tenantID uuid.UUID, pagination *models.PaginationParams) ([]*models.TenantMember, int64, error) {
	var members []*models.TenantMember
	var total int64

	query := r.db.Model(&models.TenantMember{}).Where("tenant_id = ?", tenantID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	pagination.Normalize()
	err := query.
		Preload("Role").
		Offset(pagination.GetOffset()).
		Limit(pagination.PageSize).
		Order("created_at DESC").
		Find(&members).Error

	return members, total, err
}

func (r *memberRepository) GetByUserID(userID string) ([]*models.TenantMember, error) {
	var members []*models.TenantMember
	err := r.db.
		Preload("Tenant").
		Preload("Role").
		Where("user_id = ? AND status = ?", userID, models.MemberStatusActive).
		Find(&members).Error
	return members, err
}

func (r *memberRepository) Update(member *models.TenantMember) error {
	return r.db.Save(member).Error
}

func (r *memberRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.TenantMember{}, id).Error
}

func (r *memberRepository) AssignRoles(memberID uuid.UUID, roleIDs []uuid.UUID) error {
	member := &models.TenantMember{ID: memberID}
	return r.db.Model(member).Association("Roles").Append(convertToRoles(roleIDs))
}

func (r *memberRepository) RemoveRole(memberID uuid.UUID, roleID uuid.UUID) error {
	member := &models.TenantMember{ID: memberID}
	role := &models.Role{ID: roleID}
	return r.db.Model(member).Association("Roles").Delete(role)
}

func (r *memberRepository) GetMemberWithRoles(memberID uuid.UUID) (*models.TenantMember, error) {
	var member models.TenantMember
	err := r.db.
		Preload("Role").
		Preload("Role.Policies").
		Preload("Role.Policies.Permissions").
		Where("id = ?", memberID).
		First(&member).Error
	return &member, err
}

func convertToRoles(roleIDs []uuid.UUID) []*models.Role {
	roles := make([]*models.Role, len(roleIDs))
	for i, id := range roleIDs {
		roles[i] = &models.Role{ID: id}
	}
	return roles
}
