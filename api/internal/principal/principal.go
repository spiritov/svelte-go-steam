package principal

import (
	"context"

	"github.com/escrow-tf/steam/steamid"
	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v5"
)

var (
	PrincipalContextKey = "principal"
)

type Principal struct {
	SteamID steamid.SteamID
	TokenID uuid.UUID
	Claims  *jwt.RegisteredClaims
}

func Get(ctx context.Context) (result *Principal, ok bool) {
	result, ok = ctx.Value(PrincipalContextKey).(*Principal)
	ok = ok && result != nil
	return
}

func Has(ctx context.Context) bool {
	result, ok := ctx.Value(PrincipalContextKey).(*Principal)
	return ok && result != nil
}
