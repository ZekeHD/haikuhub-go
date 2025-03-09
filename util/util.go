package util

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"

	goaway "github.com/TwiN/go-away"
	"github.com/gin-gonic/gin"
)

const maxLimit int = 100
const maxSkip int = 100000

type ListHaikusPOST struct {
	Limit int `json:"limit"`
	Skip  int `json:"skip"`
}

func ValidateLimitAndSkip(c *gin.Context) (int, int, error) {
	var body ListHaikusPOST

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

	for _, errString := range errStrings {
		var transformed string

		if GetFailedRequiredCheck(errString) {
			transformed = GetRequiredFieldErrorString(errString)
		} else if GetFailedDuplicateCheck(errString) {
			transformed = GetDuplicateUniqueColumnErrorString(errString)
		}

		if len(transformed) > 0 {
			transformedErrorStrings = append(transformedErrorStrings, transformed)
		}

		log.Println(errString)
	}

	return transformedErrorStrings
}

func LogErrorAndSetErrorResponse(
	c *gin.Context,
	err error,
	logMessagePrefix string,
	responseMessage string,
	httpStatusCode int,
) {
	log.Println(logMessagePrefix, err.Error())

	c.JSON(httpStatusCode, gin.H{
		"error": responseMessage,
	})
}

func IsProfane(str string) bool {
	return goaway.IsProfane(str)
}
