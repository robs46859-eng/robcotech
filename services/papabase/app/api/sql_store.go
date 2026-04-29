package api

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/robs46859-eng/fullstackarkham/services/papabase/app/agents"
	"gorm.io/gorm"
)

// JSONB is a custom type for map[string]string to handle PostgreSQL JSONB
type JSONB map[string]string

// Value implements the driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return "{}", nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONB)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, j)
}

// GORM Models

type LeadModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`
	TenantID  uuid.UUID `gorm:"type:uuid;not null;index"`
	Name      string    `gorm:"size:255;not null"`
	Email     string    `gorm:"size:255"`
	Phone     string    `gorm:"size:50"`
	Company   string    `gorm:"size:255"`
	Status    string    `gorm:"size:50;not null;default:'lead'"`
	Source    string    `gorm:"size:100"`
	Notes     string    `gorm:"type:text"`
	Metadata  JSONB     `gorm:"type:jsonb;default:'{}'"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (LeadModel) TableName() string {
	return "crm_leads"
}

type TaskModel struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key"`
	TenantID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	LeadID      *uuid.UUID `gorm:"type:uuid"`
	Title       string     `gorm:"size:255;not null"`
	Description string     `gorm:"type:text"`
	Status      string     `gorm:"size:50;not null;default:'pending'"`
	Priority    string     `gorm:"size:50;not null;default:'medium'"`
	Assignee    string     `gorm:"size:255"`
	DueDate     *time.Time
	Metadata    JSONB      `gorm:"type:jsonb;default:'{}'"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (TaskModel) TableName() string {
	return "crm_tasks"
}

// SQLCRMStore implements CRMStore using GORM
type SQLCRMStore struct {
	DB *gorm.DB
}

func NewSQLCRMStore(db *gorm.DB) *SQLCRMStore {
	return &SQLCRMStore{DB: db}
}

// Mappings

func toLeadModel(l *agents.Lead) *LeadModel {
	uid, _ := uuid.Parse(l.ID)
	tid, _ := uuid.Parse(l.TenantID)
	if uid == uuid.Nil {
		uid = uuid.New()
		l.ID = uid.String()
	}
	return &LeadModel{
		ID:        uid,
		TenantID:  tid,
		Name:      l.Name,
		Email:     l.Email,
		Phone:     l.Phone,
		Company:   l.Company,
		Status:    l.Status,
		Source:    l.Source,
		Notes:     l.Notes,
		Metadata:  JSONB(l.Metadata),
		CreatedAt: l.CreatedAt,
		UpdatedAt: l.UpdatedAt,
	}
}

func fromLeadModel(m *LeadModel) *agents.Lead {
	return &agents.Lead{
		ID:        m.ID.String(),
		TenantID:  m.TenantID.String(),
		Name:      m.Name,
		Email:     m.Email,
		Phone:     m.Phone,
		Company:   m.Company,
		Status:    m.Status,
		Source:    m.Source,
		Notes:     m.Notes,
		Metadata:  map[string]string(m.Metadata),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func toTaskModel(t *agents.Task) *TaskModel {
	uid, _ := uuid.Parse(t.ID)
	tid, _ := uuid.Parse(t.TenantID)
	if uid == uuid.Nil {
		uid = uuid.New()
		t.ID = uid.String()
	}
	var lid *uuid.UUID
	if t.LeadID != "" {
		parsed, err := uuid.Parse(t.LeadID)
		if err == nil {
			lid = &parsed
		}
	}
	
	var dueDate *time.Time
	if !t.DueDate.IsZero() {
		dueDate = &t.DueDate
	}

	return &TaskModel{
		ID:          uid,
		TenantID:    tid,
		LeadID:      lid,
		Title:       t.Title,
		Description: t.Description,
		Status:      t.Status,
		Priority:    t.Priority,
		Assignee:    t.Assignee,
		DueDate:     dueDate,
		Metadata:    JSONB(t.Metadata),
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func fromTaskModel(m *TaskModel) *agents.Task {
	var lid string
	if m.LeadID != nil {
		lid = m.LeadID.String()
	}
	var dueDate time.Time
	if m.DueDate != nil {
		dueDate = *m.DueDate
	}
	return &agents.Task{
		ID:          m.ID.String(),
		TenantID:    m.TenantID.String(),
		LeadID:      lid,
		Title:       m.Title,
		Description: m.Description,
		Status:      m.Status,
		Priority:    m.Priority,
		Assignee:    m.Assignee,
		DueDate:     dueDate,
		Metadata:    map[string]string(m.Metadata),
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// CRMStore Implementation

func (s *SQLCRMStore) CreateLead(ctx context.Context, lead *agents.Lead) error {
	m := toLeadModel(lead)
	return s.DB.WithContext(ctx).Create(m).Error
}

func (s *SQLCRMStore) GetLead(ctx context.Context, id string) (*agents.Lead, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid: %w", err)
	}
	var m LeadModel
	if err := s.DB.WithContext(ctx).First(&m, "id = ?", uid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return fromLeadModel(&m), nil
}

func (s *SQLCRMStore) ListLeads(ctx context.Context, tenantID string) ([]*agents.Lead, error) {
	query := s.DB.WithContext(ctx)
	if tenantID != "" {
		tid, err := uuid.Parse(tenantID)
		if err == nil {
			query = query.Where("tenant_id = ?", tid)
		}
	}
	var models []*LeadModel
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}
	results := make([]*agents.Lead, len(models))
	for i, m := range models {
		results[i] = fromLeadModel(m)
	}
	return results, nil
}

func (s *SQLCRMStore) UpdateLead(ctx context.Context, lead *agents.Lead) error {
	m := toLeadModel(lead)
	return s.DB.WithContext(ctx).Save(m).Error
}

func (s *SQLCRMStore) DeleteLead(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid uuid: %w", err)
	}
	return s.DB.WithContext(ctx).Delete(&LeadModel{}, "id = ?", uid).Error
}

func (s *SQLCRMStore) CreateTask(ctx context.Context, task *agents.Task) error {
	m := toTaskModel(task)
	return s.DB.WithContext(ctx).Create(m).Error
}

func (s *SQLCRMStore) GetTask(ctx context.Context, id string) (*agents.Task, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid: %w", err)
	}
	var m TaskModel
	if err := s.DB.WithContext(ctx).First(&m, "id = ?", uid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return fromTaskModel(&m), nil
}

func (s *SQLCRMStore) ListTasks(ctx context.Context, tenantID string) ([]*agents.Task, error) {
	query := s.DB.WithContext(ctx)
	if tenantID != "" {
		tid, err := uuid.Parse(tenantID)
		if err == nil {
			query = query.Where("tenant_id = ?", tid)
		}
	}
	var models []*TaskModel
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}
	results := make([]*agents.Task, len(models))
	for i, m := range models {
		results[i] = fromTaskModel(m)
	}
	return results, nil
}

func (s *SQLCRMStore) UpdateTask(ctx context.Context, task *agents.Task) error {
	m := toTaskModel(task)
	return s.DB.WithContext(ctx).Save(m).Error
}

func (s *SQLCRMStore) DeleteTask(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid uuid: %w", err)
	}
	return s.DB.WithContext(ctx).Delete(&TaskModel{}, "id = ?", uid).Error
}
