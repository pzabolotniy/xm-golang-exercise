package webapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/jmoiron/sqlx"
	"github.com/ory/dockertest/v3"
	"github.com/pzabolotniy/logging/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/pzabolotniy/xm-golang-exercise/internal/config"
	"github.com/pzabolotniy/xm-golang-exercise/internal/db"
	"github.com/pzabolotniy/xm-golang-exercise/internal/migration"
)

type PostCompaniesSearchSuite struct {
	suite.Suite
	pgResource *dockertest.Resource
	dockerPool *dockertest.Pool
	dbConn     *sqlx.DB
	logger     logging.Logger
	appConf    *config.App
	router     *chi.Mux
}

func TestPostCompaniesSearchSuite(t *testing.T) {
	s := new(PostCompaniesSearchSuite)
	suite.Run(t, s)
}

func (s *PostCompaniesSearchSuite) SetupSuite() {
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

	appConf := &config.App{
		DB:     dbConf,
		WebAPI: &config.WebAPI{Listen: ""},
		GeoIP: &config.GeoIP{
			AllowedCountryName: "localhost",
			EndPoint:           "http://ip.me",
			Timeout:            10 * time.Second,
		},
		ClientToken: &config.ClientToken{
			TTL:    1 * time.Hour,
			Issuer: "test",
			Secret: "secret",
		},
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

	s.pgResource = pgResource
	s.dockerPool = pool
	s.dbConn = dbConn
	s.logger = logger
	s.appConf = appConf
}

func (s *PostCompaniesSearchSuite) TearDownSuite() {
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

func (s *PostCompaniesSearchSuite) SetupTest() {
	handler := &HandlerEnv{DbConn: s.dbConn}
	routerParams := &RouterParams{
		Logger:    s.logger,
		Handler:   handler,
		GeoIPConf: s.appConf.GeoIP,
	}
	router := CreateRouter(routerParams)
	s.router = router
}

func (s *PostCompaniesSearchSuite) TearDownTest() {}

func (s *PostCompaniesSearchSuite) TestPostCompaniesSearch_OK() {
	t := s.T()

	// testdata
	companyID1 := "43fa9b5e-87bf-45d1-ad3a-b15df0037f37"
	companyID2 := "5b6e7620-808f-4c9a-887c-56fe5290f535"
	body := strings.NewReader(fmt.Sprintf(`{"companies_ids": ["%s", "%s"]}`, companyID1, companyID2))

	// make request
	testURL := "/api/v1/search/companies"
	response := makeTestRequest(s.router, http.MethodPost, testURL, body, nil)

	// assert HTTP code
	gotHTTPCode := response.Code
	gotBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response body failed: %s", err)
	}
	expectedHTTPCode := http.StatusOK
	assert.Equal(t, expectedHTTPCode, gotHTTPCode, "http code must match")

	// assert HTTP body
	expectedHTTPBody := `{
	"data": [
		{
			"id": "43fa9b5e-87bf-45d1-ad3a-b15df0037f37",
			"name": "TestGetCompany_OK",
			"code": "OK",
			"country": "Moon",
			"website": "Moon.dark",
			"phone": "+65748329",
			"created_at": "2022-09-16T16:05:15Z"
		},
		{
			"id": "5b6e7620-808f-4c9a-887c-56fe5290f535",
			"name": "TestDeleteCompanySuite_OK",
			"code": "OK",
			"country": "Sun",
			"website": "sun.info",
			"phone": "+987765543",
			"created_at": "2022-09-16T07:36:15Z"
		}
	]
}`
	assert.JSONEq(t, expectedHTTPBody, string(gotBody), "body must match")
}
