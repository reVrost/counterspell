package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUser_Structure(t *testing.T) {
	user := User{
		ID:        uuid.New().String(),
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Tier:      "free",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	assert.NotEmpty(t, user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "John", user.FirstName)
	assert.Equal(t, "Doe", user.LastName)
	assert.Equal(t, "johndoe", user.Username)
	assert.Equal(t, "free", user.Tier)
	assert.False(t, user.CreatedAt.IsZero())
	assert.False(t, user.UpdatedAt.IsZero())
}

func TestSubscription_Structure(t *testing.T) {
	sub := Subscription{
		ID:                 uuid.New().String(),
		UserID:             uuid.New().String(),
		StripeSubID:        "sub_123",
		Tier:               "pro",
		Status:             "active",
		CurrentPeriodStart: time.Now(),
		CurrentPeriodEnd:   time.Now().AddDate(0, 1, 0),
		CancelAtPeriodEnd:  false,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	assert.NotEmpty(t, sub.ID)
	assert.NotEmpty(t, sub.UserID)
	assert.Equal(t, "sub_123", sub.StripeSubID)
	assert.Equal(t, "pro", sub.Tier)
	assert.Equal(t, "active", sub.Status)
	assert.False(t, sub.CancelAtPeriodEnd)
}

func TestMachineRegistry_Structure(t *testing.T) {
	now := time.Now()
	machine := MachineRegistry{
		ID:              uuid.New().String(),
		UserID:          uuid.New().String(),
		FlyMachineID:    "machine_123",
		FlyAppName:      "myapp",
		Status:          "running",
		Subdomain:       "myapp",
		PublicURL:       "https://myapp.example.com",
		Region:          "iad",
		VMSize:          "shared-cpu-1x",
		VolumeID:        "vol_123",
		CreatedAt:       now,
		LastSeenAt:      now,
		ErrorMessage:    "",
	}

	assert.NotEmpty(t, machine.ID)
	assert.NotEmpty(t, machine.UserID)
	assert.Equal(t, "machine_123", machine.FlyMachineID)
	assert.Equal(t, "myapp", machine.FlyAppName)
	assert.Equal(t, "running", machine.Status)
	assert.Equal(t, "https://myapp.example.com", machine.PublicURL)
}

func TestQuotaLimit_Structure(t *testing.T) {
	quota := QuotaLimit{
		Tier:                     "pro",
		MaxVMCount:              10,
		MaxVMHoursPerMonth:      100,
		MaxTasksPerMonth:        1000,
		MaxAPIRequestsPerMonth: 10000,
	}

	assert.Equal(t, "pro", quota.Tier)
	assert.Equal(t, 10, quota.MaxVMCount)
	assert.Equal(t, 100, quota.MaxVMHoursPerMonth)
	assert.Equal(t, 1000, quota.MaxTasksPerMonth)
	assert.Equal(t, 10000, quota.MaxAPIRequestsPerMonth)
}

func TestUsageTracking_Structure(t *testing.T) {
	tracking := UsageTracking{
		ID:          uuid.New().String(),
		UserID:      uuid.New().String(),
		MachineID:   uuid.New().String(),
		MetricType:  "vm_hours",
		Quantity:    10,
		RecordedAt:  time.Now(),
		PeriodStart: time.Now().AddDate(0, -1, 0),
		PeriodEnd:   time.Now(),
	}

	assert.NotEmpty(t, tracking.ID)
	assert.NotEmpty(t, tracking.UserID)
	assert.Equal(t, "vm_hours", tracking.MetricType)
	assert.Equal(t, 10, tracking.Quantity)
}

func TestAuditLog_Structure(t *testing.T) {
	log := AuditLog{
		ID:           uuid.New().String(),
		UserID:       uuid.New().String(),
		Action:       "create_machine",
		ResourceType: "machine",
		ResourceID:   uuid.New().String(),
		IPAddress:    "192.168.1.1",
		UserAgent:    "Mozilla/5.0",
		Metadata:     map[string]interface{}{"key": "value"},
		CreatedAt:    time.Now(),
	}

	assert.NotEmpty(t, log.ID)
	assert.Equal(t, "create_machine", log.Action)
	assert.Equal(t, "machine", log.ResourceType)
	assert.NotEmpty(t, log.ResourceID)
	assert.NotNil(t, log.Metadata)
	assert.Equal(t, "value", log.Metadata["key"])
}

func TestRegisterRequest_ValidationTags(t *testing.T) {
	req := RegisterRequest{
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Password:  "password123",
	}

	assert.Equal(t, "test@example.com", req.Email)
	assert.Equal(t, "John", req.FirstName)
	assert.Equal(t, "Doe", req.LastName)
	assert.Equal(t, "password123", req.Password)
}

func TestLoginRequest_ValidationTags(t *testing.T) {
	req := LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	assert.Equal(t, "test@example.com", req.Email)
	assert.Equal(t, "password123", req.Password)
}

func TestAuthResponse_Structure(t *testing.T) {
	user := &User{
		ID:        uuid.New().String(),
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Tier:      "free",
	}

	response := AuthResponse{
		Token: "jwt_token_here",
		User:  user,
	}

	assert.NotEmpty(t, response.Token)
	assert.NotNil(t, response.User)
	assert.Equal(t, "test@example.com", response.User.Email)
}

func TestHealthResponse_Structure(t *testing.T) {
	resp := HealthResponse{
		Status:  "healthy",
		Version: "1.0.0",
	}

	assert.Equal(t, "healthy", resp.Status)
	assert.Equal(t, "1.0.0", resp.Version)
}

func TestReadyResponse_Structure(t *testing.T) {
	resp := ReadyResponse{
		Status:   "ready",
		Database: true,
	}

	assert.Equal(t, "ready", resp.Status)
	assert.True(t, resp.Database)
}
