package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"

	"github.com/danielgtaylor/huma/v2"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/rotisserie/eris"
	"github.com/spiritov/svelte-go-steam/internal/principal"
)

type SteamProfile struct {
	SteamID         string `json:"steamId"`
	PersonaName     string `json:"personaName"`
	ProfileURL      string `json:"profileUrl"`
	AvatarURL       string `json:"avatarUrl"`
	AvatarMediumURL string `json:"avatarMediumUrl"`
	AvatarFullURL   string `json:"avatarFullUrl"`
}

func SteamProfileFromSummary(summary PlayerSummary) SteamProfile {
	return SteamProfile{
		SteamID:         summary.SteamID,
		PersonaName:     summary.PersonaName,
		ProfileURL:      summary.ProfileURL,
		AvatarURL:       summary.AvatarURL,
		AvatarMediumURL: summary.AvatarMediumURL,
		AvatarFullURL:   summary.AvatarFullURL,
	}
}

type SteamProfileOutput struct {
	Body SteamProfile
}

type PlayerSummary struct {
	SteamID                  string `json:"steamid"`
	CommunityVisibilityState int    `json:"communityvisibilitystate"`
	ProfileState             int    `json:"profilestate"`
	PersonaName              string `json:"personaname"`
	CommentPermission        int    `json:"commentpermission"`
	ProfileURL               string `json:"profileurl"`
	AvatarURL                string `json:"avatar"`
	AvatarMediumURL          string `json:"avatarmedium"`
	AvatarFullURL            string `json:"avatarfull"`
	AvatarHash               string `json:"avatarhash"`
	PersonaState             int    `json:"personastate"`
	PrimaryClanID            string `json:"primaryclanid"`
	TimeCreated              int    `json:"timecreated"`
	PersonaStateFlags        int    `json:"personastateflags"`
	CountryCode              string `json:"loccountrycode"`
}

type PlayerSummaries struct {
	Response struct {
		Players []PlayerSummary
	}
}

func FetchProfileSummary(steamId uint64) (*PlayerSummary, error) {
	// the ISteamUser::GetPlayerSummaries API requires a steam API key...
	encodedKey := url.QueryEscape(SteamApiKey)
	encodedSteamID := url.QueryEscape(strconv.FormatUint(steamId, 10))
	summaryUrl := fmt.Sprintf("https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2?key=%s&steamids=%s", encodedKey, encodedSteamID)

	// retryablehttp uses an exponential backoff by default. If the first request fails, it will retry
	// continuously with longer and longer periods between each retry, to avoid rate limiting.
	httpResponse, err := retryablehttp.Get(summaryUrl)
	if err != nil {
		return nil, eris.Wrap(err, "Error getting steam player summaries")
	}

	// just reading the response body has bytes.
	bodyBytes, readErr := io.ReadAll(httpResponse.Body)
	if readErr != nil {
		return nil, eris.Wrap(readErr, "Error reading steam player summary response body")
	}

	// parsing the body, assuming that it is valid unicode JSON.
	var summaries PlayerSummaries
	jsonErr := json.Unmarshal(bodyBytes, &summaries)
	if jsonErr != nil {
		return nil, eris.Wrap(jsonErr, "Error parsing player summary response body")
	}

	// we only provided one steamID in the request, we should only get one in the response.
	if len(summaries.Response.Players) != 1 {
		return nil, huma.Error500InternalServerError("Unexpected player summaries in steam response")
	}

	return &summaries.Response.Players[0], nil
}

func HandleGetSteamProfile(ctx context.Context, _ *struct{}) (*SteamProfileOutput, error) {
	// if we don't have a principal, that means the user is not signed in or their session has expired.
	principal, ok := principal.Get(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("a session is required")
	}

	summary, err := FetchProfileSummary(principal.SteamID.ID())
	if err != nil {
		return nil, err
	}

	// map the steam response to only the fields we care about.
	return &SteamProfileOutput{
		Body: SteamProfileFromSummary(*summary),
	}, nil
}
