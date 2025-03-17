package types

import (
	"net/http"
	"time"
)

type Author struct {
	ID             int
	Username       string
	Password       []byte
	Email          string
	Created        time.Time
	FavoriteHaikus []int
}

type Haiku struct {
	ID       int
	Text     string
	Tags     []string
	Rating   int16
	Created  time.Time
	AuthorID int
}

type Vote struct {
	ID             int
	Upvoted        bool
	VotedTimestamp time.Time
	AuthorID       int
	HaikuID        int
}

type RegisterAuthorPOST struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email"`
}

type FilterExpression map[string]any
type Filters map[string]FilterExpression

type ListHaikusPOST struct {
	Limit   int     `json:"limit"`
	Skip    int     `json:"skip"`
	Filters Filters `json:"filters"`
}

const HTTP_OK = http.StatusOK
const HTTP_OK_NOCONTENT = http.StatusNoContent
const HTTP_BAD = http.StatusBadRequest
const HTTP_NOTFOUND = http.StatusNotFound
const HTTP_UNAUTHORIZED = http.StatusUnauthorized
const HTTP_INTERNAL = http.StatusInternalServerError

const EQ = "eq"
const NEQ = "neq"
const IN = "in"
const NIN = "nin"
const CO = "co"
const NCO = "nco"
const LT = "lt"
const GT = "gt"
