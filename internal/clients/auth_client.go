package clients

import (
	"context"

	pb "github.com/LimeOnTop/interverse-contracts/auth/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthClient struct {
	client pb.AuthServiceClient
}

func NewAuthClient(authServiceURL string) *AuthClient {
	conn, err := grpc.NewClient(authServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("connect to auth service: " + err.Error())
	}

	return &AuthClient{
		client: pb.NewAuthServiceClient(conn),
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

func (c *AuthClient) ValidateToken(ctx context.Context, token string) (*pb.ValidateTokenResponse, error) {
	req := &pb.ValidateTokenRequest{
		Token: token,
	}
	return c.client.ValidateToken(ctx, req)
}

func (c *AuthClient) RefreshToken(ctx context.Context, refreshToken string) (*pb.RefreshTokenResponse, error) {
	req := &pb.RefreshTokenRequest{
		RefreshToken: refreshToken,
	}
	return c.client.RefreshToken(ctx, req)
}

func (c *AuthClient) Logout(ctx context.Context, token string) (*pb.Response, error) {
	req := &pb.LogoutRequest{
		Token: token,
	}
	return c.client.Logout(ctx, req)
}
