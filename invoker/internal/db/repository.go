package db

import (
	"context"
	"time"

	"github.com/revrost/invoker/pkg/models"

	"github.com/revrost/invoker/internal/db/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
)

//go:generate mockgen -destination=mock_repository.go -package=db github.com/revrost/invoker/internal/db Repository

// Repository defines the interface for database operations needed by handlers
type Repository interface {
	Queries() *sqlc.Queries
	// User operations
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	UsernameExists(ctx context.Context, username string) (bool, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	// Machine registry operations
	CreateMachine(ctx context.Context, machine *models.MachineRegistry) error
	GetMachineByUserID(ctx context.Context, userID string) (*models.MachineRegistry, error)
	UpdateMachineStatus(ctx context.Context, machineID, status string, lastSeenAt time.Time, errorMessage string) error
	// OAuth pending login operations
	CreatePendingOAuthLogin(ctx context.Context, login *PendingOAuthLogin) error
	GetPendingOAuthLoginByState(ctx context.Context, state string) (*PendingOAuthLogin, error)
	DeletePendingOAuthLogin(ctx context.Context, state string) error
	CleanupExpiredOAuthLogins(ctx context.Context) error
	// Machine auth operations
	CreateMachineAuth(ctx context.Context, auth *MachineAuth) error
	GetMachineAuthByMachineID(ctx context.Context, machineID string) (*MachineAuth, error)
	GetMachineAuthByUserID(ctx context.Context, userID string) ([]*MachineAuth, error)
	GetMachineAuthBySubdomain(ctx context.Context, subdomain string) (*MachineAuth, error)
	RevokeMachineAuth(ctx context.Context, machineID string) error
	UpdateMachineAuthLastSeen(ctx context.Context, machineID string, lastSeenAt time.Time) error
	UpdateMachineAuthTunnel(ctx context.Context, machineID, tunnelProvider, tunnelToken, subdomain string) (*MachineAuth, error)
}

// PostgresRepository implements Service using *DB struct
type PostgresRepository struct {
	db *DB
}

// NewPostgresRepository creates a new DBService
func NewPostgresRepository(db *DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Queries returns the underlying sqlc queries
func (s *PostgresRepository) Queries() *sqlc.Queries {
	return s.db.queries
}

// CreateUser creates a new user in database
func (s *PostgresRepository) CreateUser(ctx context.Context, user *models.User) error {
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	params := sqlc.CreateUserParams{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Username:  user.Username,
		Tier:      user.Tier,
	}

	_, err := s.db.queries.CreateUser(ctx, params)
	return err
}

// GetUserByID retrieves a user by ID
func (s *PostgresRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	dbUser, err := s.db.queries.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return UserFromDB(dbUser), nil
}

// GetUserByEmail retrieves a user by email
func (s *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	dbUser, err := s.db.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return UserFromDB(dbUser), nil
}

// GetUserByUsername retrieves a user by username
func (s *PostgresRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	dbUser, err := s.db.queries.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	return UserFromDB(dbUser), nil
}

// UsernameExists checks if a username already exists
func (s *PostgresRepository) UsernameExists(ctx context.Context, username string) (bool, error) {
	return s.db.queries.UsernameExists(ctx, username)
}

// EmailExists checks if an email already exists
func (s *PostgresRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	return s.db.queries.EmailExists(ctx, email)
}

// CreateMachine creates a new machine registry entry
func (s *PostgresRepository) CreateMachine(ctx context.Context, machine *models.MachineRegistry) error {
	now := time.Now()
	machine.CreatedAt = now
	machine.LastSeenAt = now

	params := sqlc.CreateMachineRegistryParams{
		ID:              machine.ID,
		ProfileID:       machine.UserID,
		FlyMachineID:    machine.FlyMachineID,
		FlyAppName:      machine.FlyAppName,
		Status:          machine.Status,
		Subdomain:       machine.Subdomain,
		PublicUrl:       machine.PublicURL,
		Region:          machine.Region,
		VmSize:          machine.VMSize,
		VolumeID:        ToPgText(machine.VolumeID),
		CreatedAt:       now.UnixMilli(),
		LastSeenAt:      now.UnixMilli(),
		LastHeartbeatAt: ToPgInt8(nil),
		ErrorMessage:    ToPgText(machine.ErrorMessage),
	}

	_, err := s.db.queries.CreateMachineRegistry(ctx, params)
	return err
}

// GetMachineByUserID retrieves a machine by user ID
func (s *PostgresRepository) GetMachineByUserID(ctx context.Context, userID string) (*models.MachineRegistry, error) {
	dbMachines, err := s.db.queries.GetMachineRegistryByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(dbMachines) == 0 {
		return nil, nil
	}
	// Return the most recent machine (should only be one for MVP)
	return MachineRegistryFromDB(dbMachines[0]), nil
}

// UpdateMachineStatus updates the status of a machine
func (s *PostgresRepository) UpdateMachineStatus(ctx context.Context, machineID, status string, lastSeenAt time.Time, errorMessage string) error {
	params := sqlc.UpdateMachineRegistryStatusParams{
		ID:           machineID,
		Status:       status,
		LastSeenAt:   lastSeenAt.UnixMilli(),
		ErrorMessage: ToPgText(errorMessage),
	}

	_, err := s.db.queries.UpdateMachineRegistryStatus(ctx, params)
	return err
}

// PendingOAuthLogin represents a pending OAuth login attempt
type PendingOAuthLogin struct {
	ID            string
	State         string
	CodeChallenge string
	RedirectURI   string
	AuthCode      string
	CreatedAt     time.Time
	ExpiresAt     time.Time
}

// CreatePendingOAuthLogin creates a new pending OAuth login entry
func (s *PostgresRepository) CreatePendingOAuthLogin(ctx context.Context, login *PendingOAuthLogin) error {
	params := sqlc.CreatePendingOAuthLoginParams{
		ID:            login.ID,
		State:         login.State,
		CodeChallenge: login.CodeChallenge,
		RedirectUri:   login.RedirectURI,
		CreatedAt:     login.CreatedAt.UnixMilli(),
		ExpiresAt:     login.ExpiresAt.UnixMilli(),
	}

	_, err := s.db.queries.CreatePendingOAuthLogin(ctx, params)
	return err
}

// GetPendingOAuthLoginByState retrieves a pending OAuth login by state
func (s *PostgresRepository) GetPendingOAuthLoginByState(ctx context.Context, state string) (*PendingOAuthLogin, error) {
	dbLogin, err := s.db.queries.GetPendingOAuthLoginByState(ctx, state)
	if err != nil {
		return nil, err
	}
	return &PendingOAuthLogin{
		ID:            dbLogin.ID,
		State:         dbLogin.State,
		CodeChallenge: dbLogin.CodeChallenge,
		RedirectURI:   dbLogin.RedirectUri,
		AuthCode:      dbLogin.AuthCode,
		CreatedAt:     time.UnixMilli(dbLogin.CreatedAt),
		ExpiresAt:     time.UnixMilli(dbLogin.ExpiresAt),
	}, nil
}

// DeletePendingOAuthLogin deletes a pending OAuth login
func (s *PostgresRepository) DeletePendingOAuthLogin(ctx context.Context, state string) error {
	return s.db.queries.DeletePendingOAuthLogin(ctx, state)
}

// CleanupExpiredOAuthLogins removes expired pending OAuth logins
func (s *PostgresRepository) CleanupExpiredOAuthLogins(ctx context.Context) error {
	return s.db.queries.CleanupExpiredOAuthLogins(ctx)
}

// MachineAuth represents a machine authentication record
type MachineAuth struct {
	ID             string
	MachineID      string
	UserID         string
	Subdomain      string
	TunnelProvider string
	TunnelToken    string
	CreatedAt      time.Time
	LastSeenAt     *time.Time
	IsActive       bool
}

// CreateMachineAuth creates a new machine auth entry
func (s *PostgresRepository) CreateMachineAuth(ctx context.Context, auth *MachineAuth) error {
	var lastSeenAt pgtype.Int8
	if auth.LastSeenAt != nil {
		lastSeenAt = ToPgInt8(&[]int64{auth.LastSeenAt.UnixMilli()}[0])
	}

	params := sqlc.CreateMachineAuthParams{
		ID:             auth.ID,
		MachineID:      auth.MachineID,
		ProfileID:      auth.UserID,
		Subdomain:      auth.Subdomain,
		TunnelProvider: auth.TunnelProvider,
		TunnelToken:    auth.TunnelToken,
		CreatedAt:      auth.CreatedAt.UnixMilli(),
		LastSeenAt:     lastSeenAt,
		IsActive:       auth.IsActive,
	}

	_, err := s.db.queries.CreateMachineAuth(ctx, params)
	return err
}

// GetMachineAuthByMachineID retrieves machine auth by machine ID
func (s *PostgresRepository) GetMachineAuthByMachineID(ctx context.Context, machineID string) (*MachineAuth, error) {
	dbAuth, err := s.db.queries.GetMachineAuthByMachineID(ctx, machineID)
	if err != nil {
		return nil, err
	}

	var lastSeenAt *time.Time
	if dbAuth.LastSeenAt.Valid {
		ts := time.UnixMilli(dbAuth.LastSeenAt.Int64)
		lastSeenAt = &ts
	}

	return &MachineAuth{
		ID:             dbAuth.ID,
		MachineID:      dbAuth.MachineID,
		UserID:         dbAuth.ProfileID,
		Subdomain:      dbAuth.Subdomain,
		TunnelProvider: dbAuth.TunnelProvider,
		TunnelToken:    dbAuth.TunnelToken,
		CreatedAt:      time.UnixMilli(dbAuth.CreatedAt),
		LastSeenAt:     lastSeenAt,
		IsActive:       dbAuth.IsActive,
	}, nil
}

// GetMachineAuthByUserID retrieves all active machine auth records for a user.
func (s *PostgresRepository) GetMachineAuthByUserID(ctx context.Context, userID string) ([]*MachineAuth, error) {
	dbAuths, err := s.db.queries.GetMachineAuthByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]*MachineAuth, 0, len(dbAuths))
	for _, dbAuth := range dbAuths {
		var lastSeenAt *time.Time
		if dbAuth.LastSeenAt.Valid {
			ts := time.UnixMilli(dbAuth.LastSeenAt.Int64)
			lastSeenAt = &ts
		}
		result = append(result, &MachineAuth{
			ID:             dbAuth.ID,
			MachineID:      dbAuth.MachineID,
			UserID:         dbAuth.ProfileID,
			Subdomain:      dbAuth.Subdomain,
			TunnelProvider: dbAuth.TunnelProvider,
			TunnelToken:    dbAuth.TunnelToken,
			CreatedAt:      time.UnixMilli(dbAuth.CreatedAt),
			LastSeenAt:     lastSeenAt,
			IsActive:       dbAuth.IsActive,
		})
	}

	return result, nil
}

