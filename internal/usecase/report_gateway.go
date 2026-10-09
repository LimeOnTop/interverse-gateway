package usecase

import (
	"context"

	pb "github.com/LimeOnTop/interverse-contracts/report/gen"
)

// ReportGateway is the backend port consumed by HTTP controllers.
type ReportGateway interface {
	GenerateReport(
		ctx context.Context,
		interviewID, userID string,
		answers []*pb.SessionAnswerInput,
		subscriptionPlan string,
	) (*pb.GenerateReportResponse, error)
	CreateReport(ctx context.Context, interviewID, userID, overallRating, technicalSkills, communicationSkills, problemSolving, strengths, weaknesses, recommendations, notes string) (*pb.CreateReportResponse, error)
	GetReports(ctx context.Context, userID string, page, limit int32) (*pb.GetReportsResponse, error)
	GetReport(ctx context.Context, reportID string) (*pb.GetReportResponse, error)
	UpdateReport(ctx context.Context, reportID, overallRating, technicalSkills, communicationSkills, problemSolving, strengths, weaknesses, recommendations, notes string) (*pb.UpdateReportResponse, error)
	DeleteReport(ctx context.Context, reportID string) (*pb.Response, error)
}
