package authors

import (
	ctx "context"
	"net/mail"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/matthewhartstonge/argon2"
	passwordvalidator "github.com/wagslane/go-password-validator"

	"haikuhub.net/haikuhubapi/db"
	"haikuhub.net/haikuhubapi/sql"
	"haikuhub.net/haikuhubapi/types"
	"haikuhub.net/haikuhubapi/util"
)

const MIN_PW_ENTROPHY = 60

var usernameRegex = regexp.MustCompile(`^[\w\d-_]+$`)

func RegisterAuthor(c *gin.Context) {
	var body types.RegisterAuthorPOST

	err := c.BindJSON(&body)
	if err != nil {
		errors := strings.Split(c.Errors.Errors()[0], "\n")

		transformedErrors := util.GetTransformedErrorStrings(errors)

		c.JSON(types.HTTP_BAD, gin.H{
			"errors": transformedErrors,
		})

		return
	}

	usernameValid := IsUsernameValid(body.Username)
	if !usernameValid {
		c.JSON(types.HTTP_BAD, gin.H{
			"error": "illegal characters detected in 'username' field. Alphanumeric characters and '-', '_' allowed",
		})

		return
	}

	usernameProfane := util.IsProfane(body.Username)
	if usernameProfane {
		c.JSON(types.HTTP_BAD, gin.H{
			"error": "profanity detected in 'username' field",
		})

		return
	}

	var emailAddress string = ""

	if len(body.Email) > 0 {
		address, err := mail.ParseAddress(body.Email)
		emailAddress = address.Address

		if err != nil {
			c.JSON(types.HTTP_BAD, gin.H{
				"error": "request body field 'email' must be a standard email address",
			})

			return
		}
	}

	passwordInvalidErr := passwordvalidator.Validate(body.Password, MIN_PW_ENTROPHY)
	if passwordInvalidErr != nil {
		c.JSON(types.HTTP_BAD, gin.H{
			"error": passwordInvalidErr.Error(),
		})

		return
	}

	argon := argon2.DefaultConfig()
	encoded, err := argon.HashEncoded([]byte(body.Password))
	if err != nil {
		util.LogErrorAndSetErrorResponse(
			c,
			err,
			"error encoding password",
			"cannot encode password",
			types.HTTP_INTERNAL,
		)

		return
	}

	sql := sql.InsertAuthor()

	_, err = db.Pool.Exec(ctx.Background(), sql, body.Username, encoded, emailAddress)
	if err != nil {
		errString := err.Error()

		if util.GetFailedDuplicateCheck(errString) {
			c.JSON(types.HTTP_BAD, gin.H{
				"error": util.GetDuplicateUniqueColumnErrorString(errString),
			})
		} else {
			c.JSON(types.HTTP_INTERNAL, gin.H{
				"error": err.Error(),
			})
		}

		return
	}

	c.JSON(types.HTTP_OK, gin.H{
		"username": body.Username,
		"email":    emailAddress,
	})
}

func IsUsernameValid(username string) bool {
	return usernameRegex.MatchString(username)
}
