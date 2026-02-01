# Testing Guide

This project uses Go's testing framework with support for mocking using `go.uber.org/mock` (gomock) and assertions using `github.com/stretchr/testify`.

## Setup

All testing dependencies are already installed in `go.mod`:

- `go.uber.org/mock` - Latest version of Uber's GoMock library for generating mocks
- `github.com/stretchr/testify` - Assertion library for Go tests

## Running Tests

### Run All Tests
```bash
make test
# or
go test ./...
```

### Run Specific Package Tests
```bash
go test ./internal/auth/...
go test ./internal/config/...
go test ./internal/db/...
go test ./pkg/models/...
```

### Run Tests with Verbose Output
```bash
go test -v ./...
```

### Run Tests with Coverage
```bash
make test-coverage
# or
go test -coverprofile=coverage.out -covermode=atomic ./...
go tool cover -html=coverage.out -o coverage.html
```

### Run Tests with Race Detection
```bash
go test -race ./...
```

## Generating Mocks

The project uses mockgen to generate mocks for interfaces:

```bash
make mock-gen
# or
go run go.uber.org/mock/mockgen@latest -source=internal/db/service.go -destination=internal/db/mock_service.go -package=db
```

### Current Mocks

- `internal/db/mock_service.go` - Mock for `Service` interface (user database operations)
- `internal/db/mock.go` - Mock for sqlc `Querier` interface

## Test Structure

### Example Test with Gomock

```go
func TestHandler_Register(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockDB := db.NewMockService(ctrl)
    handler := auth.NewHandler(nil, mockDB)

    // Set up expectations
    mockDB.EXPECT().EmailExists(gomock.Any(), "test@example.com").Return(false, nil)
    mockDB.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil)

    // Execute test
    // ... test code ...

    // Assertions
    assert.Equal(t, http.StatusCreated, w.Code)
}
```

### Test Files

- `internal/auth/handler_test.go` - Auth handler tests with mocks
- `internal/config/config_test.go` - Configuration loading tests
- `internal/db/mock_test.go` - Mock database implementation tests
- `pkg/models/models_test.go` - Model structure tests

## Best Practices

1. **Use table-driven tests** for multiple test cases with the same logic
2. **Always call `ctrl.Finish()`** to verify all mock expectations were met
3. **Use `gomock.Any()`** for arguments where the exact value doesn't matter
4. **Use specific assertions** from testify for better error messages
5. **Test both happy and error paths** to ensure robust error handling
6. **Clean up resources** in defer statements

## Writing New Tests

1. Identify the interface you want to mock
2. Generate mocks using `make mock-gen`
3. Create test file: `package_test.go`
4. Write test function: `func TestFunctionName(t *testing.T)`
5. Create gomock controller and mock instances
6. Set up mock expectations
7. Execute the code under test
8. Assert on results
9. Call `ctrl.Finish()` to verify all expectations

## Common Patterns

### Testing HTTP Handlers
```go
func TestHandler_Method(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockDB := db.NewMockService(ctrl)
    handler := NewHandler(nil, mockDB)

    body, _ := json.Marshal(request)
    req := httptest.NewRequest(http.MethodPost, "/endpoint", bytes.NewBuffer(body))
    w := httptest.NewRecorder()

    handler.Method(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
}
```

### Testing Database Operations
```go
func TestDB_CreateUser(t *testing.T) {
    mockDB := db.NewMockService(ctrl)

    user := &models.User{...}
    mockDB.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil)

    err := mockDB.CreateUser(context.Background(), user)

    assert.NoError(t, err)
}
```

## Resources

- [GoMock Documentation](https://github.com/uber-go/mock)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Go Testing Package](https://golang.org/pkg/testing/)
