package clients

import (
	"context"

	pb "github.com/LimeOnTop/interverse-contracts/interview/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type InterviewClient struct {
	client pb.InterviewServiceClient
}

func NewInterviewClient(interviewServiceURL string) *InterviewClient {
	conn, err := grpc.NewClient(interviewServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("connect to interview service: " + err.Error())
	}

	return &InterviewClient{
		client: pb.NewInterviewServiceClient(conn),
	}
}

func (c *InterviewClient) CreateInterview(ctx context.Context, userID, title, description, scheduledAt, level, specialization, subscriptionPlan string, technologies []string) (*pb.CreateInterviewResponse, error) {
	req := &pb.CreateInterviewRequest{
		UserId:           userID,
		Title:            title,
		Description:      description,
		ScheduledAt:      scheduledAt,
		Technologies:     technologies,
		Level:            level,
		Specialization:   specialization,
		SubscriptionPlan: subscriptionPlan,
	}
	return c.client.CreateInterview(ctx, req)
}

func (c *InterviewClient) GetInterviews(ctx context.Context, userID, status string, page, limit int32) (*pb.GetInterviewsResponse, error) {
	req := &pb.GetInterviewsRequest{
		UserId: userID,
		Status: status,
		Pagination: &pb.Pagination{
			Page:  page,
			Limit: limit,
		},
	}
	return c.client.GetInterviews(ctx, req)
}

func (c *InterviewClient) GetScheduledInterviews(ctx context.Context, userID, date string) (*pb.GetScheduledInterviewsResponse, error) {
	req := &pb.GetScheduledInterviewsRequest{
		UserId: userID,
		Date:   date,
	}
	return c.client.GetScheduledInterviews(ctx, req)
}

func (c *InterviewClient) GetInterview(ctx context.Context, interviewID string) (*pb.GetInterviewResponse, error) {
	req := &pb.GetInterviewRequest{
		InterviewId: interviewID,
	}
	return c.client.GetInterview(ctx, req)
}

func (c *InterviewClient) UpdateInterview(ctx context.Context, interviewID, title, description, status, scheduledAt, level, specialization string, technologies []string) (*pb.UpdateInterviewResponse, error) {
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
	return c.client.UpdateInterview(ctx, req)
}

func (c *InterviewClient) DeleteInterview(ctx context.Context, interviewID string) (*pb.Response, error) {
	req := &pb.DeleteInterviewRequest{
		InterviewId: interviewID,
	}
	return c.client.DeleteInterview(ctx, req)
}

func (c *InterviewClient) StartSession(ctx context.Context, interviewID, userID string) (*pb.StartSessionResponse, error) {
	req := &pb.StartSessionRequest{
		InterviewId: interviewID,
		UserId:      userID,
	}
	return c.client.StartSession(ctx, req)
}

func (c *InterviewClient) GetSessionContent(ctx context.Context, interviewID, userID string) (*pb.GetSessionContentResponse, error) {
	req := &pb.GetSessionContentRequest{
		InterviewId: interviewID,
		UserId:      userID,
	}
	return c.client.GetSessionContent(ctx, req)
}
