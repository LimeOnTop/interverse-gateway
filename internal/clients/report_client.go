package clients

import (
	"context"
	"time"

	pb "github.com/LimeOnTop/interverse-contracts/report/gen"
	"github.com/LimeOnTop/interverse-gateway/internal/usecase"
	"google.golang.org/grpc/metadata"
)

type ReportClient struct {
	client pb.ReportServiceClient
}

func NewReportClient(reportServiceURL string) *ReportClient {
	return &ReportClient{
		client: pb.NewReportServiceClient(dialGRPCWithTimeout(reportServiceURL, "report-service", 120*time.Second)),
	}
}

func (c *ReportClient) GenerateReport(
	ctx context.Context,
	interviewID, userID string,
	answers []*pb.SessionAnswerInput,
	subscriptionPlan string,
) (*pb.GenerateReportResponse, error) {
	req := &pb.GenerateReportRequest{
		InterviewId: interviewID,
		UserId:      userID,
		Answers:     answers,
	}
	// Basic ("free") reports are built without the LLM.
	ctx = metadata.AppendToOutgoingContext(ctx, "x-subscription-plan", subscriptionPlan)
	return c.client.GenerateReport(ctx, req)
}

func (c *ReportClient) CreateReport(ctx context.Context, interviewID, userID, overallRating, technicalSkills, communicationSkills, problemSolving, strengths, weaknesses, recommendations, notes string) (*pb.CreateReportResponse, error) {
	req := &pb.CreateReportRequest{
		InterviewId:         interviewID,
		UserId:              userID,
		OverallRating:       overallRating,
		TechnicalSkills:     technicalSkills,
		CommunicationSkills: communicationSkills,
		ProblemSolving:      problemSolving,
		Strengths:           strengths,
		Weaknesses:          weaknesses,
		Recommendations:     recommendations,
		Notes:               notes,
	}
	return c.client.CreateReport(ctx, req)
}

func (c *ReportClient) GetReports(ctx context.Context, userID string, page, limit int32) (*pb.GetReportsResponse, error) {
	req := &pb.GetReportsRequest{
		UserId: userID,
		Pagination: &pb.Pagination{
			Page:  page,
			Limit: limit,
		},
	}
	return c.client.GetReports(ctx, req)
}

func (c *ReportClient) GetReport(ctx context.Context, reportID string) (*pb.GetReportResponse, error) {
	req := &pb.GetReportRequest{
		ReportId: reportID,
	}
	return c.client.GetReport(ctx, req)
}

func (c *ReportClient) UpdateReport(ctx context.Context, reportID, overallRating, technicalSkills, communicationSkills, problemSolving, strengths, weaknesses, recommendations, notes string) (*pb.UpdateReportResponse, error) {
	req := &pb.UpdateReportRequest{
		ReportId:            reportID,
		OverallRating:       overallRating,
		TechnicalSkills:     technicalSkills,
		CommunicationSkills: communicationSkills,
		ProblemSolving:      problemSolving,
		Strengths:           strengths,
		Weaknesses:          weaknesses,
		Recommendations:     recommendations,
		Notes:               notes,
	}
	return c.client.UpdateReport(ctx, req)
}

func (c *ReportClient) DeleteReport(ctx context.Context, reportID string) (*pb.Response, error) {
	req := &pb.DeleteReportRequest{
		ReportId: reportID,
	}
	return c.client.DeleteReport(ctx, req)
}

var _ usecase.ReportGateway = (*ReportClient)(nil)
