package main

import (
	ctx "context"
	"fmt"
	"log"
	"net/mail"
	"reflect"
	"strings"

	"github.com/jackc/pgx/v5"
	"haikuhub.net/haikuhubapi/auth"
	"haikuhub.net/haikuhubapi/db"
	"haikuhub.net/haikuhubapi/sql"
	"haikuhub.net/haikuhubapi/types"
	"haikuhub.net/haikuhubapi/util"

	"github.com/joho/godotenv"

	"github.com/gin-gonic/gin"
	"github.com/matthewhartstonge/argon2"
)

func main() {
	envLoadErr := godotenv.Load()
	if envLoadErr != nil {
		log.Fatal("Error loading env file")
	}

	db.InitializeTables()

	r := gin.Default()
	r.HandleMethodNotAllowed = true

	r.POST("/allHaikus", listAllHaikus)
	r.GET("/haiku/:id", getHaikuById)
	r.PUT("/haiku", putHaiku)
	r.DELETE("/haiku/:id", deleteHaikuById)

	r.PUT("/author", registerAuthor)

	r.GET("/haiku", func(c *gin.Context) {
		c.AbortWithStatusJSON(types.HTTP_BAD, gin.H{
			"error": "haiku id parameter required: GET https://haikuhub.net/haikus/121",
		})
	})

	r.DELETE("/haiku", func(c *gin.Context) {
		c.AbortWithStatusJSON(types.HTTP_BAD, gin.H{
			"error": "haiku id parameter required: DELETE https://haikuhub.net/haikus/121",
		})
	})

	r.Run()
}

// TODO: go back thru and refactor all error responses to use c.Error, like below?

func listAllHaikus(c *gin.Context) {
	limit, skip, err := util.ValidateLimitAndSkip(c)
	if err != nil {
		c.Error(err)
		c.JSON(types.HTTP_BAD, c.Errors.JSON())

		return
	}

	sql := sql.ListAllHaikus()

	rows, err := db.Pool.Query(ctx.Background(), sql, limit, skip)
	if err != nil {
		util.LogAndAbortRequest(
			c,
			err,
			"list all haikus failed",
			"unable to list haikus",
			types.HTTP_INTERNAL,
		)
	}

	haikus, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.Haiku])
	if err != nil {
		util.LogAndAbortRequest(
			c,
			err,
			"list all failed",
			"unable to read DB haikus response",
			types.HTTP_INTERNAL,
		)
	}

	c.JSON(types.HTTP_OK, gin.H{
		"haikus": haikus,
	})
}

func getHaikuById(c *gin.Context) {
	haikuId := c.Param("id")

	sql := sql.GetHaikuById()

	row := db.Pool.QueryRow(ctx.Background(), sql, haikuId)

	haiku := types.Haiku{}
	err := row.Scan(
		&haiku.ID,
		&haiku.Text,
		&haiku.Tags,
		&haiku.Rating,
		&haiku.AuthorID,
		&haiku.Created,
	)

	if err != nil {
		errMessage := "unable to find haiku"

		if err.Error() == "no rows in result set" {
			errMessage = "haiku not found!"
		} else {
			log.Println(errMessage, err.Error())
		}

		c.AbortWithStatusJSON(types.HTTP_NOTFOUND, gin.H{
			"error": errMessage,
		})
	}

	c.JSON(types.HTTP_OK, gin.H{
		"haiku": haiku,
	})
}

func putHaiku(c *gin.Context) {
	author, err := auth.GetAuthorByAuthHeader(c)
	if reflect.ValueOf(author).IsZero() {
		var errorMessage string = "unauthorized"
		if err != nil {
			errorMessage = err.Error()
		}

		c.AbortWithStatusJSON(types.HTTP_UNAUTHORIZED, gin.H{
			"error": errorMessage,
		})
	}

	var body types.HaikuPUT
	err = c.BindJSON(&body)
	if err != nil {
		errors := strings.Split(c.Errors.Errors()[0], "\n")
		transformedErrors := util.GetTransformedErrorStrings(errors)

		c.AbortWithStatusJSON(types.HTTP_BAD, gin.H{
			"errors": transformedErrors,
		})
	}

	sql := sql.InsertHaiku()

	row := db.Pool.QueryRow(ctx.Background(), sql, body.Text, body.Tags, 0, "the-authors-id")
	insertedHaiku := types.Haiku{}

	insertErr := row.Scan(
		&insertedHaiku.ID,
		&insertedHaiku.Text,
		&insertedHaiku.Tags,
		&insertedHaiku.Rating,
		&insertedHaiku.AuthorID,
		&insertedHaiku.Created,
	)

	if insertErr != nil {
		c.AbortWithStatusJSON(types.HTTP_INTERNAL, gin.H{
			"error": insertErr.Error(),
		})
	}

	c.JSON(types.HTTP_OK, gin.H{
		"haiku": insertedHaiku,
	})
}

func deleteHaikuById(c *gin.Context) {
	author, err := auth.GetAuthorByAuthHeader(c)
	if reflect.ValueOf(author).IsZero() {
		var errorMessage string = "unauthorized"
		if err != nil {
			errorMessage = err.Error()
		}

		c.AbortWithStatusJSON(types.HTTP_UNAUTHORIZED, gin.H{
			"error": errorMessage,
		})
	}

	haikuId := c.Param("id")

	sql := sql.DeleteHaikuById()

	cmd, err := db.Pool.Exec(ctx.Background(), sql, haikuId, author.ID)
	if err != nil {
		util.LogAndAbortRequest(
			c,
			err,
			"unable to delete haiku",
			"unable to delete haiku",
			types.HTTP_INTERNAL,
		)
	}

	if cmd.RowsAffected() == 0 {
		c.AbortWithStatusJSON(types.HTTP_BAD, gin.H{
			"error": fmt.Sprintf("Haiku with ID '%s' not found", haikuId),
		})
	}

	c.Status(types.HTTP_OK)
}

func registerAuthor(c *gin.Context) {
	var body types.RegisterAuthorPOST

	err := c.BindJSON(&body)
	if err != nil {
		errors := strings.Split(c.Errors.Errors()[0], "\n")

		transformedErrors := util.GetTransformedErrorStrings(errors)

		c.AbortWithStatusJSON(types.HTTP_BAD, gin.H{
			"errors": transformedErrors,
		})
	}

	email, err := mail.ParseAddress(body.Email)
	if err != nil {
		c.AbortWithStatusJSON(types.HTTP_BAD, gin.H{
			"error": "request body field 'email' must be a standard email address",
		})
	}

	argon := argon2.DefaultConfig()
	encoded, err := argon.HashEncoded([]byte(body.Password))
	if err != nil {
		c.AbortWithStatusJSON(types.HTTP_INTERNAL, gin.H{
			"error": err.Error(),
		})
	}

	sql := sql.InsertAuthor()

	_, err = db.Pool.Exec(ctx.Background(), sql, body.Username, encoded, email.Address)
	if err != nil {
		errString := err.Error()

		if util.GetFailedDuplicateCheck(errString) {
			c.AbortWithStatusJSON(types.HTTP_BAD, gin.H{
				"error": util.GetDuplicateUniqueColumnErrorString(errString),
			})
		} else {
			c.AbortWithStatusJSON(types.HTTP_INTERNAL, gin.H{
				"error": err.Error(),
			})
		}
	}

	c.JSON(types.HTTP_OK, gin.H{
		"username": body.Username,
		"email":    email.Address,
	})
}
