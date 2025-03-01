package util

import (
	"fmt"
	"reflect"

	"github.com/gin-gonic/gin"
	"haikuhub.net/haikuhubapi/types"
)

const maxLimit int = 100
const maxSkip int = 100000

func ValidateLimitAndSkip(c *gin.Context) (int, int, error) {
	var body types.ListHaikusPOST

	err := c.ShouldBindBodyWithJSON(&body)
	if err != nil {
		//
	}

	fmt.Println(body.Skip)

	limitValid := reflect.TypeOf(body.Limit).Kind() == reflect.Int &&
		body.Limit >= 0 &&
		body.Limit <= maxLimit

	if !limitValid {
		err := fmt.Errorf("'limit' value needs to be number & below %d", maxLimit)

		return 0, 0, err
	}

	skipValid := reflect.TypeOf(body.Skip).Kind() == reflect.Int &&
		body.Skip >= 0 &&
		body.Skip <= maxSkip

	if !skipValid {
		err := fmt.Errorf("'skip' value needs to be number & below %d", maxSkip)

		return 0, 0, err
	}

	return body.Limit, body.Skip, nil
}
