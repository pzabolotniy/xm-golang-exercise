package webapi

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/jmoiron/sqlx"
	"github.com/ory/dockertest/v3"
	"github.com/pzabolotniy/logging/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/pzabolotniy/xm-golang-exercise/internal/config"
	"github.com/pzabolotniy/xm-golang-exercise/internal/db"
	"github.com/pzabolotniy/xm-golang-exercise/internal/geoip/mocks"
	"github.com/pzabolotniy/xm-golang-exercise/internal/migration"
)

type DeleteCompanySuite struct {
	suite.Suite
	ctx        context.Context
	pgResource *dockertest.Resource
	dockerPool *dockertest.Pool
	dbConn     *sqlx.DB
	logger     logging.Logger
}

func TestDeleteCompanySuite(t *testing.T) {
	s := new(DeleteCompanySuite)
	suite.Run(t, s)
}

func (s *DeleteCompanySuite) SetupSuite() {
	t := s.T()
	ctx := context.Background()
	logger := logging.GetLogger()
	ctx = logging.WithContext(ctx, logger)

	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("Could not connect to docker: %s", err)
	}

	dbUser := "companies_db"
	dbName := "test_companies_db"
	sslMode := "disable"
	pgResource, dbConn, dbConf := postgresqlResource(ctx, t, pool, dbUser, dbName, sslMode)

	err = migration.MigrateUp(ctx, dbConn, dbConf)
	if err != nil {
		t.Fatalf("apply migrations failed: '%s'", err)
	}

	fixtures, err := testfixtures.New(
		testfixtures.Database(dbConn.DB),
		testfixtures.Dialect("postgres"),
		testfixtures.Files("fixtures/companies.yaml"),
	)
	if err != nil {
		t.Fatalf("prepare fixtures failed: %s", err)
	}
	err = fixtures.Load()
	if err != nil {
		t.Fatalf("load fixtures failed: %s", err)
	}

	s.ctx = ctx
	s.pgResource = pgResource
	s.dockerPool = pool
	s.dbConn = dbConn
	s.logger = logger
}

func (s *DeleteCompanySuite) TearDownSuite() {
	t := s.T()
	dbConn := s.dbConn
	if disconnectErr := db.Disconnect(dbConn); disconnectErr != nil {
		t.Fatalf("disconnect failed: '%s'", disconnectErr)
	}

	dockerPool := s.dockerPool
	pgResource := s.pgResource
	if purgeErr := dockerPool.Purge(pgResource); purgeErr != nil {
		t.Fatalf("purge pgResource '%s' failed", purgeErr)
	}
}

func (s *DeleteCompanySuite) TestDeleteCompanySuite_OK() {
	t := s.T()
	handler := &HandlerEnv{DbConn: s.dbConn}

	countryDetectorMock := &mocks.CountryDetector{}
	countryDetectorMock.
		On("CountryByIP", mock.AnythingOfType("*context.valueCtx"), "127.0.0.1").
		Return("localhost", nil)
	defer countryDetectorMock.AssertExpectations(t)
	router := CreateRouter(s.logger, handler, countryDetectorMock, &config.GeoIP{AllowedCountryName: "localhost"})

	companyID := "5b6e7620-808f-4c9a-887c-56fe5290f535"

	// make request
	metadata := &testRequestMetaData{
		remoteAddr: "127.0.0.1:63099",
	}
	testURL := fmt.Sprintf("/api/v1/companies/%s", companyID)
	response := makeTestRequest(router, http.MethodDelete, testURL, nil, metadata)

	// assert HTTP code
	gotHTTPCode := response.Code
	gotBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response body failed: %s", err)
	}
	expectedHTTPCode := http.StatusOK
	assert.Equal(t, expectedHTTPCode, gotHTTPCode, "http code must match")

	// assert HTTP body
	expectedHTTPBody := `{}`
	assert.JSONEq(t, expectedHTTPBody, string(gotBody), "body must match")

	dbCompany := selectDbCompanyByID(t, s.dbConn, companyID)
	assert.Nil(t, dbCompany, "companyID should be nil")
}

func selectDbCompanyByID(t *testing.T, dbConn *sqlx.DB, companyID string) *db.Company {
	dbCompany := new(db.Company)
	err := dbConn.QueryRowx(`SELECT id, name, code, country, website, phone, created_at FROM companies WHERE id = $1`, companyID).StructScan(dbCompany)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		t.Fatal(err)
	}
	return dbCompany
}
