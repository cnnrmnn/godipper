package main

import (
	"github.com/graphql-go/graphql"
)

// schema initializes the query and mutation fields and returns the
// application's GraphQL schema. Other objects and fields are initialized
// elsewhere.
func schema(svc *service) (graphql.Schema, error) {
	queryFields := graphql.Fields{
		"me":        me(svc),
		"addresses": addresses(svc),
	}
	queryType := graphql.NewObject(
		graphql.ObjectConfig{Name: "Query", Fields: queryFields},
	)
	mutationFields := graphql.Fields{
		"sendCode":      sendCode(svc),
		"signUp":        signUp(svc),
		"logIn":         logIn(svc),
		"logOut":        logOut(svc),
		"createAddress": createAddress(svc),
	}
	mutationType := graphql.NewObject(
		graphql.ObjectConfig{Name: "Mutation", Fields: mutationFields},
	)
	return graphql.NewSchema(
		graphql.SchemaConfig{Query: queryType, Mutation: mutationType},
	)
}
