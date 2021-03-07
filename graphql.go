package main

import (
	"github.com/graphql-go/graphql"
)

// schema initializes the query and mutation types and returns the
// the application's GraphQL schema. Other objects and fields are initialized
// by functions defined elsewhere.
func schema(app *App) (graphql.Schema, error) {
	queryFields := graphql.Fields{
		"user": user(app),
	}
	queryType := graphql.NewObject(
		graphql.ObjectConfig{Name: "Query", Fields: queryFields},
	)
	mutationFields := graphql.Fields{
		"createUser": createUser(app),
	}
	mutationType := graphql.NewObject(
		graphql.ObjectConfig{Name: "Mutation", Fields: mutationFields},
	)
	return graphql.NewSchema(
		graphql.SchemaConfig{Query: queryType, Mutation: mutationType},
	)
}
