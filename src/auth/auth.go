package auth

type ValidateAuthHeaderResponse struct {
	Author string
	Err    string
}

func validateAuthHeader(authHeader string) string {
	// x := types.Author{
	// 	Id:      "fdsa",
	// 	Created: time.Now(),
	// }

	if len(authHeader) == 0 {
		return ""
	}

	return ""
}

// func getAuthorByAuthHeader(authHeader *string) string {
// 	return
// }
