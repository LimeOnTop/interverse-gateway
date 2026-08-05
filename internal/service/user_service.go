package service

import (
	"context"
	"fmt"

	pb "github.com/LimeOnTop/interverse-user/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserService struct {
	client pb.UserServiceClient
}

func NewUserService(userServiceURL string) *UserService {
	conn, err := grpc.Dial(userServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to user service: %v", err))
	}

	client := pb.NewUserServiceClient(conn)
	return &UserService{client: client}
}

func (s *UserService) Register(name, email, password string) (interface{}, error) {
	req := &pb.RegisterRequest{
		Name:     name,
		Email:    email,
		Password: password,
	}

	response, err := s.client.Register(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *UserService) Login(email, password string) (interface{}, error) {
	req := &pb.LoginRequest{
		Email:    email,
		Password: password,
	}

	response, err := s.client.Login(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *UserService) GetUser(userID string) (interface{}, error) {
	req := &pb.GetUserRequest{
		UserId: userID,
	}

	response, err := s.client.GetUser(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *UserService) UpdateUser(userID, name, email string) (interface{}, error) {
	req := &pb.UpdateUserRequest{
		UserId: userID,
		Name:   name,
		Email:  email,
	}

	response, err := s.client.UpdateUser(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *UserService) DeleteUser(userID string) (interface{}, error) {
	req := &pb.DeleteUserRequest{
		UserId: userID,
	}

	response, err := s.client.DeleteUser(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *UserService) ValidateToken(token string) (interface{}, error) {
	req := &pb.ValidateTokenRequest{
		Token: token,
	}

	response, err := s.client.ValidateToken(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *UserService) RefreshToken(refreshToken string) (interface{}, error) {
	req := &pb.RefreshTokenRequest{
		RefreshToken: refreshToken,
	}

	response, err := s.client.RefreshToken(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *UserService) Logout(token string) (interface{}, error) {
	req := &pb.LogoutRequest{
		Token: token,
	}

	response, err := s.client.Logout(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return response, nil
}
