package auth

import (
	ctx "context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"github.com/matthewhartstonge/argon2"
	haikusSQL "haikuhub.net/haikuhubapi/sql"
	"haikuhub.net/haikuhubapi/types"
)

type ValidateAuthHeaderResponse struct {
	Author string
	Err    string
}

/*
Using the supplied Base64-encoded Authorization header, verify that the "username:password"
encoded string matches an existing Author's credentials. If so, return the Author.

If not, return empty Author struct.
*/
func GetAuthorByAuthHeader(c *gin.Context, conn *pgx.Conn) (types.Author, error) {
	authHeaderRaw := c.GetHeader("Authorization")
	decoded, err := base64.StdEncoding.DecodeString(authHeaderRaw)
	if err != nil {
		return types.Author{}, fmt.Errorf("cannot decode Authorization header: %s", err.Error())
	}

	authHeaderDecoded := string(decoded)

	usernamePassword := strings.Split(authHeaderDecoded, " ")[1]
	slices := strings.Split(usernamePassword, ":")
	username := slices[0]
	password := slices[1]

	sql := haikusSQL.GetAuthorByUsername()

	author := types.Author{}
	row := conn.QueryRow(ctx.Background(), sql, username)

	err = row.Scan(
		&author.ID,
		&author.Username,
		&author.Password,
		&author.Email,
		&author.Created,
	)
	if err != nil {
		return types.Author{}, fmt.Errorf("cannot read row: %s", err.Error())
	}

	match, err := argon2.VerifyEncoded([]byte(password), author.Password)
	if err != nil {
		return types.Author{}, fmt.Errorf("cannot verify password: %s", err.Error())
	}

	if match {
		return author, nil
	} else {
		return types.Author{}, nil
	}
}
