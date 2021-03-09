package main

import (
	"github.com/graphql-go/graphql"
)

// schema initializes the query and mutation fields and returns the
// application's GraphQL schema. Other objects and fields are initialized
// elsewhere.
func schema(a *app) (graphql.Schema, error) {
	queryFields := graphql.Fields{
		"me": me(a),
	}
	queryType := graphql.NewObject(
		graphql.ObjectConfig{Name: "Query", Fields: queryFields},
	)
	mutationFields := graphql.Fields{
		"sendCode": sendCode(a),
		"signUp":   signUp(a),
		"logIn":    logIn(a),
		"logOut":   logOut(a),
	}
	mutationType := graphql.NewObject(
		graphql.ObjectConfig{Name: "Mutation", Fields: mutationFields},
	)
	return graphql.NewSchema(
		graphql.SchemaConfig{Query: queryType, Mutation: mutationType},
	)
}
