package clients

import (
	"context"

	pb "github.com/LimeOnTop/interverse-contracts/payment/gen"
)

type PaymentClient struct {
	client pb.PaymentServiceClient
}

func NewPaymentClient(paymentServiceURL string) *PaymentClient {
	return &PaymentClient{
		client: pb.NewPaymentServiceClient(dialGRPC(paymentServiceURL, "payment-service")),
	}
}

func (c *PaymentClient) CreatePayment(ctx context.Context, userID, plan, email string) (*pb.CreatePaymentResponse, error) {
	return c.client.CreatePayment(ctx, &pb.CreatePaymentRequest{
		UserId: userID,
		Plan:   plan,
		Email:  email,
	})
}

func (c *PaymentClient) GetOffers(ctx context.Context) (*pb.GetOffersResponse, error) {
	return c.client.GetOffers(ctx, &pb.GetOffersRequest{})
}

func (c *PaymentClient) ConfirmResult(
	ctx context.Context,
	outSum string,
	invID int64,
	signature string,
	shp map[string]string,
) (*pb.ConfirmResultResponse, error) {
	return c.client.ConfirmResult(ctx, &pb.ConfirmResultRequest{
		OutSum:         outSum,
		InvId:          invID,
		SignatureValue: signature,
		Shp:            shp,
	})
}

func (c *PaymentClient) GetPaymentStats(ctx context.Context, periods *pb.PeriodBoundaries) (*pb.GetPaymentStatsResponse, error) {
	return c.client.GetPaymentStats(ctx, &pb.GetPaymentStatsRequest{Periods: periods})
}

func (c *PaymentClient) ListUserPayments(ctx context.Context, userID string) (*pb.ListUserPaymentsResponse, error) {
	return c.client.ListUserPayments(ctx, &pb.ListUserPaymentsRequest{UserId: userID})
}
