package main

import (
	"fmt"

	"github.com/pzabolotniy/logging/pkg/logging"

	"github.com/pzabolotniy/xm-golang-exercise/internal/authn"
	"github.com/pzabolotniy/xm-golang-exercise/internal/config"
)

func main() {
	logger := logging.GetLogger()
	appConf, err := config.LoadConfig()
	if err != nil {
		logger.WithError(err).Error("load config failed")

		return
	}
	tokenService := authn.NewTokenService(appConf.ClientToken)
	token, err := tokenService.IssueToken()
	if err != nil {
		logger.WithError(err).Error("issue token failed")

		return
	}
	fmt.Println(token) //nolint:forbidigo // this is output
}
