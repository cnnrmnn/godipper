package main

import (
	"log"
	"net/http"

	"github.com/graphql-go/handler"
)

func main() {
	schema, err := schema()
	if err != nil {
		log.Fatalf("starting server: %v", err)
	}

	h := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})

	http.Handle("/graphql", h)
	log.Fatal(http.ListenAndServe(":3000", nil))
}
