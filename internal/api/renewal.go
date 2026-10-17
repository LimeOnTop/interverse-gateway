package api

import (
	pb "github.com/LimeOnTop/interverse-contracts/payment/gen"
	"github.com/LimeOnTop/interverse-gateway/internal/apperr"
	"github.com/LimeOnTop/interverse-gateway/internal/middleware"
	"github.com/gin-gonic/gin"
	"net/http"
)

func writeRenewal(c *gin.Context, resp *pb.RenewalResponse, err error) {
	if err != nil {
		apperr.Internal(c, err)
		return
	}
	if resp.GetResponse() == nil || !resp.GetResponse().GetSuccess() {
		apperr.Upstream(c, http.StatusBadRequest, resp.GetResponse().GetError())
		return
	}
	c.JSON(http.StatusOK, gin.H{"enabled": resp.GetEnabled(), "has_mandate": resp.GetHasMandate(), "available": resp.GetAvailable(), "next_charge_at": resp.GetNextChargeAt(), "amount": resp.GetAmount(), "pending": resp.GetPending(), "last_error": resp.GetLastError()})
}
func (a *PaymentAPI) GetRenewal(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return
	}
	if middleware.IsAdmin(user) {
		apperr.Public(c, http.StatusBadRequest, "Администратору подписка не требуется")
		return
	}
	resp, err := a.paymentClient.GetRenewal(c.Request.Context(), user.ID)
	writeRenewal(c, resp, err)
}
func (a *PaymentAPI) SetRenewal(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return
	}
	if middleware.IsAdmin(user) {
		apperr.Public(c, http.StatusBadRequest, "Администратратору подписка не требуется")
		return
	}
	var req struct {
		Enabled *bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Enabled == nil {
		apperr.Public(c, http.StatusBadRequest, "Укажите состояние автопродления")
		return
	}
	resp, err := a.paymentClient.SetRenewal(c.Request.Context(), user.ID, *req.Enabled)
	writeRenewal(c, resp, err)
}
