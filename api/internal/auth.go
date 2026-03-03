package internal

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/escrow-tf/steam/steamid"
	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rotisserie/eris"
	"github.com/spiritov/svelte-go-steam/db"
	"github.com/spiritov/svelte-go-steam/db/queries"
	"github.com/spiritov/svelte-go-steam/env"
	"github.com/spiritov/svelte-go-steam/internal/principal"
	"github.com/yohcop/openid-go"
)

const (
	SessionCookieName = "sessionid"

	SessionIssuer   = "jump"
	SessionAudience = "jump"
	SessionDuration = time.Hour * 24 * 7
	SessionJitter   = time.Minute

	SteamOidcIssuer      = "https://steamcommunity.com/openid/"
	SteamOidRedirectPath = "internal/steam/callback"
)

var (
	SessionTokenSecret  []byte
	SessionCookieSecure = false
	OidRealm            string
	OidRealmURL         *url.URL
	SteamApiKey         string

	discoveryCache *NoOpDiscoveryCache
)

type DiscoverInput struct {
	URL url.URL
}

func (d *DiscoverInput) Resolve(ctx huma.Context) []error {
	d.URL = ctx.URL()
	return nil
}

type DiscoverOutput struct {
	Status int
	Url    string `header:"Location"`
}

func handleSteamDiscover(ctx context.Context, input *DiscoverInput) (*DiscoverOutput, error) {
	callbackURL := OidRealmURL.JoinPath(SteamOidRedirectPath)
	redirectUrl, err := openid.RedirectURL(SteamOidcIssuer, callbackURL.String(), OidRealm)
	if err != nil {
		slog.Error("couldn't create openid redirect", "error", err)
		return nil, err
	}

	return &DiscoverOutput{
		Status: http.StatusTemporaryRedirect,
		Url:    redirectUrl,
	}, nil
}

type CallbackInput struct {
	URL url.URL
}

func (d *CallbackInput) Resolve(ctx huma.Context) []error {
	d.URL = ctx.URL()
	return nil
}

type CallbackOutput struct {
	Status    int
	Url       string      `header:"Location"`
	SetCookie http.Cookie `header:"Set-Cookie"`
}

func handleSteamCallback(ctx context.Context, input *CallbackInput) (*CallbackOutput, error) {
	// our openid library verifies that the original request came from our authority, but it needs us to
	// provide a URL to verify that the incoming callback request has the authority we expect. Here we're
	// just replacing the `https://blahblah.com` part of the URL with our OidRealm
	inputURL := input.URL
	fullURL := OidRealmURL.JoinPath(inputURL.Path)
	fullURL.RawQuery = inputURL.RawQuery

	// verify the openid callback. The discovery cache caches some response information to make verification
	// faster if the callback is hit with the same user again. The nonce store ensures that a callback request
	// is never processed by our servers more than once.
	id, err := openid.Verify(fullURL.String(), discoveryCache, db.NewNonceStore(ctx, db.Queries))
	if err != nil {
		slog.Debug("Error verifying openid callback", "error", err, "uri", fullURL)
		return nil, eris.Wrap(err, "Error verifying openid callback")
	}

	slog.Debug("Verified openid callback for steam user", "user_id", id)

	// the openid id is in the format `https?://steamcommunity.com/openid/id/[0-9]+`. We only care about the
	// last part, which is the user's Steam ID 64.
	var steamID64 string
	if strings.HasPrefix(id, "https") {
		_, err = fmt.Sscanf(id, "https://steamcommunity.com/openid/id/%s", &steamID64)
	} else {
		_, err = fmt.Sscanf(id, "http://steamcommunity.com/openid/id/%s", &steamID64)
	}
	if err != nil {
		slog.Error("Verified openid callback but couldn't parse Steam ID 64 from ID.", "id", id, "error", err)
		return nil, eris.Wrap(err, "Error parsing Steam ID 64 from ID")
	}

	steamID, err := steamid.ParseSteamID64(steamID64)
	if err != nil {
		slog.Error("Verified openid callback but couldn't parse SteamID64 to a full SteamID")
		return nil, eris.Wrap(err, "Error parsing SteamID64 to full SteamID")
	}

	steamProfile, err := FetchProfileSummary(steamID.ID())
	if err != nil {
		slog.Error("Verified openid callback but couldn't fetch Steam Profile summary")
		return nil, eris.Wrap(err, "Error fetching Steam profile summary")
	}

	// upsert user
	user, err := db.Queries.UpsertUser(ctx, queries.UpsertUserParams{
		ID:        steamID.String(),
		Alias:     steamProfile.PersonaName,
		AvatarUrl: steamProfile.AvatarURL,
	})
	if err != nil {
		return nil, eris.Wrap(err, "Error creating new user")
	}

	// init default role for user
	err = db.Queries.InitRole(ctx, user.ID)
	if err != nil {
		return nil, eris.Wrap(err, "Error initializing role")
	}

	// AddSession will create a new session UUIDv7 token entry and link it to the user.
	session, dbErr := db.Queries.InsertSession(ctx, queries.InsertSessionParams{
		TokenID: uuid.Must(uuid.NewV7()).String(),
		UserID:  steamID.String(),
	})
	if dbErr != nil {
		return nil, eris.Wrap(dbErr, "Error creating new session")
	}

	// There are a handful of "claims" we need to specify in the JWT. The subject and ID are the most
	// important, since they specify the user's authenticated steam ID and the token's UUID.
	expiresAt := session.CreatedAt.Add(SessionDuration)
	claims := jwt.RegisteredClaims{
		Issuer:    SessionIssuer,
		Subject:   steamID.String(),
		Audience:  []string{SessionAudience},
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		NotBefore: jwt.NewNumericDate(session.CreatedAt.Add(-SessionJitter)),
		IssuedAt:  jwt.NewNumericDate(session.CreatedAt),
		ID:        session.TokenID,
	}

	// this is creating and signing the actual JWT with the claims & secret we provide.
	signedJwt, jwtErr := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(SessionTokenSecret)
	if jwtErr != nil {
		return nil, eris.Wrap(jwtErr, "Error creating new session")
	}

	// and finally, setting the session cookie with our session JWT!
	return &CallbackOutput{
		Status: http.StatusTemporaryRedirect,
		Url:    "http://localhost:5173",
		SetCookie: http.Cookie{
			Name:     SessionCookieName,
			Path:     "/",
			Value:    signedJwt,
			MaxAge:   int(expiresAt.Sub(time.Now().UTC()).Seconds()),
			Expires:  expiresAt,
			Secure:   SessionCookieSecure,
			SameSite: http.SameSiteStrictMode,
		},
	}, nil
}

