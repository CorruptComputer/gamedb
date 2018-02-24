package steam

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/kr/pretty"
)

/**
https://partner.steamgames.com/doc/webapi/ISteamUser#GetUserGroupList
https://partner.steamgames.com/doc/webapi/ISteamUser#GetPlayerBans
https://partner.steamgames.com/doc/webapi/ISteamUser#GetAppPriceInfo
*/

func GetFriendList(id int) (friends []GetFriendListFriend, err error) {

	options := url.Values{}
	options.Set("steamid", strconv.Itoa(id))
	options.Set("relationship", "friend")

	bytes, err := get("ISteamUser/GetFriendList/v1/", options)
	if err != nil {
		return friends, err
	}

	if strings.Contains(string(bytes), "Internal Server Error") {
		return friends, errors.New("no such user")
	}

	// Unmarshal JSON
	var resp *GetFriendListBody
	if err := json.Unmarshal(bytes, &resp); err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			pretty.Print(string(bytes))
		}
		return friends, err
	}

	return resp.Friendslist.Friends, nil
}

type GetFriendListBody struct {
	Friendslist struct {
		Friends []GetFriendListFriend `json:"friends"`
	} `json:"friendslist"`
}

type GetFriendListFriend struct {
	SteamID      string `json:"steamid"`
	Relationship string `json:"relationship"`
	FriendSince  int    `json:"friend_since"`
}

func ResolveVanityURL(id string) (resp ResolveVanityURLBody, err error) {

	options := url.Values{}
	options.Set("vanityurl", id)
	options.Set("url_type", "1")

	bytes, err := get("ISteamUser/ResolveVanityURL/v1/", options)
	if err != nil {
		return resp, err
	}

	// Unmarshal JSON
	if err := json.Unmarshal(bytes, &resp); err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			pretty.Print(string(bytes))
			fmt.Println(err.Error())
		}
		return resp, err
	}

	if resp.Response.Success != 1 {
		return resp, errors.New("no user found")
	}

	return resp, nil
}

type ResolveVanityURLBody struct {
	Response ResolveVanityURLResponse
}

type ResolveVanityURLResponse struct {
	SteamID string `json:"steamid"`
	Success int8   `json:"success"`
	Message string `json:"message"`
}

// todo, only return the needed response
func GetPlayerSummaries(ids []int) (resp PlayerSummariesBody, err error) {

	if len(ids) > 100 {
		return resp, errors.New("100 ids max")
	}

	var idsString []string
	for _, v := range ids {
		idsString = append(idsString, strconv.Itoa(v))
	}

	options := url.Values{}
	options.Set("steamids", strings.Join(idsString, ","))

	bytes, err := get("ISteamUser/GetPlayerSummaries/v2/", options)
	if err != nil {
		return resp, err
	}

	// Unmarshal JSON
	if err := json.Unmarshal(bytes, &resp); err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			pretty.Print(string(bytes))
			fmt.Println(err.Error())
		}
		return resp, err
	}

	return resp, nil
}

type PlayerSummariesBody struct {
	Response PlayerSummariesResponse
}

type PlayerSummariesResponse struct {
	Players []PlayerSummariesPlayer
}

type PlayerSummariesPlayer struct {
	SteamID                  string `json:"steamid"`
	CommunityVisibilityState int8   `json:"communityvisibilitystate"`
	ProfileState             int8   `json:"profilestate"`
	PersonaName              string `json:"personaname"`
	LastLogOff               int64  `json:"lastlogoff"`
	CommentPermission        int8   `json:"commentpermission"`
	ProfileURL               string `json:"profileurl"`
	Avatar                   string `json:"avatar"`
	AvatarMedium             string `json:"avatarmedium"`
	AvatarFull               string `json:"avatarfull"`
	PersonaState             int8   `json:"personastate"`
	RealName                 string `json:"realname"`
	PrimaryClanID            string `json:"primaryclanid"`
	TimeCreated              int64  `json:"timecreated"`
	PersonaStateFlags        int8   `json:"personastateflags"`
	LOCCountryCode           string `json:"loccountrycode"`
	LOCStateCode             string `json:"locstatecode"`
}
