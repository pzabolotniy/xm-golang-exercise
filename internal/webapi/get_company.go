package webapi

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/pzabolotniy/logging/pkg/logging"

	"github.com/pzabolotniy/xm-golang-exercise/internal/db"
)

type GetCompanyResponse struct {
	CompanyResponse
}

func (h *HandlerEnv) GetCompany(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logging.FromContext(ctx)
	dbConn := h.DbConn

	urlCompanyID := chi.URLParam(r, "companyID")
	companyID, err := uuid.Parse(urlCompanyID)
	if err != nil {
		logger.WithError(err).WithField("company_id", urlCompanyID).Warn("parse companyID failed")
		BadRequest(ctx, w, "invalid companyID")

		return
	}

	dbCompany, err := db.GetCompanyByID(ctx, dbConn, companyID)
	if err != nil {
		logger.WithError(err).WithField("company_id", companyID).Warn("get company failed")
		if errors.Is(err, sql.ErrNoRows) {
			NotFound(ctx, w, "company not found")

			return
		}
		InternalServerError(ctx, w, "get company failed")

		return
	}

	response := &GetCompanyResponse{
		CompanyResponse: CompanyResponse{
			CreatedAt: dbCompany.CreatedAt,
			InputCompany: InputCompany{
				Name:    dbCompany.Name,
				Code:    dbCompany.Code,
				Country: dbCompany.Country,
				WebSite: dbCompany.WebSite,
				Phone:   dbCompany.Phone,
			},
			ID: dbCompany.ID,
		},
	}
	OKResponse(ctx, w, response)
}
