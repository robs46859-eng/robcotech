// Package api implements the Papabase HTTP API handlers
package api

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/robs46859-eng/fullstackarkham/services/papabase/app/agents"
)

// ContextKey is a custom type for context keys
type ContextKey string

const (
	TenantIDKey ContextKey = "tenant_id"
	UserIDKey   ContextKey = "user_id"
	APIKeyKey   ContextKey = "api_key"
)

// CRMStoreKey and other API context keys
const (
	CRMStoreKey ContextKey = "crm_store"
	DadAIKey    ContextKey = "dad_ai"
	GatewayURLKey ContextKey = "gateway_url"
)

// Note: Lead and Task types are now defined in agents/types.go for sharing

// ============================================================================
// In-Memory Store (for MVP - replace with PostgreSQL in production)
// ============================================================================

// CRMStore defines the interface for CRM storage
type CRMStore interface {
	CreateLead(ctx context.Context, lead *agents.Lead) error
	GetLead(ctx context.Context, id string) (*agents.Lead, error)
	ListLeads(ctx context.Context, tenantID string) ([]*agents.Lead, error)
	UpdateLead(ctx context.Context, lead *agents.Lead) error
	DeleteLead(ctx context.Context, id string) error

	CreateTask(ctx context.Context, task *agents.Task) error
	GetTask(ctx context.Context, id string) (*agents.Task, error)
	ListTasks(ctx context.Context, tenantID string) ([]*agents.Task, error)
	UpdateTask(ctx context.Context, task *agents.Task) error
	DeleteTask(ctx context.Context, id string) error
}

// InMemoryCRMStore is an in-memory implementation for development
type InMemoryCRMStore struct {
	mu      sync.RWMutex
	leads   map[string]*agents.Lead
	tasks   map[string]*agents.Task
	tenants map[string]bool
}

// NewInMemoryCRMStore creates a new in-memory CRM store
func NewInMemoryCRMStore() *InMemoryCRMStore {
	return &InMemoryCRMStore{
		leads:   make(map[string]*agents.Lead),
		tasks:   make(map[string]*agents.Task),
		tenants: make(map[string]bool),
	}
}

// CreateLead creates a new lead
func (s *InMemoryCRMStore) CreateLead(ctx context.Context, lead *agents.Lead) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	lead.CreatedAt = time.Now()
	lead.UpdatedAt = time.Now()
	s.leads[lead.ID] = lead
	return nil
}

// GetLead retrieves a lead by ID
func (s *InMemoryCRMStore) GetLead(ctx context.Context, id string) (*agents.Lead, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	lead, ok := s.leads[id]
	if !ok {
		return nil, nil
	}
	return lead, nil
}

// ListLeads lists all leads for a tenant
func (s *InMemoryCRMStore) ListLeads(ctx context.Context, tenantID string) ([]*agents.Lead, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var leads []*agents.Lead
	for _, lead := range s.leads {
		if lead.TenantID == tenantID || tenantID == "" {
			leads = append(leads, lead)
		}
	}
	return leads, nil
}

// UpdateLead updates an existing lead
func (s *InMemoryCRMStore) UpdateLead(ctx context.Context, lead *agents.Lead) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	lead.UpdatedAt = time.Now()
	s.leads[lead.ID] = lead
	return nil
}

// DeleteLead deletes a lead
func (s *InMemoryCRMStore) DeleteLead(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.leads, id)
	return nil
}

// CreateTask creates a new task
func (s *InMemoryCRMStore) CreateTask(ctx context.Context, task *agents.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	s.tasks[task.ID] = task
	return nil
}

// GetTask retrieves a task by ID
func (s *InMemoryCRMStore) GetTask(ctx context.Context, id string) (*agents.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, ok := s.tasks[id]
	if !ok {
		return nil, nil
	}
	return task, nil
}

