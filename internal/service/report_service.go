package service

import (
	"context"
	"fmt"
	"strconv"

	pb "github.com/LimeOnTop/interverse-contracts/report/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ReportService struct {
	client pb.ReportServiceClient
}

func NewReportService(reportServiceURL string) *ReportService {
	conn, err := grpc.Dial(reportServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to report service: %v", err))
	}

	client := pb.NewReportServiceClient(conn)
	return &ReportService{client: client}
}

func (s *ReportService) CreateReport(interviewID, candidateID, interviewerID, overallRating, technicalSkills, communicationSkills, problemSolving, strengths, weaknesses, recommendations, notes string) (interface{}, error) {
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

	response, err := s.client.CreateReport(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *ReportService) GetReports(interviewerID, page, limit string) (interface{}, error) {
	pageInt, _ := strconv.ParseInt(page, 10, 32)
	limitInt, _ := strconv.ParseInt(limit, 10, 32)

	req := &pb.GetReportsRequest{
		InterviewerId: interviewerID,
		Pagination: &pb.Pagination{
			Page:  int32(pageInt),
			Limit: int32(limitInt),
		},
	}

	response, err := s.client.GetReports(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *ReportService) GetReport(reportID string) (interface{}, error) {
	req := &pb.GetReportRequest{
		ReportId: reportID,
	}

	response, err := s.client.GetReport(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *ReportService) UpdateReport(reportID, overallRating, technicalSkills, communicationSkills, problemSolving, strengths, weaknesses, recommendations, notes string) (interface{}, error) {
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

	response, err := s.client.UpdateReport(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *ReportService) DeleteReport(reportID string) (interface{}, error) {
	req := &pb.DeleteReportRequest{
		ReportId: reportID,
	}

	response, err := s.client.DeleteReport(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}
