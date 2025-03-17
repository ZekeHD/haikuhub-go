package haikus_test

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"haikuhub.net/haikuhubapi/api"
	"haikuhub.net/haikuhubapi/db"
	"haikuhub.net/haikuhubapi/types"
)

type TestCase struct {
	valueTypeString string
	value           any
	statusCode      int
	errorMsg        string
}

type HaikuTestCase struct {
	scenarioString string
	haiku          string
	errorMsg       string
}

const HaikuIncompleteErrMsg = "haiku must be in 5, 7, 5 syllable format, with each phrase separated by '//' characters"

func UnmarshalResponseBody(body io.ReadCloser) map[string]any {
	bodyBytes, _ := io.ReadAll(body)
	responseBody := make(map[string]any)
	json.Unmarshal(bodyBytes, &responseBody)

	return responseBody
}

func marshalMapToBody(body map[string]any) *strings.Reader {
	bodyJSON, _ := json.Marshal(body)
	jsonBytes := strings.NewReader(string(bodyJSON))

	return jsonBytes
}

func convertToBase64(str string) string {
	enc := b64.StdEncoding.EncodeToString([]byte(str))

	return enc
}

func GetLimitSkipTestCases(field string, maxValue int) []TestCase {
	return []TestCase{
		{valueTypeString: "string", value: "fdsa", statusCode: 400, errorMsg: fmt.Sprintf("field '%s' requires 'int' value type", field)},
		{valueTypeString: "string", value: "50", statusCode: 400, errorMsg: fmt.Sprintf("field '%s' requires 'int' value type", field)},
		{valueTypeString: "object/struct", value: map[string]any{"value": 25}, statusCode: 400, errorMsg: fmt.Sprintf("field '%s' requires 'int' value type", field)},
		{valueTypeString: "boolean", value: false, statusCode: 400, errorMsg: fmt.Sprintf("field '%s' requires 'int' value type", field)},
		{valueTypeString: "negative integer", value: -25, statusCode: 400, errorMsg: fmt.Sprintf("'%s' value needs to be number between 0 - %d", field, maxValue)},
	}
}

