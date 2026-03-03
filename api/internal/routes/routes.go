package routes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func RegisterOpenRoutes(internalApi *huma.Group) {
	huma.Register(internalApi, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/users/{user_id}",
		Tags:        []string{"users"},
		Summary:     "get user",
		Description: "get a user by user ID",
		OperationID: "get-user",
	}, HandleGetUser)
}

func RegisterSessionRoutes(internalApi *huma.Group) {
	huma.Register(internalApi, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/users/self",
		Tags:        []string{"users"},
		Summary:     "get user self",
		Description: "get user self from session",
		OperationID: "get-user-self",
	}, HandleGetUserSelf)
}

func RegisterDevRoutes(devApi *huma.Group) {

}
