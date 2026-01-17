package auth

import (
	"testing"
)

func TestExtractBearerToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		header string
		want   string
	}{
		{"valid bearer", "Bearer abc123", "abc123"},
		{"empty", "", ""},
		{"no bearer prefix", "abc123", ""},
		{"lowercase bearer", "bearer abc123", ""},
		{"just Bearer", "Bearer ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ExtractBearerToken(tt.header)
			if got != tt.want {
				t.Errorf("ExtractBearerToken(%q) = %q, want %q", tt.header, got, tt.want)
			}
		})
	}
}

func TestClaims_UserID(t *testing.T) {
	t.Parallel()

	claims := &Claims{}
	claims.Subject = "d29ad18a-8a2c-493c-850a-d1139c4dbd30"

	if got := claims.UserID(); got != "d29ad18a-8a2c-493c-850a-d1139c4dbd30" {
		t.Errorf("UserID() = %q, want UUID", got)
	}
}

func TestClaims_GithubUsername(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		metadata map[string]interface{}
		want     string
	}{
		{
			name:     "nil metadata",
			metadata: nil,
			want:     "",
		},
		{
			name:     "user_name present",
			metadata: map[string]interface{}{"user_name": "reVrost"},
			want:     "reVrost",
		},
		{
			name:     "preferred_username fallback",
			metadata: map[string]interface{}{"preferred_username": "reVrost"},
			want:     "reVrost",
		},
		{
			name:     "user_name takes priority",
			metadata: map[string]interface{}{"user_name": "primary", "preferred_username": "fallback"},
			want:     "primary",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			claims := &Claims{UserMetadata: tt.metadata}
			if got := claims.GithubUsername(); got != tt.want {
				t.Errorf("GithubUsername() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestClaims_AvatarURL(t *testing.T) {
	t.Parallel()

	t.Run("with avatar", func(t *testing.T) {
		t.Parallel()
		claims := &Claims{
			UserMetadata: map[string]interface{}{
				"avatar_url": "https://avatars.githubusercontent.com/u/1558599?v=4",
			},
		}
		want := "https://avatars.githubusercontent.com/u/1558599?v=4"
		if got := claims.AvatarURL(); got != want {
			t.Errorf("AvatarURL() = %q, want %q", got, want)
		}
	})

	t.Run("nil metadata", func(t *testing.T) {
		t.Parallel()
		claims := &Claims{UserMetadata: nil}
		if got := claims.AvatarURL(); got != "" {
			t.Errorf("AvatarURL() with nil metadata = %q, want empty", got)
		}
	})
}

func TestNewJWTValidatorWithURL(t *testing.T) {
	t.Parallel()

	v := NewJWTValidatorWithURL("https://example.supabase.co", "anon-key")

	if v.supabaseURL != "https://example.supabase.co" {
		t.Errorf("supabaseURL = %q, want https://example.supabase.co", v.supabaseURL)
	}
	if v.anonKey != "anon-key" {
		t.Errorf("anonKey = %q, want anon-key", v.anonKey)
	}
}
