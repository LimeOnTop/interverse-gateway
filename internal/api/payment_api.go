package api

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/LimeOnTop/interverse-gateway/internal/apperr"
	"github.com/LimeOnTop/interverse-gateway/internal/clients"
	"github.com/LimeOnTop/interverse-gateway/internal/middleware"
	"github.com/gin-gonic/gin"
)

type PaymentAPI struct {
	paymentClient *clients.PaymentClient
}

func NewPaymentAPI(paymentClient *clients.PaymentClient) *PaymentAPI {
	return &PaymentAPI{paymentClient: paymentClient}
}

func (a *PaymentAPI) CreatePayment(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return
	}
	if user.ID == "-1" {
		apperr.Public(c, http.StatusBadRequest, "admin account does not require subscription payment")
		return
	}

	var req struct {
		Plan string `json:"plan"`
	}
	_ = c.ShouldBindJSON(&req)
	plan := strings.TrimSpace(req.Plan)
	if plan == "" {
		plan = "paid_1m"
	}

	resp, err := a.paymentClient.CreatePayment(c.Request.Context(), user.ID, plan, user.Email)
	if err != nil {
		apperr.Internal(c, err)
		return
	}
	if resp.GetResponse() != nil && !resp.GetResponse().GetSuccess() {
		msg := resp.GetResponse().GetError()
		if strings.Contains(strings.ToLower(msg), "robokassa") || strings.Contains(strings.ToLower(msg), "не настроена") {
			apperr.Public(c, http.StatusServiceUnavailable, msg)
			return
		}
		apperr.Upstream(c, http.StatusBadRequest, msg)
		return
	}

	payment := resp.GetPayment()
	c.JSON(http.StatusOK, gin.H{
		"inv_id":      resp.GetInvId(),
		"amount":      resp.GetAmount(),
		"description": resp.GetDescription(),
		"plan":        resp.GetPlan(),
		"status":      resp.GetStatus(),
		"payment": gin.H{
			"action": payment.GetAction(),
			"method": payment.GetMethod(),
			"fields": payment.GetFields(),
		},
	})
}

func (a *PaymentAPI) GetOffers(c *gin.Context) {
	resp, err := a.paymentClient.GetOffers(c.Request.Context())
	if err != nil {
		apperr.Internal(c, err)
		return
	}
	if resp.GetResponse() != nil && !resp.GetResponse().GetSuccess() {
		apperr.Upstream(c, http.StatusBadRequest, resp.GetResponse().GetError())
		return
	}

	offers := make([]gin.H, 0, len(resp.GetOffers()))
	for _, offer := range resp.GetOffers() {
		offers = append(offers, gin.H{
			"id":            offer.GetId(),
			"name":          offer.GetName(),
			"price":         offer.GetPrice(),
			"regular_price": offer.GetRegularPrice(),
			"duration_days": offer.GetDurationDays(),
			"price_hint":    offer.GetPriceHint(),
			"description":   offer.GetDescription(),
			"badge":         offer.GetBadge(),
			"early_bird":    offer.GetEarlyBird(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"offers":               offers,
		"early_bird_remaining": resp.GetEarlyBirdRemaining(),
		"early_bird_limit":     resp.GetEarlyBirdLimit(),
	})
}

func (a *PaymentAPI) RobokassaResult(c *gin.Context) {
	values := map[string]string{}
	_ = c.Request.ParseForm()
	for k, v := range c.Request.Form {
		if len(v) > 0 {
			values[k] = v[0]
		}
	}
	for k, v := range c.Request.URL.Query() {
		if _, exists := values[k]; !exists && len(v) > 0 {
			values[k] = v[0]
		}
	}

	outSum := firstNonEmpty(values["OutSum"], values["outsum"])
	invID := firstNonEmpty(values["InvId"], values["InvID"], values["invid"])
	signature := firstNonEmpty(values["SignatureValue"], values["signaturevalue"])
	shp := extractShp(values)

	if outSum == "" || invID == "" || signature == "" {
		c.String(http.StatusBadRequest, "bad request")
		return
	}

	inv, err := strconv.ParseInt(invID, 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "bad inv id")
		return
	}

	resp, err := a.paymentClient.ConfirmResult(c.Request.Context(), outSum, inv, signature, shp)
	if err != nil {
		log.Printf("robokassa result: inv %d: confirm failed: %v", inv, err)
		c.String(http.StatusInternalServerError, "confirm failed")
		return
	}
	if resp.GetResponse() != nil && !resp.GetResponse().GetSuccess() {
		log.Printf("robokassa result: inv %d: rejected: %s", inv, resp.GetResponse().GetError())
		c.String(http.StatusBadRequest, "confirm rejected")
		return
	}

	okBody := resp.GetOkBody()
	if okBody == "" {
		okBody = "OK" + invID
	}
	c.String(http.StatusOK, okBody)
}

func extractShp(values map[string]string) map[string]string {
	out := make(map[string]string)
	for k, v := range values {
		if strings.HasPrefix(k, "Shp_") || strings.HasPrefix(k, "shp_") {
			key := k
			if strings.HasPrefix(k, "shp_") {
				key = "Shp_" + k[4:]
			}
			out[key] = v
		}
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