func TestHaikus(t *testing.T) {
	BeforeSuite(func() {
		t.Setenv("DATABASE_SCHEMA", "test")
		db.InitializeTables()
	})

	AfterSuite(func() {
		db.DropTestTables()
	})

	router := api.GetRouter()

	Describe("POST /listHaikus", func() {
		Describe("POST Body Validation", func() {
			Describe("Invalid cases", func() {
				Context("Limit", func() {
					limitTestCases := GetLimitSkipTestCases("limit", 100)

					for _, testCase := range limitTestCases {
						scenarioString := fmt.Sprintf("should not accept a %s 'limit' value", testCase.valueTypeString)

						It(scenarioString, func() {
							body := map[string]any{
								"limit": testCase.value,
							}

							jsonBytes := marshalMapToBody(body)

							req, _ := http.NewRequest("POST", "/listHaikus", jsonBytes)
							w := httptest.NewRecorder()
							router.ServeHTTP(w, req)
							responseBody := UnmarshalResponseBody(w.Result().Body)

							Expect(w.Result().StatusCode).Should(Equal(400))
							Expect(responseBody["errors"]).Should(ContainElement(ContainSubstring(testCase.errorMsg)))
						})
					}
				})

				Context("Skip", func() {
					skipTestCases := GetLimitSkipTestCases("skip", 100000)

					for _, testCase := range skipTestCases {
						scenarioString := fmt.Sprintf("should not accept a %s 'skip' value", testCase.valueTypeString)

						It(scenarioString, func() {
							body := map[string]any{
								"skip": testCase.value,
							}

							jsonBytes := marshalMapToBody(body)

							req, _ := http.NewRequest("POST", "/listHaikus", jsonBytes)
							w := httptest.NewRecorder()
							router.ServeHTTP(w, req)
							responseBody := UnmarshalResponseBody(w.Result().Body)

							Expect(w.Result().StatusCode).Should(Equal(400))
							Expect(responseBody["errors"]).Should(ContainElement(ContainSubstring(testCase.errorMsg)))
						})
					}
				})

				Context("Filters", func() {
					testCases := []TestCase{
						{
							valueTypeString: "string",
							value:           map[string]any{"authorId": "fdsa"},
							statusCode:      400,
							errorMsg:        "field 'filters' should adhere to a 'field: { <operator>: <value> }' format",
						},
						{
							valueTypeString: "integer",
							value:           map[string]any{"authorId": 32},
							statusCode:      400,
							errorMsg:        "field 'filters' should adhere to a 'field: { <operator>: <value> }' format",
						},
						{
							valueTypeString: "bool",
							value:           map[string]any{"authorId": false},
							statusCode:      400,
							errorMsg:        "field 'filters' should adhere to a 'field: { <operator>: <value> }' format",
						},
						{
							valueTypeString: "map with invalid field",
							value:           types.Filters{"invalid-field-one": {"eq": 434}},
							statusCode:      400,
							errorMsg:        "invalid field 'invalid-field-one'",
						},
						{
							valueTypeString: "map with invalid operator",
							value:           types.Filters{"authorId": {"fdsa": 434}},
							statusCode:      400,
							errorMsg:        "invalid operator 'fdsa' for field 'authorId'",
						},
					}

					for _, testCase := range testCases {
						scenarioString := fmt.Sprintf("should not accept 'filters' with %s", testCase.valueTypeString)

						It(scenarioString, func() {
							body := map[string]any{
								"filters": testCase.value,
							}

							jsonBytes := marshalMapToBody(body)

							req, _ := http.NewRequest("POST", "/listHaikus", jsonBytes)
							w := httptest.NewRecorder()
							router.ServeHTTP(w, req)
							responseBody := UnmarshalResponseBody(w.Result().Body)

							Expect(w.Result().StatusCode).Should(Equal(400))
							Expect(responseBody["errors"]).Should(ContainElement(ContainSubstring(testCase.errorMsg)))
						})
					}
				})
			})
		})
	})

	Describe("PUT /haiku", Ordered, func() {
		Describe("a logged out Author", func() {
			It("should not be able to submit a Haiku", func() {
				putHaikuBody := map[string]any{
					"text": "a black sky trembles//the oceans turn into dust//we are alone here",
					"tags": "",
				}

				jsonBytes := marshalMapToBody(putHaikuBody)

				req, _ := http.NewRequest("PUT", "/haiku", jsonBytes)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				responseBody := UnmarshalResponseBody(w.Result().Body)

				Expect(w.Result().StatusCode).Should(Equal(401))
				Expect(responseBody["error"]).Should(Equal("unauthorized"))
			})
		})

		Describe("a logged in Author", Ordered, func() {
			username := "test-user-1"
			password := "oogity654*(Sharp09)"

			usernamePasswordEncoded := convertToBase64(fmt.Sprintf("%s:%s", username, password))
			header := fmt.Sprintf("Basic %s", usernamePasswordEncoded)

			BeforeAll(func() {
				putAuthorBody := map[string]any{
					"username": username,
					"password": password,
				}

				jsonBytes := marshalMapToBody(putAuthorBody)

				req, _ := http.NewRequest("PUT", "/author", jsonBytes)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
			})

			invalidHaikus := []HaikuTestCase{
				{scenarioString: "first phrase incomplete", haiku: "an invalid haiku here", errorMsg: HaikuIncompleteErrMsg},
				{scenarioString: "second phrase incomplete", haiku: "a black sky trembles//the oceans", errorMsg: HaikuIncompleteErrMsg},
				{scenarioString: "nothing but all sets of slashes", haiku: "////", errorMsg: "phrase '' needs to be 5 syllables"},
				{scenarioString: "nothing but all sets of slashes & spaces", haiku: "// // ", errorMsg: "phrase '' needs to be 5 syllables"},
				{scenarioString: "first phrase incomplete, rest complete", haiku: "a black sky//the oceans turn into dust//we are alone here", errorMsg: "phrase 'a black sky' needs to be 5 syllables"},
				{scenarioString: "second phrase incomplete, rest complete", haiku: "a black sky trembles//the oceans //we are alone here", errorMsg: "phrase 'the oceans ' needs to be 7 syllables"},
				{scenarioString: "third phrase incomplete, rest complete", haiku: "a black sky trembles//the oceans turn into dust// we are", errorMsg: "phrase ' we are' needs to be 5 syllables"},
				{
					scenarioString: "all phrases complete, surrounding whitespace in phrase one",
					haiku:          "  a black sky trembles  //the oceans turn into dust//we are alone here",
					errorMsg:       "remove all surrounding whitespace characters from phrase '  a black sky trembles  '",
				},
				{
					scenarioString: "all phrases complete, surrounding whitespace in phrase two",
					haiku:          "a black sky trembles//  the oceans turn into dust          //we are alone here",
					errorMsg:       "remove all surrounding whitespace characters from phrase '  the oceans turn into dust          '",
				},
				{
					scenarioString: "all phrases complete, surrounding whitespace in phrase three",
					haiku:          "a black sky trembles//the oceans turn into dust// we are alone here",
					errorMsg:       "remove all surrounding whitespace characters from phrase ' we are alone here'",
				},
			}

			for _, testCase := range invalidHaikus {
				scenario := fmt.Sprintf("should not be able to submit a haiku with %s", testCase.scenarioString)

				It(scenario, func() {
					putHaikuBody := map[string]any{
						"text": testCase.haiku,
						"tags": []string{},
					}

					jsonBytes := marshalMapToBody(putHaikuBody)

					req, _ := http.NewRequest("PUT", "/haiku", jsonBytes)
					req.Header.Add("Authorization", header)
					w := httptest.NewRecorder()
					router.ServeHTTP(w, req)
					responseBody := UnmarshalResponseBody(w.Result().Body)

					Expect(w.Result().StatusCode).Should(Equal(400))
					Expect(responseBody["error"]).Should(Equal(testCase.errorMsg))
				})
			}
		})
	})

	RegisterFailHandler(Fail)
	RunSpecs(t, "Haikus Suite")

	// black sky trembles//the oceans turn into dust//we are alone here
}
