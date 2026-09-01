package clients

import (
	"context"

	pb "github.com/LimeOnTop/interverse-contracts/candidate/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type CandidateClient struct {
	client pb.CandidateServiceClient
}

func NewCandidateClient(candidateServiceURL string) *CandidateClient {
	conn, err := grpc.NewClient(candidateServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("connect to candidate service: " + err.Error())
	}

	return &CandidateClient{
		client: pb.NewCandidateServiceClient(conn),
	}
}

func (c *CandidateClient) CreateCandidate(ctx context.Context, name, email, phone, position string, experience int64, skills, resumeURL, linkedinURL, githubURL, interviewerID string) (*pb.CreateCandidateResponse, error) {
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
	return c.client.CreateCandidate(ctx, req)
}

func (c *CandidateClient) GetCandidates(ctx context.Context, interviewerID string, page, limit int32) (*pb.GetCandidatesResponse, error) {
	req := &pb.GetCandidatesRequest{
		InterviewerId: interviewerID,
		Pagination: &pb.Pagination{
			Page:  page,
			Limit: limit,
		},
	}
	return c.client.GetCandidates(ctx, req)
}

func (c *CandidateClient) GetCandidate(ctx context.Context, candidateID string) (*pb.GetCandidateResponse, error) {
	req := &pb.GetCandidateRequest{
		CandidateId: candidateID,
	}
	return c.client.GetCandidate(ctx, req)
}

func (c *CandidateClient) UpdateCandidate(ctx context.Context, candidateID, name, email, phone, position string, experience int64, skills, resumeURL, linkedinURL, githubURL, status string) (*pb.UpdateCandidateResponse, error) {
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
	return c.client.UpdateCandidate(ctx, req)
}

func (c *CandidateClient) DeleteCandidate(ctx context.Context, candidateID string) (*pb.Response, error) {
	req := &pb.DeleteCandidateRequest{
		CandidateId: candidateID,
	}
	return c.client.DeleteCandidate(ctx, req)
}

func (c *CandidateClient) SearchCandidates(ctx context.Context, query, interviewerID string, page, limit int32) (*pb.SearchCandidatesResponse, error) {
	req := &pb.SearchCandidatesRequest{
		Query:         query,
		InterviewerId: interviewerID,
		Pagination: &pb.Pagination{
			Page:  page,
			Limit: limit,
		},
	}
	return c.client.SearchCandidates(ctx, req)
}
