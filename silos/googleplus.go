package silos

import (
	"os"
)

func GetGooglePlus(id string) string {
	return "http://res.cloudinary.com/" + os.Getenv("CLOUDINARY_CLOUDNAME") + "/image/gplus/" + id
}
