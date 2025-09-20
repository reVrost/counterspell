package counterspell

import (
	"context"
	"database/sql"

	"connectrpc.com/connect"
	v1 "github.com/revrost/counterspell/gen/counterspell/v1"
	"github.com/revrost/counterspell/internal/db"
)

// ServiceHandler handles HTTP requests for the counterspell API
type Service struct {
	queries *db.Queries
	db      *sql.DB
}

// NewServiceHandler creates a new ServiceHandler
func NewService(database *sql.DB) *Service {
	return &Service{
		queries: db.New(database),
		db:      database,
	}
}

// CreateBlueprint handles POST /counterspell/api/blueprints

// GetBlueprint
func (m *Service) GetBlueprint(ctx context.Context, req *connect.Request[v1.GetBlueprintRequest]) (*connect.Response[v1.GetBlueprintResponse], error) {
}
