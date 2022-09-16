package geoip

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/pzabolotniy/logging/pkg/logging"

	"github.com/pzabolotniy/xm-golang-exercise/internal/config"
)

var ErrUnexpectedGeoIPStatus = errors.New("unexpected geoip status code")

type CountryDetector interface {
	CountryByIP(ctx context.Context, clientIP string) (string, error)
}

type GeoIPService struct {
	Client   *http.Client
	EndPoint string
}

func NewGeoIPService(geoIPConf *config.GeoIP) *GeoIPService {
	geoIPClient := &http.Client{
		Timeout: geoIPConf.Timeout,
	}
	geoIPService := &GeoIPService{
		EndPoint: geoIPConf.EndPoint,
		Client:   geoIPClient,
	}

	return geoIPService
}

type GeoIPResponse struct {
	CountryName string `json:"country_name"`
}

func (gis *GeoIPService) CountryByIP(ctx context.Context, clientIP string) (string, error) {
	logger := logging.FromContext(ctx)
	client := gis.Client
	url := fmt.Sprintf("%s/%s/%s", gis.EndPoint, clientIP, "json")
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		logger.WithError(err).Error("create geoip request failed")

		return "", fmt.Errorf("create geoip request failed: %w", err)
	}
	resp, err := client.Do(request)
	if err != nil {
		logger.WithError(err).Error("make geoip request failed")

		return "", fmt.Errorf("make geoip request failed: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.WithError(closeErr).Error("close geoip response body failed")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		logger.WithField("status_code", resp.StatusCode).Error("unexpected geoip status code")

		return "", ErrUnexpectedGeoIPStatus
	}

	geoIP := new(GeoIPResponse)
	err = json.NewDecoder(resp.Body).Decode(geoIP)
	if err != nil {
		logger.WithError(err).Error("decode geoip response failed")

		return "", fmt.Errorf("decode geoip response failed: %w", err)
	}

	return geoIP.CountryName, nil
}
