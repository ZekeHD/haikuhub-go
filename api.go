package main

import (
	ctx "context"
	"fmt"
	"net/mail"
	"os"
	"reflect"
	"strings"

	"github.com/jackc/pgx/v5"
	auth "haikuhub.net/haikuhubapi/auth"
	haikusSQL "haikuhub.net/haikuhubapi/sql"
	"haikuhub.net/haikuhubapi/types"
	"haikuhub.net/haikuhubapi/util"

	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/matthewhartstonge/argon2"
)

func main() {
	r := gin.Default()

	r.POST("/allHaikus", listAllHaikus)
	r.GET("/haiku/:id", getHaikuById)
	r.PUT("/haiku", putHaiku)
	r.DELETE("/haiku/:id", deleteHaikuById)

	r.POST("/registerAuthor", registerAuthor)
	// r.POST("/login", auth.Login)

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

	sql := haikusSQL.ListAllHaikus()

	conn := getPostgresConn()
	defer conn.Close(ctx.Background())

	rows, err := conn.Query(ctx.Background(), sql, limit, skip)
	if err != nil {
		fmt.Println("list all failed", err)
		return
	}

	haikus, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.Haiku])
	if err != nil {
		c.JSON(types.HTTP_INTERNAL, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(types.HTTP_OK, gin.H{
		"haikus": haikus,
	})
}

func getHaikuById(c *gin.Context) {
	haikuId := c.Param("id")

	sql := haikusSQL.GetHaikuById()

	conn := getPostgresConn()
	defer conn.Close(ctx.Background())

	row := conn.QueryRow(ctx.Background(), sql, haikuId)

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
		errMessage := err.Error()

		if err.Error() == "no rows in result set" {
			errMessage = "haiku not found!"
		}

		c.JSON(types.HTTP_BAD, gin.H{
			"error": errMessage,
		})

		return
	}

	c.JSON(types.HTTP_OK, gin.H{
		"haiku": haiku,
	})
}

func putHaiku(c *gin.Context) {
	conn := getPostgresConn()
	defer conn.Close(ctx.Background())

	author, err := auth.GetAuthorByAuthHeader(c, conn)
	if reflect.ValueOf(author).IsZero() && err == nil {
		c.JSON(types.HTTP_UNAUTHORIZED, gin.H{
			"message": "unauthorized",
		})

		return
	} else if err != nil {
		c.JSON(types.HTTP_INTERNAL, gin.H{
			"error": fmt.Sprintf("something bad happened: %s", err.Error()),
		})

		return
	}

	var body types.HaikuPUT
	err = c.ShouldBindBodyWithJSON(&body)
	if err != nil {
		c.JSON(types.HTTP_BAD, gin.H{"response": err})

		return
	}

	sql := haikusSQL.InsertHaiku()

	row := conn.QueryRow(ctx.Background(), sql, body.Text, body.Tags, 0, "the-authors-id")
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
		c.JSON(types.HTTP_INTERNAL, gin.H{
			"error": insertErr.Error(),
		})

		return
	}

	c.JSON(types.HTTP_OK, gin.H{
		"haiku": insertedHaiku,
	})
}

func deleteHaikuById(c *gin.Context) {
	conn := getPostgresConn()
	defer conn.Close(ctx.Background())

	author, err := auth.GetAuthorByAuthHeader(c, conn)
	if reflect.ValueOf(author).IsZero() && err == nil {
		c.JSON(types.HTTP_UNAUTHORIZED, gin.H{
			"message": "unauthorized",
		})

		return
	} else if err != nil {
		c.JSON(types.HTTP_INTERNAL, gin.H{
			"error": fmt.Sprintf("something bad happened: %s", err.Error()),
		})

		return
	}

	haikuId := c.Param("id")

	sql := haikusSQL.DeleteHaikuById()

	cmd, err := conn.Exec(ctx.Background(), sql, haikuId, author.ID)
	if err != nil {
		c.JSON(types.HTTP_INTERNAL, gin.H{
			"error": fmt.Sprintf("unable to delete haiku: %s", err.Error()),
		})
	}

	if cmd.RowsAffected() == 0 {
		c.JSON(types.HTTP_BAD, gin.H{
			"error": fmt.Sprintf("Haiku with ID '%s' not found", haikuId),
		})
	}

	c.Status(types.HTTP_OK)
}

func getPostgresConn() *pgx.Conn {
	databaseUrl := os.Getenv("DATABASE_URL")
	conn, err := pgx.Connect(ctx.Background(), databaseUrl)
	if err != nil {
		fmt.Println("FDSAA")
	}

	return conn
}

func registerAuthor(c *gin.Context) {
	var body types.RegisterAuthorPOST

	err := c.ShouldBindBodyWithJSON(&body)
	if err != nil {
		c.JSON(types.HTTP_BAD, gin.H{
			"response": err,
		})

		return
	}

	email, err := mail.ParseAddress(body.Email)
	if err != nil {
		c.JSON(types.HTTP_BAD, gin.H{
			"error": "Invalid email address!",
		})

		return
	}

	conn := getPostgresConn()
	defer conn.Close(ctx.Background())

	argon := argon2.DefaultConfig()
	encoded, err := argon.HashEncoded([]byte(body.Password))
	if err != nil {
		c.JSON(types.HTTP_INTERNAL, gin.H{
			"error": err.Error(),
		})

		return
	}

	sql := haikusSQL.InsertAuthor()

	_, err = conn.Exec(ctx.Background(), sql, body.Username, encoded, email.Address)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			errorRegex := regexp.MustCompile(`"(.{1,})_unique"`)
			uniqueField := errorRegex.FindStringSubmatch(err.Error())[1]

			c.JSON(types.HTTP_BAD, gin.H{
				"error": fmt.Sprintf("%s already taken!", uniqueField),
			})
		} else {
			c.JSON(types.HTTP_INTERNAL, gin.H{
				"error": err.Error(),
			})
		}

		return
	}

	c.JSON(types.HTTP_OK, gin.H{
		"username":        body.Username,
		"email":           email.Address,
		"encodedPassword": string(encoded),
	})
}
