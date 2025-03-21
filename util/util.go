package util

import (
	"context"
	"errors"
	"fmt"
	"log"
	"maps"
	"regexp"
	"slices"
	"strings"

	goaway "github.com/TwiN/go-away"
	"github.com/gin-gonic/gin"

	"haikuhub.net/haikuhubapi/internal"
	"haikuhub.net/haikuhubapi/proto"
	"haikuhub.net/haikuhubapi/types"
)

const maxLimit int = 100
const maxSkip int = 100000

const FIVE int = 5
const SEVEN int = 7

var syllablesAllowedByPhraseIndex = map[int]int32{
	0: 5,
	1: 7,
	2: 5,
}

var ListHaikuFiltersWhitelist = map[string]map[string]bool{
	"authorId": {
		types.EQ:  true,
		types.NEQ: true,
		types.IN:  true,
		types.NIN: true,
	},
	"text": {
		types.CO:  true,
		types.NCO: true,
	},
	"created": {
		types.LT: true,
		types.GT: true,
	},
	"tags": {
		types.IN:  true,
		types.NIN: true,
	},
}

func getBlankFilters() types.Filters {
	return types.Filters{}
}

func ParseListPOSTBody(c *gin.Context) (int, int, types.Filters, error) {
	var body types.ListHaikusPOST

	err := c.BindJSON(&body)
	if err != nil {
		return 0, 0, getBlankFilters(), err
	}

	limitValid := body.Limit >= 0 && body.Limit <= maxLimit
	if !limitValid {
		err := fmt.Errorf("'limit' value needs to be number between 0 - %d", maxLimit)

		return 0, 0, getBlankFilters(), err
	}

	skipValid := body.Skip >= 0 && body.Skip <= maxSkip
	if !skipValid {
		err := fmt.Errorf("'skip' value needs to be number between 0 - %d", maxSkip)

		return 0, 0, getBlankFilters(), err
	}

	invalidFiltersMessage := validateFilters(body.Filters)
	if len(invalidFiltersMessage) != 0 {
		return 0, 0, getBlankFilters(), errors.New(invalidFiltersMessage)
	}

	return body.Limit, body.Skip, body.Filters, nil
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

func GetFailedFieldUnmarshal(errString string) bool {
	return strings.Contains(errString, "json: cannot unmarshal")
}

func GetFailedFieldUnmarshalErrorString(errString string) string {
	errorRegex := regexp.MustCompile(`cannot unmarshal .* into Go struct field .*\.(.{1,}) of type (.{1,})`)
	matches := errorRegex.FindStringSubmatch(errString)
	field := matches[1]

	if field == "filters" {
		return fmt.Sprintf("field '%s' should adhere to a 'field: { <operator>: <value> }' format", field)
	}

	correctType := matches[2]

	transformedErrorString := fmt.Sprintf("field '%s' requires '%s' value type", field, correctType)

	return transformedErrorString
}

func GetDuplicateUniqueColumnErrorString(errString string) string {
	errorRegex := regexp.MustCompile(`"(.{1,})_unique"`)
	uniqueField := errorRegex.FindStringSubmatch(errString)[1]

	return fmt.Sprintf("%s already taken!", uniqueField)
}

func GetTransformedErrorStrings(errStrings []string) []string {
	transformedErrorStrings := []string{}

	for _, errString := range errStrings {
		transformed := errString

		if GetFailedRequiredCheck(errString) {
			transformed = GetRequiredFieldErrorString(errString)
		} else if GetFailedDuplicateCheck(errString) {
			transformed = GetDuplicateUniqueColumnErrorString(errString)
		} else if GetFailedFieldUnmarshal(errString) {
			transformed = GetFailedFieldUnmarshalErrorString(errString)
		}

		transformedErrorStrings = append(transformedErrorStrings, transformed)

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

func validateFilters(filters types.Filters) string {
	for field, expression := range filters {
		allowedOperators := ListHaikuFiltersWhitelist[field]

		if len(allowedOperators) == 0 {
			return fmt.Sprintf("invalid field '%s'", field)
		}

		operator := slices.Collect(maps.Keys(expression))[0]
		if !allowedOperators[operator] {
			return fmt.Sprintf("invalid operator '%s' for field '%s'", operator, field)
		}
	}

	return ""
}

func getPhraseInvalidMsg(phrase string, allowedSyllables int32) string {
	whitespaceRegex := regexp.MustCompile(`^[^\s].+[^\s]$`)
	surroundingWhitespaces := !whitespaceRegex.Match([]byte(phrase))
	if surroundingWhitespaces {
		return fmt.Sprintf("remove all surrounding whitespace characters from phrase '%s'", phrase)
	}

	client := proto.NewSyllablesServiceClient(internal.GrpcClientConn)
	rpcResp, _ := client.GetSyllables(context.Background(), &proto.GetSyllablesRequest{Input: phrase})
	syllablesCount := rpcResp.SyllablesCount

	if syllablesCount != allowedSyllables {
		return fmt.Sprintf("phrase '%s' needs to be %d syllables, detected %d", phrase, allowedSyllables, syllablesCount)
	}

	return ""
}

func ValidateHaiku(haiku string) string {
	phrases := strings.Split(haiku, "//")

	if len(phrases) != 3 {
		return "haiku must be in 5, 7, 5 syllable format, with each phrase separated by '//' characters"
	}

	for i, phrase := range phrases {
		allowedSyllables := syllablesAllowedByPhraseIndex[i]
		phraseInvalidMsg := getPhraseInvalidMsg(phrase, allowedSyllables)

		if len(phraseInvalidMsg) != 0 {
			return phraseInvalidMsg
		}
	}

	return ""
}
