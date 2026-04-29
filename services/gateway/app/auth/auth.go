package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	TenantIDKey contextKey = "tenant_id"
	UserIDKey   contextKey = "user_id"
	APIKeyKey   contextKey = "api_key"
)

// Tenant represents a multi-tenant organization
type Tenant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	APIKey    string    `json:"-"` // Never serialize
	Plan      string    `json:"plan"`
	Quota     int64     `json:"quota"`
	CreatedAt time.Time `json:"created_at"`
}

// Claims represents JWT claims for authenticated requests
type Claims struct {
	TenantID string `json:"tenant_id"`
	UserID   string `json:"user_id"`
	jwt.RegisteredClaims
}

// ErrUnauthorized is returned when authentication fails
var ErrUnauthorized = errors.New("unauthorized")

// ErrInvalidAPIKey is returned when API key is invalid
var ErrInvalidAPIKey = errors.New("invalid API key")

// ErrQuotaExceeded is returned when tenant quota is exceeded
var ErrQuotaExceeded = errors.New("quota exceeded")

// TenantStore defines the interface for tenant storage
type TenantStore interface {
	GetByAPIKey(ctx context.Context, apiKey string) (*Tenant, error)
	GetByID(ctx context.Context, id string) (*Tenant, error)
	CheckQuota(ctx context.Context, tenantID string) (bool, error)
	RecordUsage(ctx context.Context, tenantID string, tokens int64) error
}

// InMemoryTenantStore is a simple in-memory implementation for development
type InMemoryTenantStore struct {
	tenants map[string]*Tenant
	keys    map[string]*Tenant
}

// NewInMemoryTenantStore creates a new in-memory tenant store
func NewInMemoryTenantStore() *InMemoryTenantStore {
	return &InMemoryTenantStore{
		tenants: make(map[string]*Tenant),
		keys:    make(map[string]*Tenant),
	}
}

// GetByAPIKey retrieves a tenant by API key
func (s *InMemoryTenantStore) GetByAPIKey(ctx context.Context, apiKey string) (*Tenant, error) {
	tenant, ok := s.keys[apiKey]
	if !ok {
		return nil, ErrInvalidAPIKey
	}
	return tenant, nil
}

// GetByID retrieves a tenant by ID
func (s *InMemoryTenantStore) GetByID(ctx context.Context, id string) (*Tenant, error) {
	tenant, ok := s.tenants[id]
	if !ok {
		return nil, ErrUnauthorized
	}
	return tenant, nil
}

// CheckQuota checks if tenant has remaining quota
func (s *InMemoryTenantStore) CheckQuota(ctx context.Context, tenantID string) (bool, error) {
	// Simplified - always allow for now
	return true, nil
}

// RecordUsage records token usage for a tenant
func (s *InMemoryTenantStore) RecordUsage(ctx context.Context, tenantID string, tokens int64) error {
	// Simplified - no-op for now
	return nil
}

// ValidateAPIKey validates an API key from the request header
func ValidateAPIKey(r *http.Request, store TenantStore) (*Tenant, error) {
	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		// Try Authorization header
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			apiKey = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			return nil, ErrUnauthorized
		}
	}

	if apiKey == "" {
		return nil, ErrUnauthorized
	}

	tenant, err := store.GetByAPIKey(r.Context(), apiKey)
	if err != nil {
		return nil, err
	}

	return tenant, nil
}

// ValidateJWT validates a JWT token
func ValidateJWT(tokenString string, secret []byte) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrUnauthorized
		}
		return secret, nil
	})

	if err != nil {
		return nil, ErrUnauthorized
	}

	if !token.Valid {
		return nil, ErrUnauthorized
	}

	return claims, nil
}
