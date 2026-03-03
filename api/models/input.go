package models

type UserIDInput struct {
	UserID string `path:"user_id" doc:"user ID, SteamID64"`
}
