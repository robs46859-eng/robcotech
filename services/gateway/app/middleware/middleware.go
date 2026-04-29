package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/robs46859-eng/fullstackarkham/services/gateway/app/arkham"
	"github.com/robs46859-eng/fullstackarkham/services/gateway/app/auth"
	"github.com/robs46859-eng/fullstackarkham/services/gateway/app/observability"
)

// arkhamClient is the global Arkham security client
var arkhamClient *arkham.Client

// InitArkham initializes the Arkham security client
func InitArkham(baseURL, apiKey string) {
	arkhamClient = arkham.NewClient(baseURL, apiKey)
}

// ArkhamMiddleware integrates Arkham security detection
// 
// Flow:
// 1. Extract request features
// 2. Classify with Arkham (benign/probe/attack/scanner)
// 3. If benign → pass through
// 4. If probe → engage deception layer
// 5. If attack/scanner → block with creative response
func ArkhamMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip Arkham if not initialized (development mode)
		if arkhamClient == nil {
			next.ServeHTTP(w, r)
			return
		}

		// Get tenant ID from context
		tenantID, _ := r.Context().Value(auth.TenantIDKey).(string)
		if tenantID == "" {
			tenantID = "default"
		}

		// Extract request features for classification
		requestInfo := extractRequestFeatures(r)

		// Classify the request
		classification, err := arkhamClient.ClassifyRequest(r.Context(), tenantID, requestInfo)
		
		if err != nil {
			// Arkham unavailable - log and pass through (fail open)
			// In production, you might want to fail closed depending on security requirements
			next.ServeHTTP(w, r)
			return
		}

		// Handle based on classification
		switch classification.Classification {
		case "benign":
			// Pass through - add classification to context
			ctx := context.WithValue(r.Context(), ContextKey("arkham_classification"), classification)
			next.ServeHTTP(w, r.WithContext(ctx))

		case "probe":
			// Engage deception layer
			if classification.FingerprintHash != "" {
				deception, err := arkhamClient.GenerateDeception(
					r.Context(),
					tenantID,
					classification.FingerprintHash,
					requestInfo,
				)
				
				if err == nil {
					// Return deception payload
					w.Header().Set("Content-Type", "application/json")
					w.Header().Set("X-Arkham-Trap", "true")
					w.Header().Set("X-Engagement-ID", deception.EngagementID)
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(deception.DeceptionPayload)
					return
				}
			}
			// If deception fails, fall through to pass
			next.ServeHTTP(w, r)

		case "attack", "scanner":
			// Block with creative response
			if classification.FingerprintHash != "" {
				block, err := arkhamClient.ApplyBlock(
					r.Context(),
					tenantID,
					classification.FingerprintHash,
					"", // engagementID - would be set if coming from deception
				)
				
				if err == nil {
					// Return block response
					for key, value := range block.Headers {
						w.Header().Set(key, value)
					}
					w.Header().Set("Content-Type", "application/json")
					w.Header().Set("X-Arkham-Block", "true")
					w.WriteHeader(block.HTTPStatus)
					json.NewEncoder(w).Encode(block.Body)
					return
				}
			}
			// If block fails, return standard 403
			http.Error(w, "Forbidden", http.StatusForbidden)

		default:
			// Unknown classification - pass through
			next.ServeHTTP(w, r)
		}
	})
}

// extractRequestFeatures extracts features from HTTP request for Arkham classification
func extractRequestFeatures(r *http.Request) map[string]interface{} {
	features := map[string]interface{}{
		"source_ip":    r.RemoteAddr,
		"method":       r.Method,
		"path":         r.URL.Path,
		"user_agent":   r.UserAgent(),
		"content_type": r.Header.Get("Content-Type"),
		"timestamp":    time.Now().Unix(),
	}

	// Extract headers
	headers := make(map[string]string)
	for name, values := range r.Header {
		if len(values) > 0 {
			headers[name] = values[0]
		}
	}
	features["headers"] = headers

	// Content length
	if r.ContentLength > 0 {
		features["content_length"] = r.ContentLength
	}

	return features
}

// ContextKey is a custom type for context keys
type ContextKey string

const (
	ArkhamClassificationKey ContextKey = "arkham_classification"
)

// GetArkhamClassification retrieves Arkham classification from context
func GetArkhamClassification(ctx context.Context) *arkham.ThreatClassification {
	if classification, ok := ctx.Value(ArkhamClassificationKey).(*arkham.ThreatClassification); ok {
		return classification
	}
	return nil
}

// AuthMiddleware validates API keys and JWT tokens
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get tenant store from context (set during app initialization)
		storeRaw := r.Context().Value(ContextKey("tenant_store"))
		
		// For development, use in-memory store if no production store
		if storeRaw == nil {
			memStore := auth.NewInMemoryTenantStore()
			tenant, err := auth.ValidateAPIKey(r, memStore)
			if err == nil && tenant != nil {
				ctx := context.WithValue(r.Context(), auth.TenantIDKey, tenant.ID)
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
			return
		}
		
		// Type assert to interface
		store, ok := storeRaw.(auth.TenantStore)
		if !ok {
			// Fall back to in-memory
			memStore := auth.NewInMemoryTenantStore()
			tenant, _ := auth.ValidateAPIKey(r, memStore)
			if tenant != nil {
				ctx := context.WithValue(r.Context(), auth.TenantIDKey, tenant.ID)
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
			return
		}
		
		// Use production tenant store
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				apiKey = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}
		
		if apiKey == "" {
			http.Error(w, "Unauthorized: missing API key", http.StatusUnauthorized)
			return
		}
		
		tenant, err := store.GetByAPIKey(r.Context(), apiKey)
		if err != nil {
			http.Error(w, "Unauthorized: invalid API key", http.StatusUnauthorized)
			return
		}
		
		// Check quota
		hasQuota, err := store.CheckQuota(r.Context(), tenant.ID)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		
		if !hasQuota {
			http.Error(w, "Quota exceeded", http.StatusTooManyRequests)
			return
		}
		
		// Add tenant info to context
		ctx := context.WithValue(r.Context(), auth.TenantIDKey, tenant.ID)
		r = r.WithContext(ctx)
		
		next.ServeHTTP(w, r)
	})
}

// RateLimitMiddleware implements per-tenant rate limiting
func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement Redis-based rate limiting
		// For now, pass through
		next.ServeHTTP(w, r)
	})
}

// ObservabilityMiddleware adds logging, tracing, and metrics
func ObservabilityMiddleware(obs *observability.Observability) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Get or create trace
			ctx, span := obs.Tracer.Start(r.Context(), r.URL.Path)
			defer span.End()
			r = r.WithContext(ctx)

			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Call next handler
			next.ServeHTTP(wrapped, r)

			// Record metrics
			duration := time.Since(start)
			obs.RecordRequestDuration(r.URL.Path, r.Method, wrapped.statusCode, duration)
			obs.RecordRequestCount(r.URL.Path, r.Method, wrapped.statusCode)

			// Log request
			obs.Logger.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.statusCode,
				"duration_ms", duration.Milliseconds(),
			)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// TenantQuotaMiddleware checks and enforces tenant quotas
func TenantQuotaMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Check tenant quota from context
		// If exceeded, return 429 Too Many Requests

		next.ServeHTTP(w, r)
	})
}
