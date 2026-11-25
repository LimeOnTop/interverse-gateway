package api

import (
	"context"
	pb "github.com/LimeOnTop/interverse-contracts/payment/gen"
	"github.com/LimeOnTop/interverse-gateway/internal/apperr"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func (a *AdminAPI) ListOperations(c *gin.Context) {
	before, err := strconv.ParseInt(c.DefaultQuery("before_id", "0"), 10, 64)
	if err != nil || before < 0 {
		apperr.Public(c, http.StatusBadRequest, "Некорректная страница")
		return
	}
	reader, ok := a.paymentClient.(interface {
		ListOperations(context.Context, int64) (*pb.ListOperationsResponse, error)
	})
	if !ok {
		apperr.Upstream(c, http.StatusBadGateway, "Журнал операций недоступен")
		return
	}
	result, err := reader.ListOperations(c.Request.Context(), before)
	if err != nil {
		apperr.Internal(c, err)
		return
	}
	operations := make([]gin.H, 0, len(result.GetOperations()))
	for _, o := range result.GetOperations() {
		operations = append(operations, gin.H{"id": o.GetId(), "user_id": o.GetUserId(), "plan": o.GetPlan(), "amount": o.GetAmount(), "currency": o.GetCurrency(), "status": o.GetStatus(), "created_at": o.GetCreatedAt(), "paid_at": o.GetPaidAt()})
	}
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{"operations": operations, "next_before_id": result.GetNextBeforeId()})
}
