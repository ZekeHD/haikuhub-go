## HaikuHub-Go

The API & backend Golang/PostgreSQL repository for https://haikuhub.net.

The idea is to host all API endpoints in this repository. The wishlist is:

- return all haikus &#x2705;
  - include skip/limit/sort by popularity &#x270E;
- PUT new Haiku &#x2705;
  - new Haiku validation by syllable count for language used &#x270E;
  - multi-language support &#x270E;
  - validation against slurs and other mean language &#x270E;
- GET Haiku by ID &#x2705;
- use environment variables for Postgres auth &#x2705;
- user (Author) creation & authentication via Authentication header &#x270E;
- Go service & Postgres hosting via Heroku &#x270E;
- allow user voting by upvote/downvote & favorite &#x270E;