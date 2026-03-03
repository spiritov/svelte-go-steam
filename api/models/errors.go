package models

import "github.com/danielgtaylor/huma/v2"

func WrapDBErr(err error) error {
	return huma.Error500InternalServerError("internal database error", err)
}

func SessionErr() error {
	return huma.Error401Unauthorized("session required")
}
