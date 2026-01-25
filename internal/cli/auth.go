package cli

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/revrost/code/counterspell/internal/services"
)

// AuthCLI handles authentication from CLI.
type AuthCLI struct {
	authService *services.AuthService
}

// NewAuthCLI creates a new auth CLI.
func NewAuthCLI(authService *services.AuthService) *AuthCLI {
	return &AuthCLI{
		authService: authService,
	}
}

// CheckAuth checks if user is authenticated, prompts if not.
func (a *AuthCLI) CheckAuth(ctx context.Context, machineID string) (string, *services.ExchangeCodeResponse, error) {
	// Check if already authenticated
	authenticated, err := a.authService.IsAuthenticated(ctx, machineID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to check auth status: %w", err)
	}

	if authenticated {
		auth, err := a.authService.GetStoredAuth(ctx, machineID)
		if err != nil {
			return "", nil, fmt.Errorf("failed to get stored auth: %w", err)
		}
		slog.Info("Already authenticated", "user", auth.Email, "user_id", auth.UserID)
		return auth.UserID, nil, nil
	}

	// Not authenticated - start auth flow
	return a.startAuthFlow(ctx, machineID)
}

// startAuthFlow starts the interactive auth flow.
func (a *AuthCLI) startAuthFlow(ctx context.Context, machineID string) (string, *services.ExchangeCodeResponse, error) {
	fmt.Println("\n" + "============================================================")
	fmt.Println("Welcome to Counterspell!")
	fmt.Println("============================================================")
	fmt.Println("\nYou need to authenticate to use Counterspell.")
	fmt.Println("This allows us to:")
	fmt.Println("  - Provide your personal subdomain (e.g., username.counterspell.app)")
	fmt.Println("  - Create a secure tunnel to your machine")
	fmt.Println("  - Manage your cloud deployments")
	fmt.Println()

	// Get auth URL
	authURL, state, err := a.authService.StartAuthFlow(ctx)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get auth URL: %w", err)
	}

	fmt.Printf("\n1. Visit this URL in your browser:\n\n")
	fmt.Printf("   %s\n\n", authURL)
	fmt.Printf("2. Log in or create an account\n")
	fmt.Printf("3. After logging in, you'll receive a code\n\n")

	// Prompt for code
	fmt.Print("Enter the code from the browser: ")
	reader := bufio.NewReader(os.Stdin)
	code, err := reader.ReadString('\n')
	if err != nil {
		return "", nil, fmt.Errorf("failed to read code: %w", err)
	}
	code = code[:len(code)-1] // Remove newline

	if code == "" {
		return "", nil, fmt.Errorf("code cannot be empty")
	}

	fmt.Println("\nAuthenticating...")

	// Exchange code for JWT
	resp, err := a.authService.CompleteAuthFlow(ctx, code, state)
	if err != nil {
		return "", nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	// Store auth
	if err := a.authService.StoreAuth(ctx, machineID, resp.JWT, resp.UserID, resp.Email, resp.ExpiresAt); err != nil {
		return "", nil, fmt.Errorf("failed to store auth: %w", err)
	}

	fmt.Printf("\n✓ Authenticated as %s\n", resp.Email)
	fmt.Printf("✓ Your user ID: %s\n\n", resp.UserID)

	return resp.UserID, resp, nil
}

// RegisterMachine registers the machine and returns the subdomain.
func (a *AuthCLI) RegisterMachine(ctx context.Context, machineID, userID string) (string, error) {
	// Get stored auth
	auth, err := a.authService.GetStoredAuth(ctx, machineID)
	if err != nil {
		return "", fmt.Errorf("failed to get auth: %w", err)
	}

	fmt.Println("Registering your machine...")

	subdomain, err := a.authService.RegisterMachine(ctx, auth.JwtToken, machineID)
	if err != nil {
		return "", fmt.Errorf("failed to register machine: %w", err)
	}

	fmt.Printf("✓ Machine registered!\n")
	fmt.Printf("✓ Your subdomain: %s.counterspell.app\n\n", subdomain)

	fmt.Println("You can now access Counterspell from:")
	fmt.Printf("  https://%s.counterspell.app\n\n", subdomain)
	fmt.Println("This is a secure tunnel to your local machine.")

	return subdomain, nil
}
