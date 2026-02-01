package models

import "time"

// User represents a user in the system
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Username  string    `json:"username"`
	Tier      string    `json:"tier"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Subscription represents a user's subscription
type Subscription struct {
	ID                 string    `json:"id"`
	UserID             string    `json:"user_id"`
	StripeSubID        string    `json:"stripe_sub_id,omitempty"`
	Tier               string    `json:"tier"`
	Status             string    `json:"status"`
	CurrentPeriodStart time.Time `json:"current_period_start"`
	CurrentPeriodEnd   time.Time `json:"current_period_end"`
	CancelAtPeriodEnd  bool      `json:"cancel_at_period_end"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// MachineRegistry represents a user's VM
type MachineRegistry struct {
	ID              string     `json:"id"`
	UserID          string     `json:"user_id"`
	FlyMachineID    string     `json:"fly_machine_id"`
	FlyAppName      string     `json:"fly_app_name"`
	Status          string     `json:"status"`
	Subdomain       string     `json:"subdomain"`
	PublicURL       string     `json:"public_url"`
	Region          string     `json:"region"`
	VMSize          string     `json:"vm_size"`
	VolumeID        string     `json:"volume_id,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	LastSeenAt      time.Time  `json:"last_seen_at"`
	LastHeartbeatAt *time.Time `json:"last_heartbeat_at,omitempty"`
	ErrorMessage    string     `json:"error_message,omitempty"`
}

// RoutingTable maps subdomains to Fly.io VM URLs
type RoutingTable struct {
	Subdomain    string    `json:"subdomain"`
	FlyMachineID string    `json:"fly_machine_id"`
	FlyURL       string    `json:"fly_url"`
	IsActive     bool      `json:"is_active"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// QuotaLimit defines limits per tier
type QuotaLimit struct {
	Tier                   string `json:"tier"`
	MaxVMCount             int    `json:"max_vm_count"`
	MaxVMHoursPerMonth     int    `json:"max_vm_hours_per_month"`
	MaxTasksPerMonth       int    `json:"max_tasks_per_month"`
	MaxAPIRequestsPerMonth int    `json:"max_api_requests_per_month"`
}

// UsageTracking tracks usage metrics
type UsageTracking struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	MachineID   string    `json:"machine_id"`
	MetricType  string    `json:"metric_type"`
	Quantity    int       `json:"quantity"`
	RecordedAt  time.Time `json:"recorded_at"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
}

// AuditLog represents an audit entry
type AuditLog struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id,omitempty"`
	Action       string                 `json:"action"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   string                 `json:"resource_id,omitempty"`
	IPAddress    string                 `json:"ip_address,omitempty"`
	UserAgent    string                 `json:"user_agent,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

// HealthResponse is the response for health checks
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// ReadyResponse is the response for readiness checks
type ReadyResponse struct {
	Status   string `json:"status"`
	Database bool   `json:"database"`
}

// RegisterRequest is the request body for user registration
type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	FirstName string `json:"first_name" validate:"required,min=1,max=50"`
	LastName  string `json:"last_name" validate:"required,min=1,max=50"`
	Password  string `json:"password" validate:"required,min=8"`
}

// LoginRequest is the request body for user login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse is the response for auth endpoints
type AuthResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}
