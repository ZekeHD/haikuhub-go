package types

import (
	"net/http"
	"time"
)

type Author struct {
	ID       int
	Username string
	Password []byte
	Email    string
	Created  time.Time
}

type Haiku struct {
	ID       int
	Text     string
	Tags     string
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
	Email    string `json:"email" binding:"required"`
}

const HTTP_OK = http.StatusOK
const HTTP_OK_NOCONTENT = http.StatusNoContent
const HTTP_BAD = http.StatusBadRequest
const HTTP_NOTFOUND = http.StatusNotFound
const HTTP_UNAUTHORIZED = http.StatusUnauthorized
const HTTP_INTERNAL = http.StatusInternalServerError
