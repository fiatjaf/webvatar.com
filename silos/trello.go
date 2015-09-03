package silos

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func GetTrello(id string) string {
	resp, err := http.Get("https://api.trello.com/1/members/" + id + "?fields=avatarHash,gravatarHash")
	if err != nil {
		return ""
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	var data struct {
		AvatarHash   string `json:"avatarHash"`
		GravatarHash string `json:"gravatarHash"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return ""
	}

	if data.AvatarHash != "" {
		return "https://trello-avatars.s3.amazonaws.com/" + data.AvatarHash + "/170.png"
	} else if data.GravatarHash != "" {
		return "https://secure.gravatar.com/avatar/" + data.GravatarHash + "?s=200"
	}
	return ""
}
