package api

import (
	"context"

	"github.com/LimeOnTop/interverse-gateway/internal/usecase"
)

const (
	planFree = "free"
	planPaid = "paid"
)

// subscriptionPlan resolves the user's plan for quotas and gated content.
// Only an active subscription counts as Pro; any lookup failure means Basic.
func subscriptionPlan(ctx context.Context, authClient usecase.AuthGateway, userID string) string {
	if authClient == nil {
		return planFree
	}
	resp, err := authClient.GetUser(ctx, userID)
	if err != nil || resp.Response == nil || !resp.Response.Success || resp.User == nil {
		return planFree
	}
	if resp.User.SubscriptionActive {
		return planPaid
	}
	return planFree
}
