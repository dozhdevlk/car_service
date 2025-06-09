package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

var (
	db    *sql.DB
	store = sessions.NewCookieStore([]byte("super-secret-key"))
)

func main() {

	storeOptions()
	initDB()

	r := mux.NewRouter()

	setupRoutes(r)

	InitTelegramBot()
	go StartTelegramBot(db)

	fmt.Println("Server running on :8080")

	log.Fatal(http.ListenAndServe(":8080", r))
}
