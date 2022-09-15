package main

import (
	"context"
	"net/http"

	"github.com/pzabolotniy/logging/pkg/logging"
	"github.com/pzabolotniy/xm-golang-exercise/internal/config"
	"github.com/pzabolotniy/xm-golang-exercise/internal/db"
	"github.com/pzabolotniy/xm-golang-exercise/internal/migration"
	"github.com/pzabolotniy/xm-golang-exercise/internal/webapi"
)

func main() {
	logger := logging.GetLogger()
	appConf, err := config.LoadConfig()
	if err != nil {
		logger.WithError(err).Error("load config failed")

		return
	}
	ctx := context.Background()
	ctx = logging.WithContext(ctx, logger)

	dbConn, err := db.Connect(ctx, appConf.DB)
	if err != nil {
		logger.WithError(err).Error("db connect failed")

		return
	}
	defer func() {
		if closeErr := db.Disconnect(dbConn); closeErr != nil {
			logger.WithError(closeErr).Error("db disconnect failed")
		}
	}()

	err = migration.MigrateUp(ctx, dbConn, appConf.DB)
	if err != nil {
		logger.WithError(err).Error("migration failed")

		return
	}

	handler := &webapi.HandlerEnv{
		DbConn: dbConn,
	}
	router := webapi.CreateRouter(logger, handler)
	logger.WithField("listen", appConf.WebAPI.Listen).Trace("listen addr")
	if listenErr := http.ListenAndServe(appConf.WebAPI.Listen, router); listenErr != nil {
		logger.WithError(listenErr).WithField("listen", appConf.WebAPI.Listen).Error("listen failed")

		return
	}
}
