package clients

import (
	"context"

	pb "github.com/LimeOnTop/interverse-contracts/question/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type QuestionClient struct {
	client pb.QuestionServiceClient
}

func NewQuestionClient(questionServiceURL string) *QuestionClient {
	conn, err := grpc.NewClient(questionServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("connect to question service: " + err.Error())
	}

	return &QuestionClient{
		client: pb.NewQuestionServiceClient(conn),
	}
}

func (c *QuestionClient) GetQuestions(ctx context.Context, page, limit int32) (*pb.GetQuestionsResponse, error) {
	req := &pb.GetQuestionsRequest{
		Pagination: &pb.Pagination{
			Page:  page,
			Limit: limit,
		},
	}
	return c.client.GetQuestions(ctx, req)
}

func (c *QuestionClient) GetQuestion(ctx context.Context, questionID string) (*pb.GetQuestionResponse, error) {
	req := &pb.GetQuestionRequest{
		QuestionId: questionID,
	}
	return c.client.GetQuestion(ctx, req)
}

func (c *QuestionClient) CreateQuestion(ctx context.Context, text, category, difficulty, technology string, tags []string, answer string, options []*pb.QuestionOption) (*pb.CreateQuestionResponse, error) {
	req := &pb.CreateQuestionRequest{
		Text:       text,
		Category:   category,
		Difficulty: difficulty,
		Technology: technology,
		Tags:       tags,
		Answer:     answer,
		Options:    options,
	}
	return c.client.CreateQuestion(ctx, req)
}

func (c *QuestionClient) UpdateQuestion(ctx context.Context, questionID, text, category, difficulty, technology string, tags []string, answer string, options []*pb.QuestionOption) (*pb.UpdateQuestionResponse, error) {
	req := &pb.UpdateQuestionRequest{
		QuestionId: questionID,
		Text:       text,
		Category:   category,
		Difficulty: difficulty,
		Technology: technology,
		Tags:       tags,
		Answer:     answer,
		Options:    options,
	}
	return c.client.UpdateQuestion(ctx, req)
}

func (c *QuestionClient) DeleteQuestion(ctx context.Context, questionID string) (*pb.Response, error) {
	req := &pb.DeleteQuestionRequest{
		QuestionId: questionID,
	}
	return c.client.DeleteQuestion(ctx, req)
}

func (c *QuestionClient) SearchQuestions(ctx context.Context, query string, page, limit int32) (*pb.SearchQuestionsResponse, error) {
	req := &pb.SearchQuestionsRequest{
		Query: query,
		Pagination: &pb.Pagination{
			Page:  page,
			Limit: limit,
		},
	}
	return c.client.SearchQuestions(ctx, req)
}

func (c *QuestionClient) GetQuestionsByTechnology(ctx context.Context, technology, difficulty string, page, limit int32) (*pb.GetQuestionsByTechnologyResponse, error) {
	req := &pb.GetQuestionsByTechnologyRequest{
		Technology: technology,
		Difficulty: difficulty,
		Pagination: &pb.Pagination{
			Page:  page,
			Limit: limit,
		},
	}
	return c.client.GetQuestionsByTechnology(ctx, req)
}
