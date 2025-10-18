package service

import (
	"context"
	"fmt"

	common "github.com/inter-verse/services/api-gateway/proto"
	pb "github.com/inter-verse/services/api-gateway/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type InterviewService struct {
	client pb.InterviewServiceClient
}

func NewInterviewService(interviewServiceURL string) *InterviewService {
	conn, err := grpc.Dial(interviewServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to interview service: %v", err))
	}

	client := pb.NewInterviewServiceClient(conn)
	return &InterviewService{client: client}
}

func (s *InterviewService) CreateInterview(title, description, candidateID, specialization, level string, duration int, scheduledAt, techStack string) (interface{}, error) {
	req := &pb.CreateInterviewRequest{
		CandidateId:    candidateID,
		InterviewerId:  "550e8400-e29b-41d4-a716-446655440000", // TODO: Get from auth context
		Title:          title,
		Description:    description,
		ScheduledAt:    scheduledAt,
		Technologies:   []string{techStack}, // Convert techStack to slice
		Level:          level,
		Specialization: specialization,
	}

	response, err := s.client.CreateInterview(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *InterviewService) GetInterviews(page, limit, status string) (interface{}, error) {
	req := &pb.GetInterviewsRequest{
		InterviewerId: "550e8400-e29b-41d4-a716-446655440000", // TODO: Get from auth context
		Pagination: &common.Pagination{
			Page:  1,  // TODO: parse page
			Limit: 10, // TODO: parse limit
		},
	}

	response, err := s.client.GetInterviews(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *InterviewService) GetScheduledInterviews(page, limit string) (interface{}, error) {
	req := &pb.GetScheduledInterviewsRequest{
		InterviewerId: "550e8400-e29b-41d4-a716-446655440000", // TODO: Get from auth context
		Date:          "2025-10-18",                           // TODO: Get current date or specific date
	}

	response, err := s.client.GetScheduledInterviews(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}
