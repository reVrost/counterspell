package fly

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// Client is a Fly.io API client
type Client struct {
	apiToken  string
	apiURL    string
	orgSlug   string
	userAgent string
}

// NewClient creates a new Fly.io API client
func NewClient(apiToken, orgSlug string) *Client {
	return &Client{
		apiToken:  apiToken,
		apiURL:    "https://api.machines.dev/v1", // Fly.io Machines API
		orgSlug:   orgSlug,
		userAgent: "invoker/1.0",
	}
}

// doRequest performs an authenticated HTTP request to Fly.io API
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, c.apiURL+path, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("API error: status %d, failed to decode error", resp.StatusCode)
		}
		return nil, fmt.Errorf("API error: status %d, message: %s", resp.StatusCode, errResp.Message)
	}

	return resp, nil
}

// CreateMachineConfig is the configuration for creating a Fly.io machine
type CreateMachineConfig struct {
	AppName      string            `json:"app_name"`
	Name         string            `json:"name"`           // e.g., "counterspell-{user_id}"
	Region       string            `json:"region"`         // e.g., "iad"
	Image        string            `json:"image"`          // Docker image
	Env          map[string]string `json:"env"`           // Environment variables
	VMSize       string            `json:"vm_size"`        // e.g., "shared-cpu-1x"
	MemoryMB     int               `json:"memory_mb"`      // e.g., 1024
	InternalPort int               `json:"internal_port"`   // e.g., 8080
	VolumeID     string            `json:"volume_id"`      // Optional: pre-existing volume ID
}

// CreateMachineRequest is the request body for creating a machine
type CreateMachineRequest struct {
	Name   string            `json:"name"`
	Region string            `json:"region"`
	Config MachineConfig     `json:"config"`
	Mounts []MachineMount   `json:"mounts,omitempty"`
}

// MachineConfig is the configuration for a machine
type MachineConfig struct {
	Image    string            `json:"image"`
	VM       VMConfig          `json:"vm"`
	Env      map[string]string `json:"env,omitempty"`
	Services []ServiceConfig   `json:"services,omitempty"`
}

// VMConfig defines VM resources
type VMConfig struct {
	Size      string `json:"size"`      // e.g., "shared-cpu-1x"
	MemoryMB  int    `json:"memory_mb"`
	CPUKind   string `json:"cpu_kind"`  // e.g., "shared"
	CPUs      int    `json:"cpus"`
}

// ServiceConfig defines how services are exposed
type ServiceConfig struct {
	Protocol     string      `json:"protocol"`      // e.g., "tcp"
	InternalPort int         `json:"internal_port"`
	Ports        []PortConfig `json:"ports,omitempty"`
}

// PortConfig defines port exposure
type PortConfig struct {
	Port     int      `json:"port"`
	Handlers []string `json:"handlers"` // e.g., ["http", "tls"]
}

// MachineMount defines volume mounts
type MachineMount struct {
	Volume string `json:"volume"`
	Path   string `json:"path"`
}

// Machine represents a Fly.io machine
type Machine struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	State       string    `json:"state"`       // "starting", "running", "stopped", "error"
	Region      string    `json:"region"`
	PrivateIP   string    `json:"private_ip"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Config     MachineConfig `json:"config"`
}

// ErrorResponse represents an API error
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// CreateMachine creates a new Fly.io machine
func (c *Client) CreateMachine(ctx context.Context, cfg *CreateMachineConfig) (*Machine, error) {
	req := &CreateMachineRequest{
		Name:   cfg.Name,
		Region: cfg.Region,
		Config: MachineConfig{
			Image: cfg.Image,
			VM: VMConfig{
				Size:     cfg.VMSize,
				MemoryMB: cfg.MemoryMB,
				CPUKind:  "shared",
				CPUs:     1,
			},
			Env: cfg.Env,
			Services: []ServiceConfig{
				{
					Protocol:     "tcp",
					InternalPort: cfg.InternalPort,
					Ports: []PortConfig{
						{
							Port:     80,
							Handlers: []string{"http"},
						},
						{
							Port:     443,
							Handlers: []string{"tls", "http"},
						},
					},
				},
			},
		},
	}

	// Add volume mount if volume ID is provided
	if cfg.VolumeID != "" {
		req.Mounts = []MachineMount{
			{
				Volume: cfg.VolumeID,
				Path:   "/data",
			},
		}
	}

	slog.Info("Creating Fly.io machine",
		"app_name", cfg.AppName,
		"name", cfg.Name,
		"region", cfg.Region,
		"image", cfg.Image,
		"vm_size", cfg.VMSize,
	)

	resp, err := c.doRequest(ctx, "POST", "/apps/"+cfg.AppName+"/machines", req)
	if err != nil {
		return nil, fmt.Errorf("failed to create machine: %w", err)
	}
	defer resp.Body.Close()

	var machine Machine
	if err := json.NewDecoder(resp.Body).Decode(&machine); err != nil {
		return nil, fmt.Errorf("failed to decode machine response: %w", err)
	}

	slog.Info("Fly.io machine created successfully",
		"machine_id", machine.ID,
		"name", machine.Name,
		"state", machine.State,
		"region", machine.Region,
	)

	return &machine, nil
}

// GetMachine retrieves information about a specific machine
func (c *Client) GetMachine(ctx context.Context, appName, machineID string) (*Machine, error) {
	resp, err := c.doRequest(ctx, "GET", "/apps/"+appName+"/machines/"+machineID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get machine: %w", err)
	}
	defer resp.Body.Close()

	var machine Machine
	if err := json.NewDecoder(resp.Body).Decode(&machine); err != nil {
		return nil, fmt.Errorf("failed to decode machine response: %w", err)
	}

	return &machine, nil
}

// StopMachine stops a running machine
func (c *Client) StopMachine(ctx context.Context, appName, machineID string) error {
	resp, err := c.doRequest(ctx, "POST", "/apps/"+appName+"/machines/"+machineID+"/stop", nil)
	if err != nil {
		return fmt.Errorf("failed to stop machine: %w", err)
	}
	defer resp.Body.Close()

	slog.Info("Fly.io machine stopped successfully", "machine_id", machineID, "app_name", appName)
	return nil
}

// StartMachine starts a stopped machine
func (c *Client) StartMachine(ctx context.Context, appName, machineID string) (*Machine, error) {
	resp, err := c.doRequest(ctx, "POST", "/apps/"+appName+"/machines/"+machineID+"/start", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to start machine: %w", err)
	}
	defer resp.Body.Close()

	var machine Machine
	if err := json.NewDecoder(resp.Body).Decode(&machine); err != nil {
		return nil, fmt.Errorf("failed to decode machine response: %w", err)
	}

	slog.Info("Fly.io machine started successfully", "machine_id", machineID, "app_name", appName)
	return &machine, nil
}

// ListMachines lists all machines in an app
func (c *Client) ListMachines(ctx context.Context, appName string) ([]Machine, error) {
	resp, err := c.doRequest(ctx, "GET", "/apps/"+appName+"/machines", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list machines: %w", err)
	}
	defer resp.Body.Close()

	var machines []Machine
	if err := json.NewDecoder(resp.Body).Decode(&machines); err != nil {
		return nil, fmt.Errorf("failed to decode machines response: %w", err)
	}

	return machines, nil
}

// GetPublicURL generates the public URL for a machine
func GetPublicURL(appName string) string {
	return fmt.Sprintf("https://%s.fly.dev", appName)
}
