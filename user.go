package main

import (
	"github.com/cnnrmnn/godipper/chilis"
	"github.com/graphql-go/graphql"
)

type User struct {
	firstName string `json:"firstName"`
	lastName  string `json:"lastName"`
	phone     string `json:"phone"`
	email     string `json:"email"`
}

var userType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "User",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.Int,
			},
			"firstName": &graphql.Field{
				Type: graphql.String,
			},
			"lastName": &graphql.Field{
				Type: graphql.String,
			},
			"phone": &graphql.Field{
				Type: graphql.String,
			},
			"email": &graphql.Field{
				Type: graphql.String,
			},
		},
	},
)

func (u User) Addresss() chilis.Address {
	return chilis.Address{}
}

func (u User) FirstName() string {
	return u.firstName
}

func (u User) LastName() string {
	return u.lastName
}

func (u User) Phone() string {
	return u.phone
}

func (u User) Email() string {
	return u.email
}
