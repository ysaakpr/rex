package repository

import (
	"github.com/google/uuid"
	"github.com/ysaakpr/rex/internal/models"
	"gorm.io/gorm"
)

type TenantRepository interface {
	Create(tenant *models.Tenant) error
	GetByID(id uuid.UUID) (*models.Tenant, error)
	GetBySlug(slug string) (*models.Tenant, error)
	GetByCreatorID(creatorID string, pagination *models.PaginationParams) ([]*models.Tenant, int64, error)
	List(pagination *models.PaginationParams) ([]*models.Tenant, int64, error)
	Update(tenant *models.Tenant) error
	Delete(id uuid.UUID) error
	UpdateStatus(id uuid.UUID, status models.TenantStatus) error
}

type tenantRepository struct {
	db *gorm.DB
}

func NewTenantRepository(db *gorm.DB) TenantRepository {
	return &tenantRepository{db: db}
}

func (r *tenantRepository) Create(tenant *models.Tenant) error {
	return r.db.Create(tenant).Error
}

func (r *tenantRepository) GetByID(id uuid.UUID) (*models.Tenant, error) {
	var tenant models.Tenant
	err := r.db.Where("id = ?", id).First(&tenant).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *tenantRepository) GetBySlug(slug string) (*models.Tenant, error) {
	var tenant models.Tenant
	err := r.db.Where("slug = ?", slug).First(&tenant).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *tenantRepository) GetByCreatorID(creatorID string, pagination *models.PaginationParams) ([]*models.Tenant, int64, error) {
	var tenants []*models.Tenant
	var total int64

	query := r.db.Model(&models.Tenant{}).Where("created_by = ?", creatorID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	pagination.Normalize()
	err := query.
		Offset(pagination.GetOffset()).
		Limit(pagination.PageSize).
		Order("created_at DESC").
		Find(&tenants).Error

	return tenants, total, err
}

func (r *tenantRepository) List(pagination *models.PaginationParams) ([]*models.Tenant, int64, error) {
	var tenants []*models.Tenant
	var total int64

	query := r.db.Model(&models.Tenant{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	pagination.Normalize()
	err := query.
		Offset(pagination.GetOffset()).
		Limit(pagination.PageSize).
		Order("created_at DESC").
		Find(&tenants).Error

	return tenants, total, err
}

func (r *tenantRepository) Update(tenant *models.Tenant) error {
	return r.db.Save(tenant).Error
}

func (r *tenantRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Tenant{}, id).Error
}

func (r *tenantRepository) UpdateStatus(id uuid.UUID, status models.TenantStatus) error {
	return r.db.Model(&models.Tenant{}).
		Where("id = ?", id).
		Update("status", status).Error
}
