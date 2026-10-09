package usecase

import (
	"context"

	pb "github.com/LimeOnTop/interverse-contracts/question/gen"
)

// QuestionGateway is the backend port consumed by HTTP controllers.
type QuestionGateway interface {
	GetQuestions(ctx context.Context, page, limit int32) (*pb.GetQuestionsResponse, error)
	GetQuestion(ctx context.Context, questionID string) (*pb.GetQuestionResponse, error)
	CreateQuestion(ctx context.Context, text, category, difficulty, technology string, tags []string, answer string, options []*pb.QuestionOption) (*pb.CreateQuestionResponse, error)
	UpdateQuestion(ctx context.Context, questionID, text, category, difficulty, technology string, tags []string, answer string, options []*pb.QuestionOption) (*pb.UpdateQuestionResponse, error)
	DeleteQuestion(ctx context.Context, questionID string) (*pb.Response, error)
	SearchQuestions(ctx context.Context, query string, page, limit int32) (*pb.SearchQuestionsResponse, error)
	GetQuestionsByTechnology(ctx context.Context, technology, difficulty, category string, page, limit int32) (*pb.GetQuestionsByTechnologyResponse, error)
}
