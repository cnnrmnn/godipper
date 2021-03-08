package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/alexedwards/scs/v2"
	_ "github.com/go-sql-driver/mysql"
	"github.com/graphql-go/handler"
)

type App struct {
	sm    *scs.SessionManager
	users interface {
		FindByID(id int) (*User, error)
		Create(u *User) error
	}
}

func main() {
	db, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		log.Fatalf("opening databse: %v", err)
	}
	defer db.Close()

	sm := scs.New()

	app := &App{
		sm:    sm,
		users: UserService{db: db},
	}

	schema, err := schema(app)
	if err != nil {
		log.Fatalf("starting server: %v", err)
	}

	h := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})

	http.Handle("/graphql", sm.LoadAndSave(h))
	log.Fatal(http.ListenAndServe(":3000", nil))
}
