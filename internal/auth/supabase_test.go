package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetOAuthURL(t *testing.T) {
	t.Parallel()

	svc := &AuthService{supabaseURL: "https://example.supabase.co"}

	url, err := svc.GetOAuthURL("github", "http://localhost:8710/auth/callback")
	if err != nil {
		t.Fatalf("GetOAuthURL error: %v", err)
	}

	assertContains(t, url, "https://example.supabase.co/auth/v1/authorize")
	assertContains(t, url, "provider=github")
	assertContains(t, url, "redirect_to=http%3A%2F%2Flocalhost%3A8710%2Fauth%2Fcallback")
}

func TestSetSessionCookie(t *testing.T) {
	t.Parallel()

	svc := &AuthService{supabaseURL: "https://example.supabase.co"}

	rec := httptest.NewRecorder()
	svc.SetSessionCookie(rec, "test-token")

	cookie := findCookie(t, rec.Result().Cookies(), "sb-access-token")

	if cookie.Value != "test-token" {
		t.Errorf("cookie value = %q, want 'test-token'", cookie.Value)
	}
	if !cookie.HttpOnly {
		t.Error("cookie should be HttpOnly")
	}
}

func TestClearSessionCookies(t *testing.T) {
	t.Parallel()

	svc := &AuthService{supabaseURL: "https://example.supabase.co"}

	rec := httptest.NewRecorder()
	svc.ClearSessionCookies(rec)

	cookie := findCookie(t, rec.Result().Cookies(), "sb-access-token")

	if cookie.MaxAge != -1 {
		t.Errorf("MaxAge = %d, want -1 (delete)", cookie.MaxAge)
	}
}

// Test helpers

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("string missing %q\nGot: %s", substr, s)
	}
}

func findCookie(t *testing.T, cookies []*http.Cookie, name string) *http.Cookie {
	t.Helper()
	for _, c := range cookies {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("cookie %q not found", name)
	return nil
}
