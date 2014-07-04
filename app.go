// Package main is PooseBoard API
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "bitbucket.org/liamstask/goose/lib/goose"
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/lib/pq"
)

const (
	PARAM_API_KEY     = "apikey"
	PARAM_SCORE_COUNT = "count"
)

const STD_SCORE_COUNT = 10

const (
	DB_INSERT_SCORE_STMT = "INSERT INTO scores (name, score) values ($1, $2)"
	DB_GET_APIKEY_STMT   = "SELECT apikey FROM api_keys WHERE apikey = ($1)"
	DB_GET_SCORES_STMT   = "SELECT name, score FROM scores ORDER BY score DESC LIMIT ($1)"
)

type Score struct {
	Name  string `json:"name"`
	Score uint64 `json:"score"`
}

func main() {
	m := martini.Classic()
	m.Use(render.Renderer(render.Options{
		Charset: "ISO-8859-1",
	}))

	m.Map(SetupDB())

	m.Get("/scores/:count", Authorize, GetScores)
	m.Post("/scores", Authorize, PostScores)
	m.NotFound(NotFound)

	m.Run()
}

// Authorize checks to see if PARAM_API_KEY is set and found in db>api_keys
func Authorize(req *http.Request, r render.Render, db *sql.DB) {
	apikey := req.URL.Query().Get(PARAM_API_KEY)

	if apikey == "" {
		SendErrorAsJSON(403, "Not Authorized - Missing API Key", r)
	} else {
		if CheckAPIKeyValid(apikey, db) {
			return
		} else {
			SendErrorAsJSON(403, "Not Authorized - Invalid API Key", r)
		}
	}
}

// GetScores handles GET requests to /scores/:count and responds with a JSON array
// of Score objects depending on count sorted by high score
func GetScores(req *http.Request, r render.Render, params martini.Params, db *sql.DB) {
	countStr := params[PARAM_SCORE_COUNT]
	count, err := strconv.ParseInt(countStr, 10, 0)
	if err != nil {
		count = STD_SCORE_COUNT
	}

	statement, err := db.Prepare(DB_GET_SCORES_STMT)
	defer statement.Close()
	if err != nil {
		fmt.Println(err)
		SendErrorAsJSON(500, "Failed to get scores", r)
		return
	}

	var scoresSlice []Score = make([]Score, 0)
	rows, err := statement.Query(count)
	if err != nil {
		fmt.Println(err)
		SendErrorAsJSON(500, "Failed to get scores", r)
		return
	}
	for rows.Next() {
		score := Score{}
		err := rows.Scan(&score.Name, &score.Score)
		if err != nil {
			fmt.Println(err)
		}
		scoresSlice = append(scoresSlice, score)
	}

	r.JSON(200, scoresSlice)
}

// PostScores handles POST requests to /scores taking a single Score obj and inserting it into the db
func PostScores(req *http.Request, r render.Render, db *sql.DB) {
	decoder := json.NewDecoder(req.Body)
	var score Score
	var err = decoder.Decode(&score)
	if err != nil {
		fmt.Println(err)
		SendErrorAsJSON(500, "Failed to upload score", r)
		return
	}

	if score.Name == "" {
		SendErrorAsJSON(500, "Missing field name", r)
		return
	}

	if score.Score == 0 {
		SendErrorAsJSON(500, "Missing field score", r)
		return
	}

	statement, err := db.Prepare(DB_INSERT_SCORE_STMT)
	defer statement.Close()
	if err != nil {
		fmt.Println(err)
		SendErrorAsJSON(500, "Failed to upload score", r)
		return
	}

	res, err := statement.Exec(score.Name, score.Score)
	if err != nil || res == nil {
		fmt.Println(err)
		SendErrorAsJSON(500, "Failed to upload score", r)
		return
	}

	r.JSON(201, score)
}

// NotFound (404)
func NotFound() string {
	return "Game over poose!"
}

// CheckAPIKeyValid takes an api key and verifies it's existence in the db
func CheckAPIKeyValid(key string, db *sql.DB) bool {
	statement, err := db.Prepare(DB_GET_APIKEY_STMT)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer statement.Close()

	row := statement.QueryRow(key)
	var foundKey string

	err = row.Scan(&foundKey)
	if err != nil {
		fmt.Println(err)
		return false
	}

	return key == foundKey
}

// SendErrorAsJSON returns an error JSON obj{"error" : "msg"} with a status code
func SendErrorAsJSON(status int, msg string, r render.Render) {
	r.JSON(status, map[string]interface{}{"error": msg})
}

// SetupDB initializes a postgres db from postgres url
func SetupDB() *sql.DB {
	url := os.Getenv("DATABASE_URL")
	connection, _ := pq.ParseURL(url)

	db, err := sql.Open("postgres", connection)
	if err != nil {
		log.Println(err)
	}

	return db
}
