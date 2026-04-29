package agents

import "time"

// Lead represents a sales lead in the CRM
type Lead struct {
	ID          string            `json:"id"`
	TenantID    string            `json:"tenant_id"`
	Name        string            `json:"name"`
	Email       string            `json:"email"`
	Phone       string            `json:"phone"`
	Company     string            `json:"company"`
	Status      string            `json:"status"`
	Source      string            `json:"source"`
	Notes       string            `json:"notes"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// Task represents a task in the system
type Task struct {
	ID          string            `json:"id"`
	TenantID    string            `json:"tenant_id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      string            `json:"status"`
	Priority    string            `json:"priority"`
	Assignee    string            `json:"assignee"`
	DueDate     time.Time         `json:"due_date"`
	LeadID      string            `json:"lead_id"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}
