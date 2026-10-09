package usecase

import (
	"context"

	pb "github.com/LimeOnTop/interverse-contracts/interview/gen"
)

// InterviewGateway is the backend port consumed by HTTP controllers.
type InterviewGateway interface {
	CreateInterview(ctx context.Context, userID, title, description, scheduledAt, level, specialization, subscriptionPlan string, technologies []string) (*pb.CreateInterviewResponse, error)
	GetInterviews(ctx context.Context, userID, status string, page, limit int32) (*pb.GetInterviewsResponse, error)
	GetScheduledInterviews(ctx context.Context, userID, date string) (*pb.GetScheduledInterviewsResponse, error)
	GetInterview(ctx context.Context, interviewID string) (*pb.GetInterviewResponse, error)
	UpdateInterview(ctx context.Context, interviewID, title, description, status, scheduledAt, level, specialization string, technologies []string) (*pb.UpdateInterviewResponse, error)
	DeleteInterview(ctx context.Context, interviewID string) (*pb.Response, error)
	StartSession(ctx context.Context, interviewID, userID string) (*pb.StartSessionResponse, error)
	GetSessionContent(ctx context.Context, interviewID, userID string) (*pb.GetSessionContentResponse, error)
	GetTrainingStats(ctx context.Context, userID, subscriptionPlan string) (*pb.GetTrainingStatsResponse, error)
}
