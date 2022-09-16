package webapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/pzabolotniy/logging/pkg/logging"

	"github.com/pzabolotniy/xm-golang-exercise/internal/db"
)

func (h *HandlerEnv) DeleteCompany(w http.ResponseWriter, r *http.Request) {
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

	err = db.DeleteCompanyByID(ctx, dbConn, companyID)
	if err != nil {
		logger.WithError(err).WithField("company_id", urlCompanyID).Warn("delete company failed")
		InternalServerError(ctx, w, "delete company failed")

		return
	}

	OKResponse(ctx, w, nil)
}
