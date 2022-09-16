package webapi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pzabolotniy/logging/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/pzabolotniy/xm-golang-exercise/internal/geoip/mocks"
)

type OkHandler struct {
	ctx context.Context
}

func (h *OkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	CreatedResponse(h.ctx, w, map[string]string{"foo": "bar"})
}

func TestWithCountryRestriction_OK(t *testing.T) {
	ctx := context.Background()
	logger := logging.GetLogger()
	ctx = logging.WithContext(ctx, logger)

	myIP := "185.70.53.133"
	allowedCountryName := "Georgia"
	geoIPMock := &mocks.CountryDetector{}
	geoIPMock.On("CountryByIP", mock.AnythingOfType("*context.valueCtx"), myIP).Return(allowedCountryName, nil)
	defer geoIPMock.AssertExpectations(t)

	handlerFn := WithCountryRestriction(geoIPMock, allowedCountryName)

	testRecorder := httptest.NewRecorder()
	testRequest := httptest.NewRequest(http.MethodPost, "/any", nil)
	testRequest.RemoteAddr = fmt.Sprintf("%s:12076", myIP)
	testRequest = testRequest.WithContext(ctx)
	handlerFn(&OkHandler{ctx: ctx}).ServeHTTP(testRecorder, testRequest)

	gotHttpCode := testRecorder.Code
	gotResponseBody, err := io.ReadAll(testRecorder.Body)
	if err != nil {
		t.Fatal(err)
	}
	expectedHTTPCode := http.StatusCreated
	assert.Equal(t, expectedHTTPCode, gotHttpCode, "http code must match")

	expectedBody := `{"data": {"foo": "bar"}}`
	assert.JSONEq(t, expectedBody, string(gotResponseBody), "response body must match")
}

func TestWithCountryRestriction_SplitRemoteAddrFailed(t *testing.T) {
	ctx := context.Background()
	logger := logging.GetLogger()
	ctx = logging.WithContext(ctx, logger)

	allowedCountryName := "Georgia"
	geoIPMock := &mocks.CountryDetector{}
	handlerFn := WithCountryRestriction(geoIPMock, allowedCountryName)

	testRecorder := httptest.NewRecorder()
	testRequest := httptest.NewRequest(http.MethodPost, "/any", nil)
	testRequest.RemoteAddr = "" // <-- invalid remote Addr
	testRequest = testRequest.WithContext(ctx)
	handlerFn(nil).ServeHTTP(testRecorder, testRequest)

	gotHttpCode := testRecorder.Code
	gotResponseBody, err := io.ReadAll(testRecorder.Body)
	if err != nil {
		t.Fatal(err)
	}
	expectedHTTPCode := http.StatusInternalServerError
	assert.Equal(t, expectedHTTPCode, gotHttpCode, "http code must match")

	expectedBody := `{"error": "verify client country failed"}`
	assert.JSONEq(t, expectedBody, string(gotResponseBody), "response body must match")
}

func TestWithCountryRestriction_CountryByIPFailed(t *testing.T) {
	ctx := context.Background()
	logger := logging.GetLogger()
	ctx = logging.WithContext(ctx, logger)

	myIP := "185.70.53.133"
	allowedCountryName := "Georgia"
	geoIPMock := &mocks.CountryDetector{}
	geoIPMock.On("CountryByIP", mock.AnythingOfType("*context.valueCtx"), myIP).Return("", errors.New("geoip call failed"))
	defer geoIPMock.AssertExpectations(t)

	handlerFn := WithCountryRestriction(geoIPMock, allowedCountryName)

	testRecorder := httptest.NewRecorder()
	testRequest := httptest.NewRequest(http.MethodPost, "/any", nil)
	testRequest.RemoteAddr = fmt.Sprintf("%s:12076", myIP)
	testRequest = testRequest.WithContext(ctx)
	handlerFn(nil).ServeHTTP(testRecorder, testRequest)

	gotHttpCode := testRecorder.Code
	gotResponseBody, err := io.ReadAll(testRecorder.Body)
	if err != nil {
		t.Fatal(err)
	}
	expectedHTTPCode := http.StatusInternalServerError
	assert.Equal(t, expectedHTTPCode, gotHttpCode, "http code must match")

	expectedBody := `{"error": "verify client country failed"}`
	assert.JSONEq(t, expectedBody, string(gotResponseBody), "response body must match")
}

func TestWithCountryRestriction_CountryMisMatch(t *testing.T) {
	ctx := context.Background()
	logger := logging.GetLogger()
	ctx = logging.WithContext(ctx, logger)

	myIP := "185.70.53.133"
	allowedCountryName := "Georgia"
	myCountry := "Hawkins"
	geoIPMock := &mocks.CountryDetector{}
	geoIPMock.On("CountryByIP", mock.AnythingOfType("*context.valueCtx"), myIP).Return(myCountry, nil)
	defer geoIPMock.AssertExpectations(t)

	handlerFn := WithCountryRestriction(geoIPMock, allowedCountryName)

	testRecorder := httptest.NewRecorder()
	testRequest := httptest.NewRequest(http.MethodPost, "/any", nil)
	testRequest.RemoteAddr = fmt.Sprintf("%s:12076", myIP)
	testRequest = testRequest.WithContext(ctx)
	handlerFn(nil).ServeHTTP(testRecorder, testRequest)

	gotHttpCode := testRecorder.Code
	gotResponseBody, err := io.ReadAll(testRecorder.Body)
	if err != nil {
		t.Fatal(err)
	}
	expectedHTTPCode := http.StatusForbidden
	assert.Equal(t, expectedHTTPCode, gotHttpCode, "http code must match")

	expectedBody := `{"error": "access denied"}`
	assert.JSONEq(t, expectedBody, string(gotResponseBody), "response body must match")
}
