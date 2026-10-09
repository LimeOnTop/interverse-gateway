package clients

import (
	"context"

	pb "github.com/LimeOnTop/interverse-contracts/user/gen"
	"github.com/LimeOnTop/interverse-gateway/internal/usecase"
)

// SupportClient talks to the support tickets API hosted by user-service.
type SupportClient struct {
	client pb.SupportServiceClient
}

func NewSupportClient(userServiceURL string) *SupportClient {
	return &SupportClient{
		client: pb.NewSupportServiceClient(dialGRPC(userServiceURL, "user-service (support)")),
	}
}

func (c *SupportClient) CreateTicket(ctx context.Context, userID, userEmail, subject, message string) (*pb.TicketResponse, error) {
	return c.client.CreateTicket(ctx, &pb.CreateTicketRequest{
		UserId:    userID,
		UserEmail: userEmail,
		Subject:   subject,
		Message:   message,
	})
}

func (c *SupportClient) ListTickets(ctx context.Context, userID, status string, limit, offset int32) (*pb.ListTicketsResponse, error) {
	return c.client.ListTickets(ctx, &pb.ListTicketsRequest{UserId: userID, Status: status, Limit: limit, Offset: offset})
}

func (c *SupportClient) GetTicket(ctx context.Context, ticketID int64, userID string) (*pb.TicketResponse, error) {
	return c.client.GetTicket(ctx, &pb.GetTicketRequest{TicketId: ticketID, UserId: userID})
}

func (c *SupportClient) ListMessages(ctx context.Context, ticketID int64, userID string, limit, offset int32) (*pb.ListMessagesResponse, error) {
	return c.client.ListMessages(ctx, &pb.ListMessagesRequest{TicketId: ticketID, UserId: userID, Limit: limit, Offset: offset})
}

func (c *SupportClient) AddMessage(ctx context.Context, ticketID int64, userID, authorRole, authorID, body string) (*pb.AddMessageResponse, error) {
	return c.client.AddMessage(ctx, &pb.AddMessageRequest{
		TicketId:   ticketID,
		UserId:     userID,
		AuthorRole: authorRole,
		AuthorId:   authorID,
		Body:       body,
	})
}

func (c *SupportClient) CloseTicket(ctx context.Context, ticketID int64, userID, closedBy string) (*pb.TicketResponse, error) {
	return c.client.CloseTicket(ctx, &pb.CloseTicketRequest{TicketId: ticketID, UserId: userID, ClosedBy: closedBy})
}

var _ usecase.SupportGateway = (*SupportClient)(nil)
