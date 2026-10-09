package usecase

import (
	"context"

	pb "github.com/LimeOnTop/interverse-contracts/technology/gen"
)

// TechnologyGateway is the backend port consumed by HTTP controllers.
type TechnologyGateway interface {
	GetTechnologies(ctx context.Context, page, limit int32) (*pb.GetTechnologiesResponse, error)
	GetTechnology(ctx context.Context, technologyID string) (*pb.GetTechnologyResponse, error)
	CreateTechnology(ctx context.Context, name, category, description string, tags []string) (*pb.CreateTechnologyResponse, error)
	UpdateTechnology(ctx context.Context, technologyID, name, category, description string, tags []string) (*pb.UpdateTechnologyResponse, error)
	DeleteTechnology(ctx context.Context, technologyID string) (*pb.Response, error)
	SearchTechnologies(ctx context.Context, query string, page, limit int32) (*pb.SearchTechnologiesResponse, error)
}
