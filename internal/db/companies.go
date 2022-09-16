package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pzabolotniy/logging/pkg/logging"
)

type Company struct {
	Name      string    `db:"name"`
	Code      string    `db:"code"`
	Country   string    `db:"country"`
	WebSite   string    `db:"website"`
	CreatedAt time.Time `db:"created_at"`
	Phone     string    `db:"phone"`
	ID        uuid.UUID `db:"id"`
}

type NamedExerContext interface {
	NamedExecContext(ctx context.Context, query string, arg any) (sql.Result, error)
}

func CreateCompany(ctx context.Context, dbConn NamedExerContext, dbCompany *Company) error {
	logger := logging.FromContext(ctx)
	query := `INSERT INTO companies (
    id, name, code, country, website, phone, created_at
) VALUES (
	:id, :name, :code, :country, :website, :phone, :created_at
)`
	_, err := dbConn.NamedExecContext(ctx, query, dbCompany)
	if err != nil {
		logger.WithError(err).Error("insert company failed")

		return err
	}

	return nil
}

func DeleteCompanyByID(ctx context.Context, dbConn sqlx.ExecerContext, companyID uuid.UUID) error {
	logger := logging.FromContext(ctx)
	query := `DELETE FROM companies WHERE id = $1`
	_, err := dbConn.ExecContext(ctx, query, companyID)
	if err != nil {
		logger.WithError(err).WithField("company_id", companyID).Error("delete company failed")

		return err
	}

	return nil
}
