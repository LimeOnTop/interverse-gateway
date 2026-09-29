package clients

import (
	"context"
	"time"

	pb "github.com/LimeOnTop/interverse-contracts/vacancy/gen"
)

type VacancyClient struct {
	client pb.VacancyServiceClient
}

func NewVacancyClient(vacancyServiceURL string) *VacancyClient {
	return &VacancyClient{
		client: pb.NewVacancyServiceClient(dialGRPCWithTimeout(vacancyServiceURL, "vacancy-service", 45*time.Second)),
	}
}

func (c *VacancyClient) GetVacanciesForUser(
	ctx context.Context,
	userID string,
	page, perPage int32,
	area string,
	sources []string,
) (*pb.GetVacanciesForUserResponse, error) {
	return c.client.GetVacanciesForUser(ctx, &pb.GetVacanciesForUserRequest{
		UserId:  userID,
		Page:    page,
		PerPage: perPage,
		Area:    area,
		Sources: sources,
	})
}

func (c *VacancyClient) SearchVacancies(
	ctx context.Context,
	text, area string,
	page, perPage int32,
	sources []string,
) (*pb.SearchVacanciesResponse, error) {
	return c.client.SearchVacancies(ctx, &pb.SearchVacanciesRequest{
		Text:    text,
		Area:    area,
		Page:    page,
		PerPage: perPage,
		Sources: sources,
	})
}
