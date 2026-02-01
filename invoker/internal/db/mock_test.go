package db

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/revrost/invoker/pkg/models"
)

// MockDBQuerier is a test helper that implements Querier for testing
type MockDBQuerier struct {
	users      map[string]*models.User
	byEmail    map[string]*models.User
	byUsername map[string]*models.User
	machines   map[string]*models.MachineRegistry
	byUserID   map[string]*models.MachineRegistry
}

func NewMockDBQuerier() *MockDBQuerier {
	return &MockDBQuerier{
		users:      make(map[string]*models.User),
		byEmail:    make(map[string]*models.User),
		byUsername: make(map[string]*models.User),
		machines:   make(map[string]*models.MachineRegistry),
		byUserID:   make(map[string]*models.MachineRegistry),
	}
}

func (m *MockDBQuerier) CreateUser(ctx context.Context, user *models.User) error {
	if _, exists := m.users[user.ID]; exists {
		return assert.AnError
	}
	m.users[user.ID] = user
	m.byEmail[user.Email] = user
	m.byUsername[user.Username] = user
	return nil
}

func (m *MockDBQuerier) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	user, exists := m.users[id]
	if !exists {
		return nil, assert.AnError
	}
	return user, nil
}

func (m *MockDBQuerier) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user, exists := m.byEmail[email]
	if !exists {
		return nil, assert.AnError
	}
	return user, nil
}

func (m *MockDBQuerier) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	user, exists := m.byUsername[username]
	if !exists {
		return nil, assert.AnError
	}
	return user, nil
}

func (m *MockDBQuerier) UsernameExists(ctx context.Context, username string) (bool, error) {
	_, exists := m.byUsername[username]
	return exists, nil
}

func (m *MockDBQuerier) EmailExists(ctx context.Context, email string) (bool, error) {
	_, exists := m.byEmail[email]
	return exists, nil
}

func (m *MockDBQuerier) CreateMachine(ctx context.Context, machine *models.MachineRegistry) error {
	if _, exists := m.machines[machine.ID]; exists {
		return assert.AnError
	}
	m.machines[machine.ID] = machine
	m.byUserID[machine.UserID] = machine
	return nil
}

func (m *MockDBQuerier) GetMachineByUserID(ctx context.Context, userID string) (*models.MachineRegistry, error) {
	machine, exists := m.byUserID[userID]
	if !exists {
		return nil, nil
	}
	return machine, nil
}

func (m *MockDBQuerier) UpdateMachineStatus(ctx context.Context, machineID, status string, lastSeenAt time.Time, errorMessage string) error {
	machine, exists := m.machines[machineID]
	if !exists {
		return assert.AnError
	}
	machine.Status = status
	machine.LastSeenAt = lastSeenAt
	machine.ErrorMessage = errorMessage
	return nil
}

func TestMockDBQuerier_CreateUser(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDBQuerier()

	user := &models.User{
		ID:        uuid.New().String(),
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Tier:      "free",
	}

	err := mockDB.CreateUser(ctx, user)
	require.NoError(t, err)

	retrieved, err := mockDB.GetUserByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.Email, retrieved.Email)
	assert.Equal(t, user.FirstName, retrieved.FirstName)
	assert.Equal(t, user.LastName, retrieved.LastName)
}

func TestMockDBQuerier_EmailExists(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDBQuerier()

	user := &models.User{
		ID:        uuid.New().String(),
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Tier:      "free",
	}

	exists, err := mockDB.EmailExists(ctx, "test@example.com")
	require.NoError(t, err)
	assert.False(t, exists)

	err = mockDB.CreateUser(ctx, user)
	require.NoError(t, err)

	exists, err = mockDB.EmailExists(ctx, "test@example.com")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestMockDBQuerier_UsernameExists(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDBQuerier()

	user := &models.User{
		ID:        uuid.New().String(),
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Tier:      "free",
	}

	exists, err := mockDB.UsernameExists(ctx, "johndoe")
	require.NoError(t, err)
	assert.False(t, exists)

	err = mockDB.CreateUser(ctx, user)
	require.NoError(t, err)

	exists, err = mockDB.UsernameExists(ctx, "johndoe")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestMockDBQuerier_GetUserByEmail(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDBQuerier()

	user := &models.User{
		ID:        uuid.New().String(),
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Tier:      "free",
	}

	err := mockDB.CreateUser(ctx, user)
	require.NoError(t, err)

	retrieved, err := mockDB.GetUserByEmail(ctx, "test@example.com")
	require.NoError(t, err)
	assert.Equal(t, user.ID, retrieved.ID)
	assert.Equal(t, user.Email, retrieved.Email)
}

func TestMockDBQuerier_GetUserByUsername(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDBQuerier()

	user := &models.User{
		ID:        uuid.New().String(),
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Tier:      "free",
	}

	err := mockDB.CreateUser(ctx, user)
	require.NoError(t, err)

	retrieved, err := mockDB.GetUserByUsername(ctx, "johndoe")
	require.NoError(t, err)
	assert.Equal(t, user.ID, retrieved.ID)
	assert.Equal(t, user.Username, retrieved.Username)
}

func TestMockDBQuerier_GetUserByID_NotFound(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDBQuerier()

	_, err := mockDB.GetUserByID(ctx, "non-existent-id")
	assert.Error(t, err)
}

func TestMockDBQuerier_GetUserByEmail_NotFound(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDBQuerier()

	_, err := mockDB.GetUserByEmail(ctx, "non-existent@example.com")
	assert.Error(t, err)
}

func TestMockDBQuerier_GetUserByUsername_NotFound(t *testing.T) {
	ctx := context.Background()
	mockDB := NewMockDBQuerier()

	_, err := mockDB.GetUserByUsername(ctx, "nonexistent")
	assert.Error(t, err)
}
