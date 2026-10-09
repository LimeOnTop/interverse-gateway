package usecase

import (
	"context"

	pb "github.com/LimeOnTop/interverse-contracts/auth/gen"
)

// AuthGateway is the backend port consumed by HTTP controllers.
type AuthGateway interface {
	ListUsers(ctx context.Context, request *pb.ListUsersRequest) (*pb.ListUsersResponse, error)
	Register(ctx context.Context, name, email, password string) (*pb.RegisterResponse, error)
	Login(ctx context.Context, email, password string) (*pb.LoginResponse, error)
	GetUser(ctx context.Context, id string) (*pb.GetUserResponse, error)
	UpdateUser(ctx context.Context, id, name, email, password, role string) (*pb.UpdateUserResponse, error)
	DeleteUser(ctx context.Context, id string) (*pb.Response, error)
	Logout(ctx context.Context, token string) (*pb.Response, error)
	RefreshToken(ctx context.Context, refreshToken string) (*pb.RefreshTokenResponse, error)
	OAuthLogin(ctx context.Context, email, name, provider string) (*pb.OAuthLoginResponse, error)
	SendEmailVerification(ctx context.Context, email string) (*pb.SendEmailVerificationResponse, error)
	VerifyEmail(ctx context.Context, email, code, password string) (*pb.VerifyEmailResponse, error)
	GetRegistrationStats(ctx context.Context, periods *pb.PeriodBoundaries) (*pb.GetRegistrationStatsResponse, error)
}
