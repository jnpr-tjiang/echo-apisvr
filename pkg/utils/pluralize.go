package utils

import "github.com/gertd/go-pluralize"

var pluralizer *pluralize.Client = pluralize.NewClient()

// Pluralize a string
func Pluralize(s string) string {
	return pluralizer.Plural(s)
}

// Singularize a string
func Singularize(s string) string {
	return pluralizer.Singular(s)
}
