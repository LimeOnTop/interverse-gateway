package clients

import (
	"context"

	pb "github.com/LimeOnTop/interverse-contracts/user/gen"
	"github.com/LimeOnTop/interverse-gateway/internal/usecase"
)

type ProfileClient struct {
	client pb.UserProfileServiceClient
}

func NewProfileClient(profileServiceURL string) *ProfileClient {
	return &ProfileClient{
		client: pb.NewUserProfileServiceClient(dialGRPC(profileServiceURL, "user-service")),
	}
}

func (c *ProfileClient) GetProfile(ctx context.Context, userID string) (*pb.GetProfileResponse, error) {
	req := &pb.GetProfileRequest{UserId: userID}
	return c.client.GetProfile(ctx, req)
}

func (c *ProfileClient) UpdateProfile(
	ctx context.Context,
	userID, workExperience, avatarURL, aboutMe, higherEducation, englishLevel, skills string,
) (*pb.UpdateProfileResponse, error) {
	req := &pb.UpdateProfileRequest{
		UserId:          userID,
		WorkExperience:  workExperience,
		AvatarUrl:       avatarURL,
		AboutMe:         aboutMe,
		HigherEducation: higherEducation,
		EnglishLevel:    englishLevel,
		Skills:          skills,
	}
	return c.client.UpdateProfile(ctx, req)
}

var _ usecase.ProfileGateway = (*ProfileClient)(nil)
