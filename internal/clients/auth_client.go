package clients

import (
	"context"

	pb "github.com/LimeOnTop/interverse-contracts/auth/gen"
	"github.com/LimeOnTop/interverse-gateway/internal/usecase"
)

type AuthClient struct {
	client pb.AuthServiceClient
}

func NewAuthClient(authServiceURL string) *AuthClient {
	return &AuthClient{
		client: pb.NewAuthServiceClient(dialGRPC(authServiceURL, "auth-service")),
	}
}

func (c *AuthClient) Register(ctx context.Context, name, email, password string) (*pb.RegisterResponse, error) {
	req := &pb.RegisterRequest{
		Name:     name,
		Email:    email,
		Password: password,
	}
	return c.client.Register(ctx, req)
}

func (c *AuthClient) Login(ctx context.Context, email, password string) (*pb.LoginResponse, error) {
	req := &pb.LoginRequest{
		Email:    email,
		Password: password,
	}
	return c.client.Login(ctx, req)
}

func (c *AuthClient) GetUser(ctx context.Context, id string) (*pb.GetUserResponse, error) {
	req := &pb.GetUserRequest{
		Id: id,
	}
	return c.client.GetUser(ctx, req)
}

func (c *AuthClient) UpdateUser(ctx context.Context, id, name, email, password, role string) (*pb.UpdateUserResponse, error) {
	req := &pb.UpdateUserRequest{
		Id:       id,
		Name:     name,
		Email:    email,
		Password: password,
		Role:     role,
	}
	return c.client.UpdateUser(ctx, req)
}

func (c *AuthClient) DeleteUser(ctx context.Context, id string) (*pb.Response, error) {
	req := &pb.DeleteUserRequest{
		Id: id,
	}
	return c.client.DeleteUser(ctx, req)
}

func (c *AuthClient) Logout(ctx context.Context, token string) (*pb.Response, error) {
	req := &pb.LogoutRequest{
		Token: token,
	}
	return c.client.Logout(ctx, req)
}

func (c *AuthClient) RefreshToken(ctx context.Context, refreshToken string) (*pb.RefreshTokenResponse, error) {
	req := &pb.RefreshTokenRequest{
		RefreshToken: refreshToken,
	}
	return c.client.RefreshToken(ctx, req)
}

func (c *AuthClient) OAuthLogin(ctx context.Context, email, name, provider string) (*pb.OAuthLoginResponse, error) {
	return c.client.OAuthLogin(ctx, &pb.OAuthLoginRequest{
		Email:    email,
		Name:     name,
		Provider: provider,
	})
}

func (c *AuthClient) SendEmailVerification(ctx context.Context, email string) (*pb.SendEmailVerificationResponse, error) {
	return c.client.SendEmailVerification(ctx, &pb.SendEmailVerificationRequest{Email: email})
}

func (c *AuthClient) VerifyEmail(ctx context.Context, email, code, password string) (*pb.VerifyEmailResponse, error) {
	return c.client.VerifyEmail(ctx, &pb.VerifyEmailRequest{
		Email:    email,
		Code:     code,
		Password: password,
	})
}

func (c *AuthClient) GetRegistrationStats(ctx context.Context, periods *pb.PeriodBoundaries) (*pb.GetRegistrationStatsResponse, error) {
	return c.client.GetRegistrationStats(ctx, &pb.GetRegistrationStatsRequest{Periods: periods})
}

var _ usecase.AuthGateway = (*AuthClient)(nil)

func (c *AuthClient) ListUsers(ctx context.Context, request *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	return c.client.ListUsers(ctx, request)
}
