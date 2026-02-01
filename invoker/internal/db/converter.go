package db

import (
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/revrost/invoker/internal/db/sqlc"
	"github.com/revrost/invoker/pkg/models"
)

// Model converters from sqlc models to pkg/models

// UserFromDB converts database User model to pkg/models.User
func UserFromDB(u sqlc.Profile) *models.User {
	return &models.User{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Username:  u.Username,
		Tier:      u.Tier,
		CreatedAt: time.UnixMilli(u.CreatedAt),
		UpdatedAt: time.UnixMilli(u.UpdatedAt),
	}
}

// SubscriptionFromDB converts database Subscription model to pkg/models.Subscription
func SubscriptionFromDB(s sqlc.Subscription) *models.Subscription {
	return &models.Subscription{
		ID:                 s.ID,
		UserID:             s.ProfileID,
		StripeSubID:        s.StripeSubID.String,
		Tier:               s.Tier,
		Status:             s.Status,
		CurrentPeriodStart: time.UnixMilli(s.CurrentPeriodStart),
		CurrentPeriodEnd:   time.UnixMilli(s.CurrentPeriodEnd),
		CancelAtPeriodEnd:  s.CancelAtPeriodEnd,
		CreatedAt:          time.UnixMilli(s.CreatedAt),
		UpdatedAt:          time.UnixMilli(s.UpdatedAt),
	}
}

// MachineRegistryFromDB converts database MachineRegistry model to pkg/models.MachineRegistry
func MachineRegistryFromDB(m sqlc.MachineRegistry) *models.MachineRegistry {
	var lastHeartbeat *time.Time
	if m.LastHeartbeatAt.Valid {
		t := time.UnixMilli(m.LastHeartbeatAt.Int64)
		lastHeartbeat = &t
	}

	return &models.MachineRegistry{
		ID:              m.ID,
		UserID:          m.ProfileID,
		FlyMachineID:    m.FlyMachineID,
		FlyAppName:      m.FlyAppName,
		Status:          m.Status,
		Subdomain:       m.Subdomain,
		PublicURL:       m.PublicUrl,
		Region:          m.Region,
		VMSize:          m.VmSize,
		VolumeID:        m.VolumeID.String,
		CreatedAt:       time.UnixMilli(m.CreatedAt),
		LastSeenAt:      time.UnixMilli(m.LastSeenAt),
		LastHeartbeatAt: lastHeartbeat,
		ErrorMessage:    m.ErrorMessage.String,
	}
}

// RoutingTableFromDB converts database RoutingTable model to pkg/models.RoutingTable
func RoutingTableFromDB(r sqlc.RoutingTable) *models.RoutingTable {
	return &models.RoutingTable{
		Subdomain:    r.Subdomain,
		FlyMachineID: r.FlyMachineID,
		FlyURL:       r.FlyUrl,
		IsActive:     r.IsActive,
		UpdatedAt:    time.UnixMilli(r.UpdatedAt),
	}
}

// QuotaLimitFromDB converts database QuotaLimit model to pkg/models.QuotaLimit
func QuotaLimitFromDB(q sqlc.QuotaLimit) *models.QuotaLimit {
	return &models.QuotaLimit{
		Tier:                   q.Tier,
		MaxVMCount:             int(q.MaxVmCount),
		MaxVMHoursPerMonth:     int(q.MaxVmHoursPerMonth),
		MaxTasksPerMonth:       int(q.MaxTasksPerMonth),
		MaxAPIRequestsPerMonth: int(q.MaxApiRequestsPerMonth),
	}
}

// UsageTrackingFromDB converts database UsageTracking model to pkg/models.UsageTracking
func UsageTrackingFromDB(u sqlc.UsageTracking) *models.UsageTracking {
	return &models.UsageTracking{
		ID:          u.ID,
		UserID:      u.ProfileID,
		MachineID:   u.MachineID,
		MetricType:  u.MetricType,
		Quantity:    int(u.Quantity),
		RecordedAt:  time.UnixMilli(u.RecordedAt),
		PeriodStart: time.UnixMilli(u.PeriodStart),
		PeriodEnd:   time.UnixMilli(u.PeriodEnd),
	}
}

// AuditLogFromDB converts database AuditLog model to pkg/models.AuditLog
func AuditLogFromDB(a sqlc.AuditLog) *models.AuditLog {
	return &models.AuditLog{
		ID:           a.ID,
		UserID:       a.ProfileID.String,
		Action:       a.Action,
		ResourceType: a.ResourceType,
		ResourceID:   a.ResourceID.String,
		IPAddress:    a.IpAddress.String,
		UserAgent:    a.UserAgent.String,
		CreatedAt:    time.UnixMilli(a.CreatedAt),
	}
}

// CreateUserParamsFromRequest converts RegisterRequest to CreateUserParams
func CreateUserParamsFromRequest(req *models.RegisterRequest, id string) sqlc.CreateUserParams {
	// Generate username from email (everything before @)
	username := req.Email
	if atIndex := strings.Index(req.Email, "@"); atIndex > 0 {
		username = req.Email[:atIndex]
	}

	return sqlc.CreateUserParams{
		ID:        id,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Username:  username,
		Tier:      "free",
	}
}

// ToPgText converts string to pgtype.Text
func ToPgText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}

// FromPgText converts pgtype.Text to string
func FromPgText(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

// ToPgInt8 converts *int64 to pgtype.Int8
func ToPgInt8(i *int64) pgtype.Int8 {
	if i == nil {
		return pgtype.Int8{Valid: false}
	}
	return pgtype.Int8{Int64: *i, Valid: true}
}

// FromPgInt8 converts pgtype.Int8 to *int64
func FromPgInt8(i pgtype.Int8) *int64 {
	if !i.Valid {
		return nil
	}
	return &i.Int64
}
