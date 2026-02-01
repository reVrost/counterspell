package fly

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/revrost/invoker/pkg/models"

	"github.com/google/uuid"
)

// Service manages Fly.io machines for users
type Service struct {
	client      *Client
	db          MachineRegistry
	appName     string // Fly.io app name for data plane machines
	dockerImage string // Docker image to deploy
	region      string // Default region
}

// MachineRegistry is the interface for machine registry operations
type MachineRegistry interface {
	CreateMachine(ctx context.Context, machine *models.MachineRegistry) error
	GetMachineByUserID(ctx context.Context, userID string) (*models.MachineRegistry, error)
}

// NewService creates a new Fly.io machine service
func NewService(client *Client, db MachineRegistry, appName, dockerImage, region string) *Service {
	return &Service{
		client:      client,
		db:          db,
		appName:     appName,
		dockerImage: dockerImage,
		region:      region,
	}
}

// ProvisionUserMachine creates a Fly.io machine for a new user
// This should be called during user registration (SyncProfile)
func (s *Service) ProvisionUserMachine(ctx context.Context, user *models.User) (*models.MachineRegistry, error) {
	// Check if user already has a machine
	existingMachine, err := s.db.GetMachineByUserID(ctx, user.ID)
	if err == nil && existingMachine != nil {
		slog.Info("User already has a machine", "user_id", user.ID, "machine_id", existingMachine.FlyMachineID)
		return existingMachine, nil
	}

	// Generate machine name
	machineName := fmt.Sprintf("counterspell-%s", user.Username)

	slog.Info("Provisioning machine for user",
		"user_id", user.ID,
		"username", user.Username,
		"machine_name", machineName,
	)

	// Create machine on Fly.io
	// Note: For now, we don't create a volume. Volumes must be pre-created via flyctl
	// TODO: Add volume creation support later
	machine, err := s.client.CreateMachine(ctx, &CreateMachineConfig{
		AppName:  s.appName,
		Name:     machineName,
		Region:   s.region,
		Image:    s.dockerImage,
		VMSize:   "shared-cpu-1x",
		MemoryMB: 1024,
		Env: map[string]string{
			"USER_ID":    user.ID,
			"USERNAME":   user.Username,
			"SUBDOMAIN":  user.Username,
			"TIER":       user.Tier,
		},
		InternalPort: 8080,
		// VolumeID will be added later once volume creation is implemented
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Fly.io machine: %w", err)
	}

	// Generate public URL
	publicURL := GetPublicURL(s.appName)

	// Create machine registry entry
	registryEntry := &models.MachineRegistry{
		ID:           uuid.New().String(),
		UserID:       user.ID,
		FlyMachineID: machine.ID,
		FlyAppName:   s.appName,
		Status:       "creating", // Will be updated to "running" by health check
		Subdomain:    user.Username,
		PublicURL:    publicURL,
		Region:       s.region,
		VMSize:       "shared-cpu-1x",
		VolumeID:     "", // Will be filled in later when volume support is added
		CreatedAt:    time.Now(),
		LastSeenAt:   time.Now(),
	}

	// Save to database
	if err := s.db.CreateMachine(ctx, registryEntry); err != nil {
		// Attempt to cleanup the machine if database insert fails
		slog.Error("Failed to create machine registry entry, cleaning up machine",
			"error", err,
			"machine_id", machine.ID,
		)
		_ = s.client.StopMachine(ctx, s.appName, machine.ID)
		return nil, fmt.Errorf("failed to create machine registry entry: %w", err)
	}

	slog.Info("Machine provisioned successfully for user",
		"user_id", user.ID,
		"machine_id", machine.ID,
		"fly_machine_id", machine.ID,
		"subdomain", user.Username,
		"public_url", publicURL,
	)

	return registryEntry, nil
}

// GetMachineStatus retrieves the current status of a user's machine
func (s *Service) GetMachineStatus(ctx context.Context, userID string) (*models.MachineRegistry, error) {
	registry, err := s.db.GetMachineByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get machine from registry: %w", err)
	}

	// Get current status from Fly.io
	machine, err := s.client.GetMachine(ctx, s.appName, registry.FlyMachineID)
	if err != nil {
		return nil, fmt.Errorf("failed to get machine status from Fly.io: %w", err)
	}

	// Update registry with latest status
	registry.Status = machine.State
	registry.LastSeenAt = time.Now()

	return registry, nil
}

// StopUserMachine stops a user's machine
func (s *Service) StopUserMachine(ctx context.Context, userID string) error {
	registry, err := s.db.GetMachineByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get machine from registry: %w", err)
	}

	if registry.Status == "stopped" {
		slog.Info("Machine already stopped", "user_id", userID)
		return nil
	}

	if err := s.client.StopMachine(ctx, s.appName, registry.FlyMachineID); err != nil {
		return fmt.Errorf("failed to stop machine: %w", err)
	}

	// Update registry status
	registry.Status = "stopped"
	registry.LastSeenAt = time.Now()

	slog.Info("Machine stopped successfully", "user_id", userID, "machine_id", registry.FlyMachineID)
	return nil
}

// StartUserMachine starts a user's stopped machine
func (s *Service) StartUserMachine(ctx context.Context, userID string) (*models.MachineRegistry, error) {
	registry, err := s.db.GetMachineByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get machine from registry: %w", err)
	}

	if registry.Status == "running" {
		slog.Info("Machine already running", "user_id", userID)
		return registry, nil
	}

	machine, err := s.client.StartMachine(ctx, s.appName, registry.FlyMachineID)
	if err != nil {
		return nil, fmt.Errorf("failed to start machine: %w", err)
	}

	// Update registry with new status
	registry.Status = machine.State
	registry.LastSeenAt = time.Now()

	slog.Info("Machine started successfully", "user_id", userID, "machine_id", registry.FlyMachineID)
	return registry, nil
}
