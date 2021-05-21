package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/alexedwards/scs/v2"
	_ "github.com/go-sql-driver/mysql"
	"github.com/graphql-go/handler"
	"github.com/rs/cors"
)

func main() {
	db, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		log.Fatalf("opening databse: %v", err)
	}
	defer db.Close()

	sm := scs.New()

	us := userService{db: db, sm: sm}
	as := addressService{db: db, us: us}
	es := extraService{db: db}
	is := itemService{db: db, es: es}
	tds := tripleDipperService{db: db, is: is}
	ors := orderService{db: db, as: as, tds: tds, us: us}
	svc := &service{
		user:         us,
		address:      as,
		extra:        es,
		item:         is,
		tripleDipper: tds,
		order:        ors,
	}

	mux := http.NewServeMux()

	schema, err := schema(svc)
	if err != nil {
		log.Fatalf("starting server: %v", err)
	}
	mux.Handle("/graphql", handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	}))

	assetServer := http.FileServer(http.Dir("./assets"))
	// The /assets prefix must be stripped. Otherwise, all of the paths that
	// the file server would search for in the local assets directory would
	// start with the prefix.
	// For example, /assets/file => ./assets/assets/file
	mux.Handle("/assets/", http.StripPrefix("/assets", assetServer))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{os.Getenv("CLIENT_ORIGIN")},
		AllowCredentials: true,
	})

	log.Fatal(http.ListenAndServe(":3000", sm.LoadAndSave(c.Handler(mux))))
}
