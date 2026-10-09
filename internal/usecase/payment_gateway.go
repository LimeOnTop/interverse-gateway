package usecase

import (
	"context"

	pb "github.com/LimeOnTop/interverse-contracts/payment/gen"
)

// PaymentGateway is the backend port consumed by HTTP controllers.
type PaymentGateway interface {
	CreatePayment(ctx context.Context, userID, plan, email string) (*pb.CreatePaymentResponse, error)
	GetOffers(ctx context.Context) (*pb.GetOffersResponse, error)
	ConfirmResult(
		ctx context.Context,
		outSum string,
		invID int64,
		signature string,
		shp map[string]string,
	) (*pb.ConfirmResultResponse, error)
	GetPaymentStats(ctx context.Context, periods *pb.PeriodBoundaries) (*pb.GetPaymentStatsResponse, error)
	ListUserPayments(ctx context.Context, userID string) (*pb.ListUserPaymentsResponse, error)
}
