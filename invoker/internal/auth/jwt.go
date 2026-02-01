package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// generateMachineJWT issues a machine-scoped JWT for a user.
func generateMachineJWT(jwtSecret, userID, machineID string) (string, error) {
	if jwtSecret == "" {
		return "", fmt.Errorf("SUPABASE_JWT_SECRET not configured")
	}

	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour)

	claims := MachineClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "invoker",
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
		UserID:    userID,
		MachineID: machineID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}
