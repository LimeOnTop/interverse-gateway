package service

import (
	"context"
	"fmt"
	"strconv"

	pb "github.com/LimeOnTop/interverse-candidate/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type CandidateService struct {
	client pb.CandidateServiceClient
}

func NewCandidateService(candidateServiceURL string) *CandidateService {
	conn, err := grpc.Dial(candidateServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to candidate service: %v", err))
	}

	client := pb.NewCandidateServiceClient(conn)
	return &CandidateService{client: client}
}

func (s *CandidateService) CreateCandidate(name, email, phone, position, experience, skills, resumeURL, linkedinURL, githubURL, interviewerID string) (interface{}, error) {
	req := &pb.CreateCandidateRequest{
		Name:          name,
		Email:         email,
		Phone:         phone,
		Position:      position,
		Experience:    experience,
		Skills:        skills,
		ResumeUrl:     resumeURL,
		LinkedinUrl:   linkedinURL,
		GithubUrl:     githubURL,
		InterviewerId: interviewerID,
	}

	response, err := s.client.CreateCandidate(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *CandidateService) GetCandidates(interviewerID, page, limit string) (interface{}, error) {
	pageInt, _ := strconv.ParseInt(page, 10, 32)
	limitInt, _ := strconv.ParseInt(limit, 10, 32)

	req := &pb.GetCandidatesRequest{
		InterviewerId: interviewerID,
		Pagination: &pb.Pagination{
			Page:  int32(pageInt),
			Limit: int32(limitInt),
		},
	}

	response, err := s.client.GetCandidates(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *CandidateService) GetCandidate(candidateID string) (interface{}, error) {
	req := &pb.GetCandidateRequest{
		CandidateId: candidateID,
	}

	response, err := s.client.GetCandidate(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *CandidateService) UpdateCandidate(candidateID, name, email, phone, position, experience, skills, resumeURL, linkedinURL, githubURL, status string) (interface{}, error) {
	req := &pb.UpdateCandidateRequest{
		CandidateId: candidateID,
		Name:        name,
		Email:       email,
		Phone:       phone,
		Position:    position,
		Experience:  experience,
		Skills:      skills,
		ResumeUrl:   resumeURL,
		LinkedinUrl: linkedinURL,
		GithubUrl:   githubURL,
		Status:      status,
	}

	response, err := s.client.UpdateCandidate(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *CandidateService) DeleteCandidate(candidateID string) (interface{}, error) {
	req := &pb.DeleteCandidateRequest{
		CandidateId: candidateID,
	}

	response, err := s.client.DeleteCandidate(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *CandidateService) SearchCandidates(query, interviewerID, page, limit string) (interface{}, error) {
	pageInt, _ := strconv.ParseInt(page, 10, 32)
	limitInt, _ := strconv.ParseInt(limit, 10, 32)

	req := &pb.SearchCandidatesRequest{
		Query:         query,
		InterviewerId: interviewerID,
		Pagination: &pb.Pagination{
			Page:  int32(pageInt),
			Limit: int32(limitInt),
		},
	}

	response, err := s.client.SearchCandidates(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}