type SignOutOutput struct {
	SetCookie http.Cookie `header:"Set-Cookie"`
}

func handleSteamSignOut(ctx context.Context, _ *struct{}) (*SignOutOutput, error) {
	// if we don't have a principal, that means the user is not signed in or their session has expired.
	principal, ok := principal.Get(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("a session is required")
	}

	// but, if we do have a session, we should forcefully invalidate the session to ensure the user's token
	// can't be re-used. TokenIDs are UUIDv7, which have a time-based monotonic counter as part of the ID...
	// as a result, it's virtually impossible for the same token to be generated twice.
	err := db.Queries.InsertDisallowToken(ctx, principal.TokenID.String())
	if err != nil {
		return nil, eris.Wrap(err, "error signing out session")
	}

	// then we expire their session cookie.
	return &SignOutOutput{
		SetCookie: http.Cookie{
			Name:     SessionCookieName,
			Path:     "/",
			Value:    "",
			MaxAge:   0,
			Expires:  time.Now(),
			Secure:   SessionCookieSecure,
			SameSite: http.SameSiteStrictMode,
		},
	}, nil
}

func registerAuth(internalApi *huma.Group, sessionApi *huma.Group) {
	OidRealm = env.GetString("API_OID_REALM")
	oidRealmURL, err := url.Parse(OidRealm)
	if err != nil {
		log.Fatalf("[fatal] error parsing API_OID_REALM: %v", err)
	}

	OidRealmURL = oidRealmURL
	SessionTokenSecret = []byte(env.GetString("API_SESSION_TOKEN_SECRET"))
	SessionCookieSecure = env.GetBool("API_SESSION_COOKIE_SECURE")
	SteamApiKey = env.GetString("API_STEAM_API_KEY")

	// the OpenID flow works like this:
	// - The user is redirected to `internal/session/steam/discover`
	// - /internal/session/steam/discover does some magic & redirects the user to the Steam OpenID auth flow
	// - Once the user logs in, steam redirects the user to `internal/session/steam/callback`,
	//   with some information about the user's auth session
	// - `/internal/session/steam/callback` creates a new session token for the user
	// - the user is redirected back to home with their session cookies set
	huma.Register(internalApi, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/steam/discover",
		OperationID: "steam-discover",
		Summary:     "Steam discover",
		Description: "steam discover",
		Tags:        []string{"auth"},
	}, handleSteamDiscover)

	huma.Register(internalApi, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/steam/callback",
		OperationID: "steam-callback",
		Summary:     "Steam callback",
		Description: "steam callback",
		Tags:        []string{"auth"},
	}, handleSteamCallback)

	huma.Register(sessionApi, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/steam/profile",
		OperationID: "steam-profile",
		Summary:     "Steam profile",
		Tags:        []string{"auth"},
		Description: "get the authenticated user's steam profile info",
		Errors:      []int{http.StatusUnauthorized},

		Security: sessionCookieSecurityMap,
	}, HandleGetSteamProfile)
	//
	// huma.Register(sessionApi, huma.Operation{
	// 	Method:      http.MethodGet,
	// 	Path:        "",
	// 	OperationID: "session",
	// 	Summary:     "session profile",
	// 	Tags:        []string{"Auth"},
	// 	Description: "get the authenticated user's session profile",
	// 	Errors:      []int{http.StatusUnauthorized},
	//
	// 	Security:    sessionCookieSecurityMap,
	// 	Middlewares: requireSessionMiddlewares,
	// }, HandleGetSession)

	huma.Register(sessionApi, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/sign-out",
		OperationID: "sign-out",
		Summary:     "sign out",
		Description: "sign out & clear session",
		Tags:        []string{"auth"},
		Errors:      []int{http.StatusUnauthorized},

		Security: sessionCookieSecurityMap,
	}, handleSteamSignOut)

	huma.Register(sessionApi, huma.Operation{
		Method:      http.MethodGet,
		Path:        "",
		Tags:        []string{"auth"},
		Summary:     "get session",
		Description: "get session info",
		OperationID: "get-session",
	}, HandleGetSession)

}
