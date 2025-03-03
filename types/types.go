package types

import (
	"net/http"
	"time"
)

type Author struct {
	ID       string
	Username string
	Password []byte
	Email    string
	Created  time.Time
}

type Haiku struct {
	ID       string
	Text     string
	Created  time.Time
	Rating   int16
	Tags     string
	AuthorID string
}

type ListHaikusPOST struct {
	Limit int `json:"limit"`
	Skip  int `json:"skip"`
}

type RegisterAuthorPOST struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required"`
}

type HaikuPUT struct {
	Text string `json:"text" binding:"required"`
	Tags string `json:"tags"`
}

const HTTP_OK = http.StatusOK
const HTTP_BAD = http.StatusBadRequest
const HTTP_NOTFOUND = http.StatusNotFound
const HTTP_UNAUTHORIZED = http.StatusUnauthorized
const HTTP_INTERNAL = http.StatusInternalServerError
