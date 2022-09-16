package webapi

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/pzabolotniy/logging/pkg/logging"

	"github.com/pzabolotniy/xm-golang-exercise/internal/db"
)

type CompaniesSearchRequest struct {
	CompaniesIDs []string `json:"companies_ids"` //nolint:tagliatelle // false positive
}

type CompaniesSearchResponse []CompanyResponse

// PostCompaniesSearch is used to fetch list of companies
// POST was chosen instead GET
// we can pass parameters only in URL, but URL max length is 2048,
// so we will pass parameters in body in POST request.
func (h *HandlerEnv) PostCompaniesSearch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logging.FromContext(ctx)
	dbConn := h.DbConn

	input := new(CompaniesSearchRequest)
	err := json.NewDecoder(r.Body).Decode(input)
	if err != nil {
		logger.WithError(err).Error("decode input failed")
		BadRequest(ctx, w, "decode request failed")

		return
	}

	companyIDs := make([]uuid.UUID, 0)
	for _, inputID := range input.CompaniesIDs {
		companyID, parseErr := uuid.Parse(inputID)
		if parseErr != nil {
			logger.WithError(parseErr).WithField("company_id", inputID).Warn("parse companyID failed")

			continue
		}
		companyIDs = append(companyIDs, companyID)
	}

	dbCompanies, err := db.GetCompaniesListByID(ctx, dbConn, companyIDs)
	if err != nil {
		logger.WithError(err).Error("select companies failed")
		BadRequest(ctx, w, "select companies failed")

		return
	}

	response := make(CompaniesSearchResponse, 0)
	for i := range dbCompanies {
		dbCompany := dbCompanies[i]
		companyResponse := CompanyResponse{
			CreatedAt: dbCompany.CreatedAt,
			InputCompany: InputCompany{
				Name:    dbCompany.Name,
				Code:    dbCompany.Code,
				Country: dbCompany.Country,
				WebSite: dbCompany.WebSite,
				Phone:   dbCompany.Phone,
			},
			ID: dbCompany.ID,
		}
		response = append(response, companyResponse)
	}
	OKResponse(ctx, w, response)
}