// GetMachineAuthBySubdomain retrieves machine auth by subdomain
func (s *PostgresRepository) GetMachineAuthBySubdomain(ctx context.Context, subdomain string) (*MachineAuth, error) {
	dbAuth, err := s.db.queries.GetMachineAuthBySubdomain(ctx, subdomain)
	if err != nil {
		return nil, err
	}

	var lastSeenAt *time.Time
	if dbAuth.LastSeenAt.Valid {
		ts := time.UnixMilli(dbAuth.LastSeenAt.Int64)
		lastSeenAt = &ts
	}

	return &MachineAuth{
		ID:             dbAuth.ID,
		MachineID:      dbAuth.MachineID,
		UserID:         dbAuth.ProfileID,
		Subdomain:      dbAuth.Subdomain,
		TunnelProvider: dbAuth.TunnelProvider,
		TunnelToken:    dbAuth.TunnelToken,
		CreatedAt:      time.UnixMilli(dbAuth.CreatedAt),
		LastSeenAt:     lastSeenAt,
		IsActive:       dbAuth.IsActive,
	}, nil
}

// RevokeMachineAuth revokes a machine auth entry
func (s *PostgresRepository) RevokeMachineAuth(ctx context.Context, machineID string) error {
	_, err := s.db.queries.RevokeMachineAuth(ctx, machineID)
	return err
}

