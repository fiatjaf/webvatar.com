package silos

import (
	"os"
	"strconv"
)

func GetTwitter(id string) string {
	var idkind string
	if _, err := strconv.Atoi(id); err == nil {
		idkind = "twitter"
	} else {
		idkind = "twitter_name"
	}
	return "http://res.cloudinary.com/" + os.Getenv("CLOUDINARY_CLOUDNAME") + "/image/" + idkind + "/w_200/" + id
}
