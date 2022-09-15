package webapi

import "github.com/jmoiron/sqlx"

type HandlerEnv struct {
	DbConn *sqlx.DB
}
