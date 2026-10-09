package usecase

import (
	"context"

	pb "github.com/LimeOnTop/interverse-contracts/vacancy/gen"
)

// VacancyGateway is the backend port consumed by HTTP controllers.
type VacancyGateway interface {
	GetVacanciesForUser(
		ctx context.Context,
		userID string,
		page, perPage int32,
		area string,
		sources []string,
	) (*pb.GetVacanciesForUserResponse, error)
	SearchVacancies(
		ctx context.Context,
		text, area string,
		page, perPage int32,
		sources []string,
	) (*pb.SearchVacanciesResponse, error)
}
