package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type Test struct {
	Message string `json:"message"`
}

func main() {
	http.HandleFunc("/", homePage)
	http.ListenAndServe(":8080", nil)
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Connected to the home page\n")
	decoder := json.NewDecoder(r.Body)
	var t Test
	err := decoder.Decode(&t)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(w, "Message: %s", t.Message)
}
