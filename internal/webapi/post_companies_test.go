package webapi

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/pzabolotniy/logging/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/pzabolotniy/xm-golang-exercise/internal/authn"
	"github.com/pzabolotniy/xm-golang-exercise/internal/config"
	"github.com/pzabolotniy/xm-golang-exercise/internal/db"
	geoipMocks "github.com/pzabolotniy/xm-golang-exercise/internal/geoip/mocks"
	"github.com/pzabolotniy/xm-golang-exercise/internal/migration"
)

type PostCompaniesSuite struct {
	suite.Suite
	pgResource          *dockertest.Resource
	dockerPool          *dockertest.Pool
	dbConn              *sqlx.DB
	logger              logging.Logger
	appConf             *config.App
	router              *chi.Mux
	countryDetectorMock *geoipMocks.CountryDetector
	testJWT             string
}

func TestPostCompaniesSuite(t *testing.T) {
	s := new(PostCompaniesSuite)
	suite.Run(t, s)
}

func (s *PostCompaniesSuite) SetupSuite() {
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

	s.pgResource = pgResource
	s.dockerPool = pool
	s.dbConn = dbConn
	s.logger = logger
	s.appConf = appConf
}

func (s *PostCompaniesSuite) TearDownSuite() {
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

func (s *PostCompaniesSuite) SetupTest() {
	countryDetectorMock := &geoipMocks.CountryDetector{}
	countryDetectorMock.
		On("CountryByIP", mock.AnythingOfType("*context.valueCtx"), "127.0.0.1").
		Return("localhost", nil)

	handler := &HandlerEnv{DbConn: s.dbConn}
	tokenService := authn.NewTokenService(s.appConf.ClientToken)
	testJWT, err := tokenService.IssueToken()
	if err != nil {
		s.T().Fatal(err)
	}
	routerParams := &RouterParams{
		Logger:          s.logger,
		Handler:         handler,
		CountryDetector: countryDetectorMock,
		GeoIPConf:       s.appConf.GeoIP,
		TokenService:    tokenService,
	}
	router := CreateRouter(routerParams)

	s.router = router
	s.testJWT = testJWT
	s.countryDetectorMock = countryDetectorMock
}

func (s *PostCompaniesSuite) TearDownTest() {
	s.countryDetectorMock.AssertExpectations(s.T())
}

func (s *PostCompaniesSuite) TestPostCompanies_OK() {
	t := s.T()

	// prepare fake variadic parameters
	fakeUUID := uuid.MustParse("3997db3d-f747-4f00-adf8-1d2c71d2a911")
	uuidPatch := gomonkey.ApplyFunc(NewCompanyID, func() uuid.UUID {
		return fakeUUID
	})
	defer uuidPatch.Reset()

	fakeTime := time.Date(2022, 9, 15, 15, 04, 17, 0, time.UTC)
	timePatch := gomonkey.ApplyFunc(NewCreatedAt, func() time.Time {
		return fakeTime
	})
	defer timePatch.Reset()

	// prepare input data
	inputData := strings.NewReader(`{
	"name": "ltd",
	"code": "007",
	"country": "md",
	"website": "http://google.com",
	"phone": "+995987655443"
}`)

	// make request
	metadata := &testRequestMetaData{
		remoteAddr: "127.0.0.1:63099",
		headers: map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", s.testJWT),
		},
	}
	response := makeTestRequest(s.router, http.MethodPost, "/api/v1/companies", inputData, metadata)

	// assert HTTP code
	gotHTTPCode := response.Code
	expectedHTTPCode := http.StatusCreated
	assert.Equal(t, expectedHTTPCode, gotHTTPCode, "http code must match")

	// assert HTTP body
	gotBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response body failed: %s", err)
	}
	expectedHTTPBody := fmt.Sprintf(`{
	"data": {
		"id": "%s",
		"name": "ltd",
		"code": "007",
		"country": "md",
		"website": "http://google.com",
		"phone": "+995987655443",
		"created_at": "%s"
	}
}`, fakeUUID, fakeTime.Format(time.RFC3339))
	assert.JSONEq(t, expectedHTTPBody, string(gotBody), "body must match")

	// assert db values
	dbCompany := selectDbCompanyByID(t, s.dbConn, fakeUUID.String())
	expectedDbCompany := &db.Company{
		Name:      "ltd",
		Code:      "007",
		Country:   "md",
		WebSite:   "http://google.com",
		CreatedAt: fakeTime,
		Phone:     "+995987655443",
		ID:        fakeUUID,
	}
	assert.Equal(t, expectedDbCompany, dbCompany, "db company must match")
}

func postgresqlResource(ctx context.Context, t *testing.T, pool *dockertest.Pool, dbUser, dbName, sslMode string) (*dockertest.Resource, *sqlx.DB, *config.DB) {
	t.Helper()
	// pulls an image, creates a container based on it and runs it
	pgResource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "14.5",
		Env: []string{
			fmt.Sprintf("POSTGRES_USER=%s", dbUser),
			fmt.Sprintf("POSTGRES_DB=%s", dbName),
			"POSTGRES_HOST_AUTH_METHOD=trust",
			"listen_addresses = '*'",
		},
		Cmd: []string{"postgres", "-c", "log_statement=all", "-c", "log_destination=stderr"},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		t.Fatalf("Could not start pgResource: %s", err)
	}

	hostAndPort := pgResource.GetHostPort("5432/tcp")
	dbHost, dbPort, err := net.SplitHostPort(hostAndPort)
	if err != nil {
		t.Fatalf("split host-port '%s' failed: '%s'", hostAndPort, err)
	}
	dbConf := &config.DB{
		ConnString:      fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s", dbHost, dbPort, dbUser, dbName, sslMode),
		MigrationDir:    "../../sql-migrations",
		MigrationTable:  "migrations",
		MaxOpenConns:    20,
		ConnMaxLifetime: 10 * time.Second,
	}

	var dbConn *sqlx.DB
	pool.MaxWait = 30 * time.Second
	if err = pool.Retry(func() error {
		dbConn, err = db.Connect(ctx, dbConf)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		t.Fatalf("Could not connect to docker: %s", err)
	}

	return pgResource, dbConn, dbConf
}

type testRequestMetaData struct {
	headers    map[string]string
	remoteAddr string
}

func makeTestRequest(r http.Handler, method, path string, body io.Reader, metadata *testRequestMetaData) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, body)
	httpHeaders := make(http.Header)
	for k, v := range metadata.headers {
		httpHeaders[k] = []string{v}
	}
	req.Header = httpHeaders
	req.RemoteAddr = metadata.remoteAddr
	req.RequestURI = path
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
