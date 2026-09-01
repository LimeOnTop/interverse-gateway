package clients

import (
	"context"

	pb "github.com/LimeOnTop/interverse-contracts/report/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ReportClient struct {
	client pb.ReportServiceClient
}

func NewReportClient(reportServiceURL string) *ReportClient {
	conn, err := grpc.NewClient(reportServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("connect to report service: " + err.Error())
	}

	return &ReportClient{
		client: pb.NewReportServiceClient(conn),
	}
}

func (c *ReportClient) CreateReport(ctx context.Context, interviewID, candidateID, interviewerID, overallRating, technicalSkills, communicationSkills, problemSolving, strengths, weaknesses, recommendations, notes string) (*pb.CreateReportResponse, error) {
	req := &pb.CreateReportRequest{
		InterviewId:         interviewID,
		CandidateId:         candidateID,
		InterviewerId:       interviewerID,
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

func (c *ReportClient) GetReports(ctx context.Context, interviewerID string, page, limit int32) (*pb.GetReportsResponse, error) {
	req := &pb.GetReportsRequest{
		InterviewerId: interviewerID,
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
