package models

import "time"

type Session struct {
	ID        string `json:"id"`
	Role      string `json:"role"`
	Alias     string `json:"alias"`
	AvatarURL string `json:"avatar_url"`
}

type User struct {
	ID        string    `json:"id"`
	Alias     string    `json:"alias"`
	AvatarURL string    `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
}
