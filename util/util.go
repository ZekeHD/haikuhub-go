package util

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"

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

func GetFailedRequiredCheck(errString string) bool {
	return strings.Contains(errString, "failed on the 'required' tag")
}

func GetRequiredFieldErrorString(errString string) string {
	errorRegex := regexp.MustCompile("Field validation for '(.{1,})' failed on the 'required' tag")
	errorField := errorRegex.FindStringSubmatch(errString)[1]

	return fmt.Sprintf("request body requires a non-zero length '%s' field", strings.ToLower((errorField)))
}

func GetFailedDuplicateCheck(errString string) bool {
	return strings.Contains(errString, "duplicate key value violates unique constraint")
}

func GetDuplicateUniqueColumnErrorString(errString string) string {
	errorRegex := regexp.MustCompile(`"(.{1,})_unique"`)
	uniqueField := errorRegex.FindStringSubmatch(errString)[1]

	return fmt.Sprintf("%s already taken!", uniqueField)
}

func GetTransformedErrorStrings(errStrings []string) []string {
	transformedErrorStrings := []string{}

	for _, err := range errStrings {
		var transformed string

		if GetFailedRequiredCheck(err) {
			transformed = GetRequiredFieldErrorString(err)
		} else if GetFailedDuplicateCheck(err) {
			transformed = GetDuplicateUniqueColumnErrorString(err)
		}

		if len(transformed) > 0 {
			transformedErrorStrings = append(transformedErrorStrings, transformed)
		}

		log.Println(err)
	}

	return transformedErrorStrings
}

func LogAndAbortRequest(
	c *gin.Context,
	err error,
	logMessagePrefix string,
	responseMessage string,
	httpStatusCode int,
) {
	log.Println(logMessagePrefix, err.Error())

	c.AbortWithStatusJSON(httpStatusCode, gin.H{
		"error": responseMessage,
	})
}
