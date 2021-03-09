package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/alexedwards/scs/v2"
	_ "github.com/go-sql-driver/mysql"
	"github.com/graphql-go/handler"
)

// app defines interface types for services used by GraphQL resolvers
// throughout the application.
type app struct {
	users interface {
		findByPhone(phone string) (*User, error)
		signUp(u *User, code string, ctx context.Context) error
		logIn(phone, code string, ctx context.Context) (*User, error)
		logOut(ctx context.Context) error
		phoneFromSession(ctx context.Context) string
	}
}

func main() {
	db, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		log.Fatalf("opening databse: %v", err)
	}
	defer db.Close()

	sm := scs.New()

	a := &app{
		users: userService{db: db, sm: sm},
	}

	schema, err := schema(a)
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