// ListTasks lists all tasks for a tenant
func (s *InMemoryCRMStore) ListTasks(ctx context.Context, tenantID string) ([]*agents.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var tasks []*agents.Task
	for _, task := range s.tasks {
		if task.TenantID == tenantID || tenantID == "" {
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}

// UpdateTask updates an existing task
func (s *InMemoryCRMStore) UpdateTask(ctx context.Context, task *agents.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task.UpdatedAt = time.Now()
	s.tasks[task.ID] = task
	return nil
}

// DeleteTask deletes a task
func (s *InMemoryCRMStore) DeleteTask(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.tasks, id)
	return nil
}

// ============================================================================
// Health Endpoints
// ============================================================================

// HealthHandler handles GET /health
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":  "healthy",
		"service": "papabase",
		"version": version,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ReadyHandler handles GET /ready
func ReadyHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status": "ready",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

var version = "0.1.0"

// ============================================================================
// Lead Endpoints
// ============================================================================

// CreateLeadRequest is the request body for creating a lead
type CreateLeadRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Company  string `json:"company"`
	Source   string `json:"source"`
	Notes    string `json:"notes"`
	TenantID string `json:"tenant_id"`
}

// CreateLeadHandler handles POST /api/v1/leads
func CreateLeadHandler(w http.ResponseWriter, r *http.Request) {
	store := r.Context().Value(CRMStoreKey).(CRMStore)

	var req CreateLeadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	lead := &agents.Lead{
		ID:       uuid.New().String(),
		TenantID: req.TenantID,
		Name:     req.Name,
		Email:    req.Email,
		Phone:    req.Phone,
		Company:  req.Company,
		Status:   "lead",
		Source:   req.Source,
		Notes:    req.Notes,
		Metadata: make(map[string]string),
	}

	if err := store.CreateLead(r.Context(), lead); err != nil {
		http.Error(w, `{"error": "failed to create lead"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(lead)
}

// ListLeadsHandler handles GET /api/v1/leads
func ListLeadsHandler(w http.ResponseWriter, r *http.Request) {
	store := r.Context().Value(CRMStoreKey).(CRMStore)
	tenantID := r.URL.Query().Get("tenant_id")

	leads, err := store.ListLeads(r.Context(), tenantID)
	if err != nil {
		http.Error(w, `{"error": "failed to list leads"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"leads": leads,
		"total": len(leads),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetLeadHandler handles GET /api/v1/leads/{id}
func GetLeadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	leadID := vars["id"]

	store := r.Context().Value(CRMStoreKey).(CRMStore)
	lead, err := store.GetLead(r.Context(), leadID)
	if err != nil {
		http.Error(w, `{"error": "failed to get lead"}`, http.StatusInternalServerError)
		return
	}

	if lead == nil {
		http.Error(w, `{"error": "lead not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lead)
}

// UpdateLeadRequest is the request body for updating a lead
type UpdateLeadRequest struct {
	Name     string            `json:"name"`
	Email    string            `json:"email"`
	Phone    string            `json:"phone"`
	Company  string            `json:"company"`
	Status   string            `json:"status"`
	Notes    string            `json:"notes"`
	Metadata map[string]string `json:"metadata"`
}

// UpdateLeadHandler handles PUT /api/v1/leads/{id}
func UpdateLeadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	leadID := vars["id"]

	store := r.Context().Value(CRMStoreKey).(CRMStore)
	lead, err := store.GetLead(r.Context(), leadID)
	if err != nil {
		http.Error(w, `{"error": "failed to get lead"}`, http.StatusInternalServerError)
		return
	}

	if lead == nil {
		http.Error(w, `{"error": "lead not found"}`, http.StatusNotFound)
		return
	}

	var req UpdateLeadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Name != "" {
		lead.Name = req.Name
	}
	if req.Email != "" {
		lead.Email = req.Email
	}
	if req.Phone != "" {
		lead.Phone = req.Phone
	}
	if req.Company != "" {
		lead.Company = req.Company
	}
	if req.Status != "" {
		lead.Status = req.Status
	}
	if req.Notes != "" {
		lead.Notes = req.Notes
	}
	if req.Metadata != nil {
		lead.Metadata = req.Metadata
	}

	if err := store.UpdateLead(r.Context(), lead); err != nil {
		http.Error(w, `{"error": "failed to update lead"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lead)
}

// DeleteLeadHandler handles DELETE /api/v1/leads/{id}
func DeleteLeadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	leadID := vars["id"]

	store := r.Context().Value(CRMStoreKey).(CRMStore)
	if err := store.DeleteLead(r.Context(), leadID); err != nil {
		http.Error(w, `{"error": "failed to delete lead"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ============================================================================
// Task Endpoints
// ============================================================================

// CreateTaskRequest is the request body for creating a task
type CreateTaskRequest struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Priority    string            `json:"priority"`
	Assignee    string            `json:"assignee"`
	DueDate     time.Time         `json:"due_date"`
	LeadID      string            `json:"lead_id"`
	TenantID    string            `json:"tenant_id"`
	Metadata    map[string]string `json:"metadata"`
}

// CreateTaskHandler handles POST /api/v1/tasks
func CreateTaskHandler(w http.ResponseWriter, r *http.Request) {
	store := r.Context().Value(CRMStoreKey).(CRMStore)

	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	task := &agents.Task{
		ID:          uuid.New().String(),
		TenantID:    req.TenantID,
		Title:       req.Title,
		Description: req.Description,
		Status:      "todo",
		Priority:    req.Priority,
		Assignee:    req.Assignee,
		DueDate:     req.DueDate,
		LeadID:      req.LeadID,
		Metadata:    req.Metadata,
	}

	if err := store.CreateTask(r.Context(), task); err != nil {
		http.Error(w, `{"error": "failed to create task"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

// ListTasksHandler handles GET /api/v1/tasks
func ListTasksHandler(w http.ResponseWriter, r *http.Request) {
	store := r.Context().Value(CRMStoreKey).(CRMStore)
	tenantID := r.URL.Query().Get("tenant_id")

	tasks, err := store.ListTasks(r.Context(), tenantID)
	if err != nil {
		http.Error(w, `{"error": "failed to list tasks"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"tasks": tasks,
		"total": len(tasks),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetTaskHandler handles GET /api/v1/tasks/{id}
func GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	store := r.Context().Value(CRMStoreKey).(CRMStore)
	task, err := store.GetTask(r.Context(), taskID)
	if err != nil {
		http.Error(w, `{"error": "failed to get task"}`, http.StatusInternalServerError)
		return
	}

	if task == nil {
		http.Error(w, `{"error": "task not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// UpdateTaskRequest is the request body for updating a task
type UpdateTaskRequest struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      string            `json:"status"`
	Priority    string            `json:"priority"`
	Assignee    string            `json:"assignee"`
	DueDate     time.Time         `json:"due_date"`
	Metadata    map[string]string `json:"metadata"`
}

// UpdateTaskHandler handles PUT /api/v1/tasks/{id}
func UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	store := r.Context().Value(CRMStoreKey).(CRMStore)
	task, err := store.GetTask(r.Context(), taskID)
	if err != nil {
		http.Error(w, `{"error": "failed to get task"}`, http.StatusInternalServerError)
		return
	}

	if task == nil {
		http.Error(w, `{"error": "task not found"}`, http.StatusNotFound)
		return
	}

	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Title != "" {
		task.Title = req.Title
	}
	if req.Description != "" {
		task.Description = req.Description
	}
	if req.Status != "" {
		task.Status = req.Status
	}
	if req.Priority != "" {
		task.Priority = req.Priority
	}
	if req.Assignee != "" {
		task.Assignee = req.Assignee
	}
	if !req.DueDate.IsZero() {
		task.DueDate = req.DueDate
	}
	if req.Metadata != nil {
		task.Metadata = req.Metadata
	}

	if err := store.UpdateTask(r.Context(), task); err != nil {
		http.Error(w, `{"error": "failed to update task"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// DeleteTaskHandler handles DELETE /api/v1/tasks/{id}
func DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	store := r.Context().Value(CRMStoreKey).(CRMStore)
	if err := store.DeleteTask(r.Context(), taskID); err != nil {
		http.Error(w, `{"error": "failed to delete task"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
