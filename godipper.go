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
	svc := &service{
		user:         us,
		address:      as,
		extra:        es,
		item:         is,
		tripleDipper: tds,
	}

	schema, err := schema(svc)
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