// UpdateMachineAuthLastSeen updates the last_seen_at timestamp for a machine
func (s *PostgresRepository) UpdateMachineAuthLastSeen(ctx context.Context, machineID string, lastSeenAt time.Time) error {
	params := sqlc.UpdateMachineAuthLastSeenParams{
		MachineID:  machineID,
		LastSeenAt: pgtype.Int8{Int64: lastSeenAt.UnixMilli(), Valid: true},
	}
	_, err := s.db.queries.UpdateMachineAuthLastSeen(ctx, params)
	return err
}

// UpdateMachineAuthTunnel updates the tunnel metadata for a machine.
func (s *PostgresRepository) UpdateMachineAuthTunnel(ctx context.Context, machineID, tunnelProvider, tunnelToken, subdomain string) (*MachineAuth, error) {
	params := sqlc.UpdateMachineAuthTunnelParams{
		MachineID:      machineID,
		TunnelProvider: tunnelProvider,
		TunnelToken:    tunnelToken,
		Subdomain:      subdomain,
	}

	dbAuth, err := s.db.queries.UpdateMachineAuthTunnel(ctx, params)
	if err != nil {
		return nil, err
	}

	var lastSeenAt *time.Time
	if dbAuth.LastSeenAt.Valid {
		ts := time.UnixMilli(dbAuth.LastSeenAt.Int64)
		lastSeenAt = &ts
	}

	return &MachineAuth{
		ID:             dbAuth.ID,
		MachineID:      dbAuth.MachineID,
		UserID:         dbAuth.ProfileID,
		Subdomain:      dbAuth.Subdomain,
		TunnelProvider: dbAuth.TunnelProvider,
		TunnelToken:    dbAuth.TunnelToken,
		CreatedAt:      time.UnixMilli(dbAuth.CreatedAt),
		LastSeenAt:     lastSeenAt,
		IsActive:       dbAuth.IsActive,
	}, nil
}
