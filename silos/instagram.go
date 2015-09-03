package silos

import (
	"os"
	"strconv"
)

func GetInstagram(id string) string {
	var idkind string
	if _, err := strconv.Atoi(id); err == nil {
		idkind = "instagram"
	} else {
		idkind = "instagram_name"
	}
	return "http://res.cloudinary.com/" + os.Getenv("CLOUDINARY_CLOUDNAME") + "/image/" + idkind + "/w_200/" + id
}
