package clients

import (
	"context"

	pb "github.com/LimeOnTop/interverse-contracts/user/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ProfileClient struct {
	client pb.UserProfileServiceClient
}

func NewProfileClient(profileServiceURL string) *ProfileClient {
	conn, err := grpc.NewClient(profileServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("connect to profile service: " + err.Error())
	}

	return &ProfileClient{
		client: pb.NewUserProfileServiceClient(conn),
	}
}

func (c *ProfileClient) GetProfile(ctx context.Context, userID string) (*pb.GetProfileResponse, error) {
	req := &pb.GetProfileRequest{UserId: userID}
	return c.client.GetProfile(ctx, req)
}

func (c *ProfileClient) UpdateProfile(
	ctx context.Context,
	userID, workExperience, avatarURL, aboutMe, higherEducation, englishLevel string,
) (*pb.UpdateProfileResponse, error) {
	req := &pb.UpdateProfileRequest{
		UserId:          userID,
		WorkExperience:  workExperience,
		AvatarUrl:       avatarURL,
		AboutMe:         aboutMe,
		HigherEducation: higherEducation,
		EnglishLevel:    englishLevel,
	}
	return c.client.UpdateProfile(ctx, req)
}
