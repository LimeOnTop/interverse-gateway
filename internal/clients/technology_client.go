package clients

import (
	"context"

	pb "github.com/LimeOnTop/interverse-contracts/technology/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TechnologyClient struct {
	client pb.TechnologyServiceClient
}

func NewTechnologyClient(technologyServiceURL string) *TechnologyClient {
	conn, err := grpc.NewClient(technologyServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("connect to technology service: " + err.Error())
	}

	return &TechnologyClient{
		client: pb.NewTechnologyServiceClient(conn),
	}
}

func (c *TechnologyClient) GetTechnologies(ctx context.Context, page, limit int32) (*pb.GetTechnologiesResponse, error) {
	req := &pb.GetTechnologiesRequest{
		Pagination: &pb.Pagination{
			Page:  page,
			Limit: limit,
		},
	}
	return c.client.GetTechnologies(ctx, req)
}

func (c *TechnologyClient) GetTechnology(ctx context.Context, technologyID string) (*pb.GetTechnologyResponse, error) {
	req := &pb.GetTechnologyRequest{
		TechnologyId: technologyID,
	}
	return c.client.GetTechnology(ctx, req)
}

func (c *TechnologyClient) CreateTechnology(ctx context.Context, name, category, description string, tags []string) (*pb.CreateTechnologyResponse, error) {
	req := &pb.CreateTechnologyRequest{
		Name:        name,
		Category:    category,
		Description: description,
		Tags:        tags,
	}
	return c.client.CreateTechnology(ctx, req)
}

func (c *TechnologyClient) UpdateTechnology(ctx context.Context, technologyID, name, category, description string, tags []string) (*pb.UpdateTechnologyResponse, error) {
	req := &pb.UpdateTechnologyRequest{
		TechnologyId: technologyID,
		Name:         name,
		Category:     category,
		Description:  description,
		Tags:         tags,
	}
	return c.client.UpdateTechnology(ctx, req)
}

func (c *TechnologyClient) DeleteTechnology(ctx context.Context, technologyID string) (*pb.Response, error) {
	req := &pb.DeleteTechnologyRequest{
		TechnologyId: technologyID,
	}
	return c.client.DeleteTechnology(ctx, req)
}

func (c *TechnologyClient) SearchTechnologies(ctx context.Context, query string, page, limit int32) (*pb.SearchTechnologiesResponse, error) {
	req := &pb.SearchTechnologiesRequest{
		Query: query,
		Pagination: &pb.Pagination{
			Page:  page,
			Limit: limit,
		},
	}
	return c.client.SearchTechnologies(ctx, req)
}
