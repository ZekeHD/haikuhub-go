package auth

import (
	ctx "context"
	"encoding/base64"
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/matthewhartstonge/argon2"
	"haikuhub.net/haikuhubapi/db"
	"haikuhub.net/haikuhubapi/sql"
	"haikuhub.net/haikuhubapi/types"
)

type ValidateAuthHeaderResponse struct {
	Author string
	Err    string
}

/*
Using the supplied Authorization header, verify that the "username:password"
encoded string matches an existing Author's credentials. If so, return the Author.

Typically formatted: "Basic dXNlcm5hbWU6cGFzc3dvcmQ=" -> "Basic username:password"

If not, return empty Author struct.
*/
func GetAuthorByAuthHeader(c *gin.Context) (types.Author, error) {
	authHeaderRaw := c.GetHeader("Authorization")
	encodedCredentials := strings.Split(authHeaderRaw, " ")[1]

	decoded, err := base64.StdEncoding.DecodeString(encodedCredentials)
	if err != nil {
		return types.Author{}, fmt.Errorf("cannot decode Authorization header: %s", err.Error())
	}

	decodedCredentials := string(decoded)
	slices := strings.Split(decodedCredentials, ":")
	username := slices[0]
	password := slices[1]

	sql := sql.GetAuthorByUsername()

	author := types.Author{}
	row := db.Pool.QueryRow(ctx.Background(), sql, username)

	err = row.Scan(
		&author.ID,
		&author.Username,
		&author.Password,
		&author.Email,
		&author.Created,
	)
	if err != nil {
		log.Printf("error while scanning Row: %s", err.Error())
	}

	match, err := argon2.VerifyEncoded([]byte(password), author.Password)
	if err != nil {
		log.Printf("cannot verify password: %s", err.Error())
	}

	if match {
		return author, nil
	} else {
		return types.Author{}, nil
	}
}
