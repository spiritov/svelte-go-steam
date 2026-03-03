package routes

import (
	"context"

	"github.com/spiritov/svelte-go-steam/db"
	"github.com/spiritov/svelte-go-steam/internal/principal"
	"github.com/spiritov/svelte-go-steam/models"
)

// open

func HandleGetUser(ctx context.Context, input *models.UserIDInput) (*models.UserOutput, error) {
	user, err := db.Queries.SelectUser(ctx, input.UserID)
	if err != nil {
		return nil, models.WrapDBErr(err)
	}
	resp := &models.UserOutput{
		Body: models.GetUserResponse(user),
	}
	return resp, nil
}

// session

func HandleGetUserSelf(ctx context.Context, input *struct{}) (*models.UserOutput, error) {
	// use session principal steam ID
	principal, ok := principal.Get(ctx)
	if !ok {
		return nil, models.SessionErr()
	}
	user, err := db.Queries.SelectUser(ctx, principal.SteamID.String())
	if err != nil {
		return nil, models.WrapDBErr(err)
	}
	resp := &models.UserOutput{
		Body: models.GetUserResponse(user),
	}
	return resp, nil
}
