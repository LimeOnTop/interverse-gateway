package usecase

import (
	"context"

	pb "github.com/LimeOnTop/interverse-contracts/user/gen"
)

// ProfileGateway is the backend port consumed by HTTP controllers.
type ProfileGateway interface {
	GetProfile(ctx context.Context, userID string) (*pb.GetProfileResponse, error)
	UpdateProfile(
		ctx context.Context,
		userID, workExperience, avatarURL, aboutMe, higherEducation, englishLevel, skills string,
	) (*pb.UpdateProfileResponse, error)
}
