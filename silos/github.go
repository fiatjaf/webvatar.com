package silos

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func GetGitHub(id string) string {
	resp, err := http.Get("https://api.github.com/users/" + id)
	if err != nil {
		return ""
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	var data struct {
		Avatar string `json:"avatar_url"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return ""
	}
	return data.Avatar
}
