package counterspell

import (
	"context"
	"database/sql"

	"connectrpc.com/connect"
	"github.com/revrost/counterspell/pkg/db"
	v1 "github.com/revrost/counterspell/pkg/gen/proto/counterspell/v1"
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
	return nil, nil
}

func (m *Service) CreateBlueprint(ctx context.Context, req *connect.Request[v1.CreateBlueprintRequest]) (*connect.Response[v1.CreateBlueprintResponse], error) {
	return nil, nil
}

// GetBlueprints
func (m *Service) ListBlueprints(ctx context.Context, req *connect.Request[v1.ListBlueprintsRequest]) (*connect.Response[v1.ListBlueprintsResponse], error) {
	return nil, nil
}

func (m *Service) ListLogs(ctx context.Context, req *connect.Request[v1.ListLogsRequest]) (*connect.Response[v1.ListLogsResponse], error) {
	return nil, nil
}

func (m *Service) GetTrace(ctx context.Context, req *connect.Request[v1.GetTraceRequest]) (*connect.Response[v1.GetTraceResponse], error) {
	return nil, nil
}

func (m *Service) ListTraces(ctx context.Context, req *connect.Request[v1.ListTracesRequest]) (*connect.Response[v1.ListTracesResponse], error) {
	return nil, nil
}
