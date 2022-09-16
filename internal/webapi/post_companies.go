package webapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/pzabolotniy/logging/pkg/logging"

	"github.com/pzabolotniy/xm-golang-exercise/internal/db"
)

type InputCompany struct {
	Name    string `json:"name"`
	Code    string `json:"code"`
	Country string `json:"country"`
	WebSite string `json:"website"` //nolint:tagliatelle // need to discuss
	Phone   string `json:"phone"`
}

type PostCompanyResponse struct {
	CreatedAt time.Time `json:"created_at"`
	InputCompany
	ID uuid.UUID `json:"id"`
}

func (h *HandlerEnv) PostCompanies(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logging.FromContext(ctx)
	dbConn := h.DbConn

	input := new(InputCompany)
	err := json.NewDecoder(r.Body).Decode(input)
	if err != nil {
		logger.WithError(err).Error("decode input failed")
		BadRequest(ctx, w, "Error while decoding user from request")

		return
	}

	companyID := NewCompanyID()
	createdAt := NewCreatedAt()
	dbCompany := &db.Company{
		ID:        companyID,
		Name:      input.Name,
		Code:      input.Code,
		Country:   input.Country,
		WebSite:   input.WebSite,
		Phone:     input.Phone,
		CreatedAt: createdAt,
	}
	err = db.CreateCompany(ctx, dbConn, dbCompany)
	if err != nil {
		logger.WithError(err).Error("create company failed")
		InternalServerError(ctx, w, "create company failed")

		return
	}

	response := &PostCompanyResponse{
		InputCompany: *input,
		ID:           companyID,
		CreatedAt:    createdAt,
	}
	CreatedResponse(ctx, w, response)
}

func NewCreatedAt() time.Time {
	return time.Now().UTC()
}

func NewCompanyID() uuid.UUID {
	return uuid.New()
}
