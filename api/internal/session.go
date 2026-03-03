package internal

import (
	"context"

	"github.com/spiritov/svelte-go-steam/db"
	"github.com/spiritov/svelte-go-steam/internal/principal"
	"github.com/spiritov/svelte-go-steam/models"
)

func HandleGetSession(ctx context.Context, _ *struct{}) (*models.SessionOutput, error) {
	principal, ok := principal.Get(ctx)
	if !ok {
		return nil, models.SessionErr()
	}

	user, err := db.Queries.SelectUser(ctx, principal.SteamID.String())
	if err != nil {
		return nil, models.WrapDBErr(err)
	}
	role, err := db.Queries.SelectRole(ctx, principal.SteamID.String())

	resp := &models.SessionOutput{Body: models.GetSessionResponse(user, role)}
	return resp, nil
}
