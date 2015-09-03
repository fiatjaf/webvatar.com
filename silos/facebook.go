package silos

func GetFacebook(id string) string {
	return "https://graph.facebook.com/v2.2/" + id + "/picture?type=large"
}
