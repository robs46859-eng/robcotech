package database

import (
	"context"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Tenant represents a multi-tenant organization
type Tenant struct {
	ID                  uuid.UUID `gorm:"type:uuid;primary_key"`
	Name                string    `gorm:"size:255;not null"`
	Slug                string    `gorm:"size:100;uniqueIndex;not null"`
	APIKeyHash          string    `gorm:"size:255;not null;index"`
	APIKeyPrefix        string    `gorm:"size:20;not null"`
	Plan                string    `gorm:"size:50;not null;default:'free'"`
	Status              string    `gorm:"size:50;not null;default:'active'"`
	QuotaMonthly        int64     `gorm:"not null;default:10000"`
	QuotaUsed           int64     `gorm:"not null;default:0"`
	StripeCustomerID    string    `gorm:"size:100"`
	StripeSubscriptionID string   `gorm:"size:100"`
	Settings            string    `gorm:"type:jsonb;default:'{}'"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// APIKey represents an API key for tenant authentication
type APIKey struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key"`
	TenantID     uuid.UUID      `gorm:"type:uuid;not null;index"`
	Name         string         `gorm:"size:255;not null"`
	KeyHash      string         `gorm:"size:255;not null;index"`
	KeyPrefix    string         `gorm:"size:20;not null;index"`
	Permissions  string         `gorm:"type:jsonb;default:'[\"infer\", \"read\"]'"`
	ExpiresAt    *time.Time
	LastUsedAt   *time.Time
	CreatedAt    time.Time
	RevokedAt    *time.Time
}

// TenantStore provides database-backed tenant operations
type TenantStore struct {
	DB *gorm.DB
}

// NewTenantStore creates a new tenant store
func NewTenantStore(db *gorm.DB) *TenantStore {
	return &TenantStore{DB: db}
}

// CreateTenant creates a new tenant
func (s *TenantStore) CreateTenant(ctx context.Context, tenant *Tenant) error {
	tenant.ID = uuid.New()
	return s.DB.WithContext(ctx).Create(tenant).Error
}

// GetByAPIKey retrieves a tenant by API key
func (s *TenantStore) GetByAPIKey(ctx context.Context, apiKey string) (*Tenant, error) {
	// Find all API keys with matching prefix
	var apiKeys []APIKey
	if err := s.DB.WithContext(ctx).
		Where("key_prefix = ?", extractKeyPrefix(apiKey)).
		Find(&apiKeys).Error; err != nil {
		return nil, err
	}

	// Check each key
	for _, key := range apiKeys {
		if err := bcrypt.CompareHashAndPassword([]byte(key.KeyHash), []byte(apiKey)); err == nil {
			// Found matching key, get tenant
			var tenant Tenant
			if err := s.DB.WithContext(ctx).
				Where("id = ?", key.TenantID).
				First(&tenant).Error; err != nil {
				return nil, err
			}

			// Update last used
			now := time.Now()
			key.LastUsedAt = &now
			s.DB.WithContext(ctx).Save(&key)

			return &tenant, nil
		}
	}

	return nil, ErrInvalidAPIKey
}

// GetByID retrieves a tenant by ID
func (s *TenantStore) GetByID(ctx context.Context, id uuid.UUID) (*Tenant, error) {
	var tenant Tenant
	err := s.DB.WithContext(ctx).First(&tenant, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, ErrTenantNotFound
	}
	return &tenant, err
}

// CheckQuota checks if tenant has remaining quota
func (s *TenantStore) CheckQuota(ctx context.Context, tenantID uuid.UUID) (bool, int64, error) {
	var tenant Tenant
	if err := s.DB.WithContext(ctx).First(&tenant, "id = ?", tenantID).Error; err != nil {
		return false, 0, err
	}

	remaining := tenant.QuotaMonthly - tenant.QuotaUsed
	return remaining > 0, remaining, nil
}

// RecordUsage records token usage for a tenant
func (s *TenantStore) RecordUsage(ctx context.Context, tenantID uuid.UUID, tokens int64) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Exec(
			"UPDATE tenants SET quota_used = quota_used + ?, updated_at = ? WHERE id = ?",
			tokens, time.Now(), tenantID,
		)
		return result.Error
	})
}

// CreateAPIKey creates a new API key for a tenant
func (s *TenantStore) CreateAPIKey(ctx context.Context, tenantID uuid.UUID, name string) (string, *APIKey, error) {
	// Generate API key
	apiKey := "fsa_" + uuid.New().String()[:24]

	// Hash the key
	hashedKey, err := bcrypt.GenerateFromPassword([]byte(apiKey), bcrypt.DefaultCost)
	if err != nil {
		return "", nil, err
	}

	// Create API key record
	apiKeyRecord := &APIKey{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Name:      name,
		KeyHash:   string(hashedKey),
		KeyPrefix: "fsa_" + apiKey[:8],
	}

	if err := s.DB.WithContext(ctx).Create(apiKeyRecord).Error; err != nil {
		return "", nil, err
	}

	return apiKey, apiKeyRecord, nil
}

// RevokeAPIKey revokes an API key
func (s *TenantStore) RevokeAPIKey(ctx context.Context, keyID uuid.UUID) error {
	now := time.Now()
	return s.DB.WithContext(ctx).
		Model(&APIKey{}).
		Where("id = ?", keyID).
		Update("revoked_at", now).Error
}

// ListAPIKeys lists all API keys for a tenant
func (s *TenantStore) ListAPIKeys(ctx context.Context, tenantID uuid.UUID) ([]APIKey, error) {
	var keys []APIKey
	err := s.DB.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Find(&keys).Error
	return keys, err
}

// Helper functions
func extractKeyPrefix(apiKey string) string {
	if len(apiKey) < 8 {
		return apiKey
	}
	return apiKey[:8]
}

// Errors
var (
	ErrInvalidAPIKey = &APIError{Message: "invalid API key"}
	ErrTenantNotFound = &APIError{Message: "tenant not found"}
)

// APIError represents an API error
type APIError struct {
	Message string
}

func (e *APIError) Error() string {
	return e.Message
}
