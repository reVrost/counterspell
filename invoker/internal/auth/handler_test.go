package auth

import (
	"context"
	"testing"

	"github.com/revrost/invoker/internal/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGenerateUsername(t *testing.T) {
	tests := []struct {
		name      string
		firstName string
		email     string
		expected  string
	}{
		{
			name:      "simple name",
			firstName: "John",
			email:     "john@example.com",
			expected:  "john",
		},
		{
			name:      "name with special characters",
			firstName: "John-Paul",
			email:     "john@example.com",
			expected:  "johnpaul",
		},
		{
			name:      "short name",
			firstName: "Jo",
			email:     "johndoe@example.com",
			expected:  "joh",
		},
		{
			name:      "short name uses email",
			firstName: "",
			email:     "johndoe@example.com",
			expected:  "joh",
		},
		{
			name:      "long name truncated",
			firstName: "VeryLongFirstNameThatExceedsLimit",
			email:     "user@example.com",
			expected:  "verylongfirstnametha",
		},
		{
			name:      "uppercase to lowercase",
			firstName: "JOHN",
			email:     "john@example.com",
			expected:  "john",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateUsername(tt.firstName, tt.email)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEnsureUniqueUsername(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := db.NewMockRepository(ctrl)

	tests := []struct {
		name        string
		username    string
		setupMock   func()
		expected    string
		expectError bool
	}{
		{
			name:     "username already unique",
			username: "johndoe",
			setupMock: func() {
				mockDB.EXPECT().UsernameExists(gomock.Any(), "johndoe").Return(false, nil)
			},
			expected:    "johndoe",
			expectError: false,
		},
		{
			name:     "username exists - add number",
			username: "johndoe",
			setupMock: func() {
				mockDB.EXPECT().UsernameExists(gomock.Any(), "johndoe").Return(true, nil)
				mockDB.EXPECT().UsernameExists(gomock.Any(), "johndoe1").Return(false, nil)
			},
			expected:    "johndoe1",
			expectError: false,
		},
		{
			name:     "username exists - add higher number",
			username: "johndoe",
			setupMock: func() {
				mockDB.EXPECT().UsernameExists(gomock.Any(), "johndoe").Return(true, nil)
				mockDB.EXPECT().UsernameExists(gomock.Any(), "johndoe1").Return(true, nil)
				mockDB.EXPECT().UsernameExists(gomock.Any(), "johndoe2").Return(false, nil)
			},
			expected:    "johndoe2",
			expectError: false,
		},
		{
			name:     "database error",
			username: "johndoe",
			setupMock: func() {
				mockDB.EXPECT().UsernameExists(gomock.Any(), "johndoe").Return(false, assert.AnError)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			result, err := ensureUniqueUsername(context.Background(), mockDB, tt.username)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
