package models

import "github.com/spiritov/svelte-go-steam/db/queries"

func GetSessionResponse(u queries.User, ur queries.UserRole) Session {
	return Session{
		ID:        u.ID,
		Role:      ur.Role,
		Alias:     u.Alias,
		AvatarURL: u.AvatarUrl,
	}
}

func GetUserResponse(u queries.User) User {
	return User{
		ID:        u.ID,
		Alias:     u.Alias,
		AvatarURL: u.AvatarUrl,
		CreatedAt: u.CreatedAt,
	}
}
