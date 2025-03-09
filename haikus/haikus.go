package haikus

import (
	ctx "context"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"haikuhub.net/haikuhubapi/auth"
	"haikuhub.net/haikuhubapi/db"
	"haikuhub.net/haikuhubapi/sql"
	"haikuhub.net/haikuhubapi/types"
	"haikuhub.net/haikuhubapi/util"
)

type HaikuPUT struct {
	Text string `json:"text" binding:"required"`
	Tags string `json:"tags"`
}

// TODO: go back thru and refactor all error responses to use c.Error, like below?

func ListAllHaikus(c *gin.Context) {
	limit, skip, err := util.ValidateLimitAndSkip(c)
	if err != nil {
		c.Error(err)
		c.JSON(types.HTTP_BAD, c.Errors.JSON())

		return
	}

	sql := sql.ListAllHaikus()

	rows, err := db.Pool.Query(ctx.Background(), sql, limit, skip)
	if err != nil {
		util.LogErrorAndSetErrorResponse(
			c,
			err,
			"list all haikus failed",
			"unable to list haikus",
			types.HTTP_INTERNAL,
		)

		return
	}

	haikus, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.Haiku])
	if err != nil {
		util.LogErrorAndSetErrorResponse(
			c,
			err,
			"list all failed",
			"unable to read DB haikus response",
			types.HTTP_INTERNAL,
		)

		return
	}

	c.JSON(types.HTTP_OK, gin.H{
		"haikus": haikus,
	})
}

func GetHaikuById(c *gin.Context) {
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

		c.JSON(types.HTTP_NOTFOUND, gin.H{
			"error": errMessage,
		})

		return
	}

	c.JSON(types.HTTP_OK, gin.H{
		"haiku": haiku,
	})
}

func PutHaiku(c *gin.Context) {
	author, err := auth.GetAuthorByAuthHeader(c)
	if reflect.ValueOf(author).IsZero() {
		var errorMessage string = "unauthorized"
		if err != nil {
			errorMessage = err.Error()
		}

		c.JSON(types.HTTP_UNAUTHORIZED, gin.H{
			"error": errorMessage,
		})

		return
	}

	var body HaikuPUT
	err = c.BindJSON(&body)
	if err != nil {
		errors := strings.Split(c.Errors.Errors()[0], "\n")
		transformedErrors := util.GetTransformedErrorStrings(errors)

		c.JSON(types.HTTP_BAD, gin.H{
			"errors": transformedErrors,
		})

		return
	}

	sql := sql.InsertHaiku()

	row := db.Pool.QueryRow(ctx.Background(), sql, body.Text, body.Tags, 0, author.ID)
	insertedHaiku := types.Haiku{}

	insertErr := row.Scan(
		&insertedHaiku.ID,
		&insertedHaiku.Text,
		&insertedHaiku.Tags,
		&insertedHaiku.Rating,
		&insertedHaiku.Created,
		&insertedHaiku.AuthorID,
	)

	if insertErr != nil {
		c.JSON(types.HTTP_INTERNAL, gin.H{
			"error": insertErr.Error(),
		})

		return
	}

	c.JSON(types.HTTP_OK, gin.H{
		"haiku": insertedHaiku,
	})
}

func DeleteHaikuById(c *gin.Context) {
	author, err := auth.GetAuthorByAuthHeader(c)
	if reflect.ValueOf(author).IsZero() {
		var errorMessage string = "unauthorized"
		if err != nil {
			errorMessage = err.Error()
		}

		c.JSON(types.HTTP_UNAUTHORIZED, gin.H{
			"error": errorMessage,
		})

		return
	}

	haikuId := c.Param("id")

	sql := sql.DeleteHaikuById()

	cmd, err := db.Pool.Exec(ctx.Background(), sql, haikuId, author.ID)
	if err != nil {
		util.LogErrorAndSetErrorResponse(
			c,
			err,
			"unable to delete haiku",
			"unable to delete haiku",
			types.HTTP_INTERNAL,
		)

		return
	}

	if cmd.RowsAffected() == 0 {
		c.JSON(types.HTTP_BAD, gin.H{
			"error": fmt.Sprintf("Haiku with ID '%s' not found", haikuId),
		})

		return
	}

	c.Status(types.HTTP_OK)
}
