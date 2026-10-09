package usecase

import (
	"context"

	pb "github.com/LimeOnTop/interverse-contracts/user/gen"
)

// SupportGateway is the backend port consumed by HTTP controllers.
type SupportGateway interface {
	CreateTicket(ctx context.Context, userID, userEmail, subject, message string) (*pb.TicketResponse, error)
	ListTickets(ctx context.Context, userID, status string, limit, offset int32) (*pb.ListTicketsResponse, error)
	GetTicket(ctx context.Context, ticketID int64, userID string) (*pb.TicketResponse, error)
	ListMessages(ctx context.Context, ticketID int64, userID string, limit, offset int32) (*pb.ListMessagesResponse, error)
	AddMessage(ctx context.Context, ticketID int64, userID, authorRole, authorID, body string) (*pb.AddMessageResponse, error)
	CloseTicket(ctx context.Context, ticketID int64, userID, closedBy string) (*pb.TicketResponse, error)
}
