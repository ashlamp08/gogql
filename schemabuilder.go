package gogql

import (
	"github.com/graphql-go/graphql"
	"log"
	"reflect"
)

type SchemaBuilder struct {
	Query          *graphql.Object
	Mutation       *graphql.Object
	typeCollection map[string]*graphql.Object
}

func NewSchemaBuilder() *SchemaBuilder {
	return &SchemaBuilder{
		Query:          graphql.NewObject(graphql.ObjectConfig{Name: "Query", Fields: graphql.Fields{}}),
		Mutation:       graphql.NewObject(graphql.ObjectConfig{Name: "Mutation", Fields: graphql.Fields{}}),
		typeCollection: map[string]*graphql.Object{},
	}
}

func (schemaBuilder *SchemaBuilder) AddQueryAction(name string, description string, object interface{}, resolver func(graphql.ResolveParams) (interface{}, error)) *SchemaBuilder {
	gqlField := schemaBuilder.getGqlField(description, object, resolver)
	schemaBuilder.Query.AddFieldConfig(name, gqlField)
	return schemaBuilder
}

func (schemaBuilder *SchemaBuilder) AddMutationAction(name string, description string, object interface{}, resolver func(graphql.ResolveParams) (interface{}, error)) *SchemaBuilder {
	gqlField := schemaBuilder.getGqlField(description, object, resolver)
	schemaBuilder.Mutation.AddFieldConfig(name, gqlField)
	return schemaBuilder
}

func (schemaBuilder *SchemaBuilder) Build() graphql.Schema {
	schemaConfig := graphql.SchemaConfig{
		Query:    schemaBuilder.Query,
		Mutation: schemaBuilder.Mutation,
	}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		log.Fatalf("failed to create new schema, error : %v", err)
	}
	return schema
}

func getArgsFromType(objectType *graphql.Object) graphql.FieldConfigArgument {
	args := graphql.FieldConfigArgument{}
	for field, gqlField := range objectType.Fields() {
		args[field] = &graphql.ArgumentConfig{
			Type: gqlField.Type,
		}
	}
	return args
}

func (schemaBuilder *SchemaBuilder) getGqlField(description string, object interface{}, resolver func(graphql.ResolveParams) (interface{}, error)) *graphql.Field {
	objectType := reflect.TypeOf(object)
	if objectType.Kind() == reflect.Struct {
		gqlType, _ := schemaBuilder.getGqlObject(objectType)
		arguments := getArgsFromType(gqlType)
		return &graphql.Field{
			Type:        gqlType,
			Description: description,
			Args:        arguments,
			Resolve:     resolver,
		}
	} else if objectType.Kind() == reflect.Slice && objectType.Elem().Kind() == reflect.Struct {
		gqlType, _ := schemaBuilder.getGqlObject(objectType.Elem())
		arguments := getArgsFromType(gqlType)
		return &graphql.Field{
			Type:        graphql.NewList(gqlType),
			Description: description,
			Args:        arguments,
			Resolve:     resolver,
		}
	}
	panic("Object of unknown type is used. Only Struct and Slice of Struct are valid")
	return nil
}
