package votes

import (
	ctx "context"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"

	"haikuhub.net/haikuhubapi/auth"
	"haikuhub.net/haikuhubapi/db"
	"haikuhub.net/haikuhubapi/sql"
	"haikuhub.net/haikuhubapi/types"
	"haikuhub.net/haikuhubapi/util"
)

type VotePOST struct {
	HaikuId   string `json:"haikuId" binding:"required"`
	Direction int8   `json:"direction" binding:"required" validate:"gte=-1,lte=1"`
}

func PostVote(c *gin.Context) {
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

	var body VotePOST
	err = c.BindJSON(&body)
	if err != nil {
		errors := strings.Split(c.Errors.Errors()[0], "\n")
		transformedErrors := util.GetTransformedErrorStrings(errors)

		c.JSON(types.HTTP_BAD, gin.H{
			"errors": transformedErrors,
		})

		return
	}

	getVoteSQL := sql.GetVoteByHaikuAndAuthor()

	voteRow := db.Pool.QueryRow(ctx.Background(), getVoteSQL, author.ID, body.HaikuId)
	vote := types.Vote{}

	getVoteErr := voteRow.Scan(
		&vote.ID,
		&vote.Upvoted,
		&vote.VotedTimestamp,
		&vote.AuthorID,
		&vote.HaikuID,
	)

	voteExists := false

	if getVoteErr == nil {
		voteExists = true
	}

	if !voteExists && body.Direction == 0 {
		c.JSON(types.HTTP_BAD, gin.H{
			"error": "vote 'direction' must be -1 or 1",
		})
	} else if voteExists && body.Direction == 0 {
		// remove vote

		sql := sql.DeleteVoteById()

		_, err := db.Pool.Exec(ctx.Background(), sql, vote.ID, vote.AuthorID)
		if err != nil {
			util.LogErrorAndSetErrorResponse(
				c,
				err,
				"unable to remove vote",
				"unable to remove vote",
				types.HTTP_INTERNAL,
			)

			return
		}

		c.Status(types.HTTP_OK_NOCONTENT)
	} else {
		// upsert vote

		sql := sql.UpsertVote()

		row := db.Pool.QueryRow(ctx.Background(), sql, body.Direction, author.ID, body.HaikuId)
		upsertedVote := types.Vote{}

		upsertErr := row.Scan(
			&upsertedVote.ID,
			&upsertedVote.Upvoted,
			&upsertedVote.VotedTimestamp,
			&upsertedVote.AuthorID,
			&upsertedVote.HaikuID,
		)

		if upsertErr != nil {
			c.JSON(types.HTTP_INTERNAL, gin.H{
				"error": "unable to process vote",
			})

			return
		}

		c.JSON(types.HTTP_OK, gin.H{
			"voteID": vote.ID,
		})
	}
}
