package service

import (
	"context"
	"fmt"
	"strconv"

	pb "github.com/LimeOnTop/interverse-technology/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TechnologyService struct {
	client pb.TechnologyServiceClient
}

func NewTechnologyService(technologyServiceURL string) *TechnologyService {
	conn, err := grpc.Dial(technologyServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to technology service: %v", err))
	}

	client := pb.NewTechnologyServiceClient(conn)
	return &TechnologyService{client: client}
}

func (s *TechnologyService) GetTechnologies(page, limit string) (interface{}, error) {
	pageInt, _ := strconv.ParseInt(page, 10, 32)
	limitInt, _ := strconv.ParseInt(limit, 10, 32)

	req := &pb.GetTechnologiesRequest{
		Pagination: &pb.Pagination{
			Page:  int32(pageInt),
			Limit: int32(limitInt),
		},
	}

	response, err := s.client.GetTechnologies(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *TechnologyService) GetTechnology(technologyID string) (interface{}, error) {
	req := &pb.GetTechnologyRequest{
		TechnologyId: technologyID,
	}

	response, err := s.client.GetTechnology(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *TechnologyService) CreateTechnology(name, category, description string, tags []string) (interface{}, error) {
	req := &pb.CreateTechnologyRequest{
		Name:        name,
		Category:    category,
		Description: description,
		Tags:        tags,
	}

	response, err := s.client.CreateTechnology(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *TechnologyService) UpdateTechnology(technologyID, name, category, description string, tags []string) (interface{}, error) {
	req := &pb.UpdateTechnologyRequest{
		TechnologyId: technologyID,
		Name:         name,
		Category:     category,
		Description:  description,
		Tags:         tags,
	}

	response, err := s.client.UpdateTechnology(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *TechnologyService) DeleteTechnology(technologyID string) (interface{}, error) {
	req := &pb.DeleteTechnologyRequest{
		TechnologyId: technologyID,
	}

	response, err := s.client.DeleteTechnology(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *TechnologyService) SearchTechnologies(query, page, limit string) (interface{}, error) {
	pageInt, _ := strconv.ParseInt(page, 10, 32)
	limitInt, _ := strconv.ParseInt(limit, 10, 32)

	req := &pb.SearchTechnologiesRequest{
		Query: query,
		Pagination: &pb.Pagination{
			Page:  int32(pageInt),
			Limit: int32(limitInt),
		},
	}

	response, err := s.client.SearchTechnologies(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}
