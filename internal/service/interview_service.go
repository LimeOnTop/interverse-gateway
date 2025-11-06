package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	pb "github.com/inter-verse/interview-service/gen"
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
	fmt.Printf("DEBUG: InterviewService.CreateInterview called with candidateID=%s, title=%s\n", candidateID, title)

	// Parse tech_stack JSON string to array
	var technologies []string
	if techStack != "" {
		if err := json.Unmarshal([]byte(techStack), &technologies); err != nil {
			// If parsing fails, treat as single value
			technologies = []string{techStack}
		}
	}

	req := &pb.CreateInterviewRequest{
		CandidateId:    candidateID,
		InterviewerId:  "550e8400-e29b-41d4-a716-446655440000", // TODO: Get from auth context
		Title:          title,
		Description:    description,
		ScheduledAt:    scheduledAt,
		Technologies:   technologies,
		Level:          level,
		Specialization: specialization,
	}

	fmt.Printf("DEBUG: Sending gRPC request with CandidateId=%s, InterviewerId=%s\n", req.CandidateId, req.InterviewerId)

	response, err := s.client.CreateInterview(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *InterviewService) GetInterviews(page, limit, status, level string) (interface{}, error) {
	// Parse page and limit
	pageInt := 1
	limitInt := 10

	if page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			pageInt = p
		}
	}

	if limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			limitInt = l
		}
	}

	req := &pb.GetInterviewsRequest{
		InterviewerId: "550e8400-e29b-41d4-a716-446655440000", // TODO: Get from auth context
		Pagination: &pb.Pagination{
			Page:  int32(pageInt),
			Limit: int32(limitInt),
		},
	}

	response, err := s.client.GetInterviews(context.Background(), req)
	if err != nil {
		return nil, err
	}

	// Filter by status and/or level if provided
	if (status != "" || level != "") && response.Interviews != nil {
		filteredInterviews := []*pb.Interview{}
		for _, interview := range response.Interviews {
			statusMatch := status == "" || interview.Status == status
			levelMatch := level == "" || interview.Level == level

			if statusMatch && levelMatch {
				filteredInterviews = append(filteredInterviews, interview)
			}
		}
		response.Interviews = filteredInterviews
		// Update pagination total
		if response.Pagination != nil {
			response.Pagination.Total = int32(len(filteredInterviews))
		}
	}

	return response, nil
}

func (s *InterviewService) GetScheduledInterviews(page, limit string) (interface{}, error) {
	// Use current date by default
	date := time.Now().Format("2006-01-02")

	req := &pb.GetScheduledInterviewsRequest{
		InterviewerId: "550e8400-e29b-41d4-a716-446655440000", // TODO: Get from auth context
		Date:          date,
	}

	response, err := s.client.GetScheduledInterviews(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *InterviewService) GetInterview(interviewID string) (interface{}, error) {
	req := &pb.GetInterviewRequest{
		InterviewId: interviewID,
	}

	response, err := s.client.GetInterview(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *InterviewService) UpdateInterview(interviewID, title, description, status, scheduledAt, specialization, level, techStack string) (interface{}, error) {
	// Parse tech_stack JSON string to array
	var technologies []string
	if techStack != "" {
		if err := json.Unmarshal([]byte(techStack), &technologies); err != nil {
			// If parsing fails, treat as single value
			technologies = []string{techStack}
		}
	}

	req := &pb.UpdateInterviewRequest{
		InterviewId:    interviewID,
		Title:          title,
		Description:    description,
		Status:         status,
		ScheduledAt:    scheduledAt,
		Technologies:   technologies,
		Level:          level,
		Specialization: specialization,
	}

	response, err := s.client.UpdateInterview(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}
