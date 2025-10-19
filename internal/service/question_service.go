package service

import (
	"context"
	"fmt"
	"strconv"

	common "github.com/inter-verse/services/proto/gen"
	pb "github.com/inter-verse/services/question-service/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type QuestionService struct {
	client pb.QuestionServiceClient
}

func NewQuestionService(questionServiceURL string) *QuestionService {
	conn, err := grpc.Dial(questionServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to question service: %v", err))
	}

	client := pb.NewQuestionServiceClient(conn)
	return &QuestionService{client: client}
}

func (s *QuestionService) GetQuestions(page, limit string) (interface{}, error) {
	pageInt, _ := strconv.ParseInt(page, 10, 32)
	limitInt, _ := strconv.ParseInt(limit, 10, 32)

	req := &pb.GetQuestionsRequest{
		Pagination: &common.Pagination{
			Page:  int32(pageInt),
			Limit: int32(limitInt),
		},
	}

	response, err := s.client.GetQuestions(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *QuestionService) GetQuestion(questionID string) (interface{}, error) {
	req := &pb.GetQuestionRequest{
		QuestionId: questionID,
	}

	response, err := s.client.GetQuestion(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *QuestionService) CreateQuestion(text, category, difficulty, technology string, tags []string, answer string) (interface{}, error) {
	req := &pb.CreateQuestionRequest{
		Text:       text,
		Category:   category,
		Difficulty: difficulty,
		Technology: technology,
		Tags:       tags,
		Answer:     answer,
	}

	response, err := s.client.CreateQuestion(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *QuestionService) UpdateQuestion(questionID, text, category, difficulty, technology string, tags []string, answer string) (interface{}, error) {
	req := &pb.UpdateQuestionRequest{
		QuestionId: questionID,
		Text:       text,
		Category:   category,
		Difficulty: difficulty,
		Technology: technology,
		Tags:       tags,
		Answer:     answer,
	}

	response, err := s.client.UpdateQuestion(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *QuestionService) DeleteQuestion(questionID string) (interface{}, error) {
	req := &pb.DeleteQuestionRequest{
		QuestionId: questionID,
	}

	response, err := s.client.DeleteQuestion(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *QuestionService) SearchQuestions(query, page, limit string) (interface{}, error) {
	pageInt, _ := strconv.ParseInt(page, 10, 32)
	limitInt, _ := strconv.ParseInt(limit, 10, 32)

	req := &pb.SearchQuestionsRequest{
		Query: query,
		Pagination: &common.Pagination{
			Page:  int32(pageInt),
			Limit: int32(limitInt),
		},
	}

	response, err := s.client.SearchQuestions(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *QuestionService) GetQuestionsByTechnology(technology, difficulty, page, limit string) (interface{}, error) {
	pageInt, _ := strconv.ParseInt(page, 10, 32)
	limitInt, _ := strconv.ParseInt(limit, 10, 32)

	req := &pb.GetQuestionsByTechnologyRequest{
		Technology: technology,
		Difficulty: difficulty,
		Pagination: &common.Pagination{
			Page:  int32(pageInt),
			Limit: int32(limitInt),
		},
	}

	response, err := s.client.GetQuestionsByTechnology(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}
