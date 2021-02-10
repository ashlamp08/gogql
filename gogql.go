package gogql

import (
	"errors"
	"github.com/graphql-go/graphql"
	"reflect"
)

/**
This library takes heavy reference from https://github.com/SonicRoshan/straf
under MIT License for learning purpose.
*/

func (schemaBuilder *SchemaBuilder) getGqlObject(objectType reflect.Type) (*graphql.Object, error) {
	gqlObject, ok := schemaBuilder.typeCollection[objectType.Name()]
	if ok {
		return gqlObject, nil
	}
	gqlFields := schemaBuilder.getGqlFields(objectType)

	gqlObject = graphql.NewObject(
		graphql.ObjectConfig{
			Name:   objectType.Name(),
			Fields: gqlFields,
		},
	)
	schemaBuilder.typeCollection[objectType.Name()] = gqlObject
	return gqlObject, nil
}

func (schemaBuilder *SchemaBuilder) getGqlObjectFromFieldType(fieldType reflect.Type) *graphql.Object {
	gqlObject, ok := schemaBuilder.typeCollection[fieldType.Name()]
	if ok {
		return gqlObject
	}
	gqlFields := schemaBuilder.getGqlFields(fieldType)

	gqlObject = graphql.NewObject(
		graphql.ObjectConfig{
			Name:   fieldType.Name(),
			Fields: gqlFields,
		},
	)
	schemaBuilder.typeCollection[fieldType.Name()] = gqlObject
	return gqlObject
}

func (schemaBuilder *SchemaBuilder) getGqlFields(structType reflect.Type) graphql.Fields {
	gqlFields := graphql.Fields{}

	for i := 0; i < structType.NumField(); i++ {
		currentField := structType.Field(i)
		exclude := getTagValue(currentField, "exclude")
		if exclude != "true" {
			currentFieldType := schemaBuilder.getFieldType(currentField)
			gqlFields[getTagValue(currentField, "json")] = &graphql.Field{
				Name:              getTagValue(currentField, "json"),
				Type:              currentFieldType,
				DeprecationReason: getTagValue(currentField, "deprecationReason"),
				Description:       getTagValue(currentField, "description"),
			}
		}
	}

	return gqlFields
}

func (schemaBuilder *SchemaBuilder) getFieldType(field reflect.StructField) graphql.Output {

	isId, ok := field.Tag.Lookup("unique")
	if ok && isId == "true" {
		return graphql.ID
	}

	if field.Type.Kind() == reflect.Struct {
		return schemaBuilder.getGqlObjectFromFieldType(field.Type)
	} else if field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.Struct {
		elementType := schemaBuilder.getGqlObjectFromFieldType(field.Type.Elem())
		return graphql.NewList(elementType)
	} else if field.Type.Kind() == reflect.Slice {
		elementType, err := getSimpleGqlType(field.Type.Elem())
		if err != nil {
			panic("invalid type for struct to graphql conversion")
		}
		return graphql.NewList(elementType)
	}

	gqlType, err := getSimpleGqlType(field.Type)
	if err != nil {
		panic("invalid type for struct to graphql conversion")
	}

	return gqlType
}

func getSimpleGqlType(fieldType reflect.Type) (*graphql.Scalar, error) {
	typeMap := map[reflect.Kind]*graphql.Scalar{
		reflect.String:  graphql.String,
		reflect.Bool:    graphql.Boolean,
		reflect.Int:     graphql.Int,
		reflect.Int8:    graphql.Int,
		reflect.Int16:   graphql.Int,
		reflect.Int32:   graphql.Int,
		reflect.Int64:   graphql.Int,
		reflect.Float32: graphql.Float,
		reflect.Float64: graphql.Float,
	}
	gqlType, ok := typeMap[fieldType.Kind()]
	if !ok {
		return &graphql.Scalar{}, errors.New("invalid type")
	}
	return gqlType, nil
}

func getTagValue(objectType reflect.StructField, tagName string) string {
	tag := objectType.Tag
	value, ok := tag.Lookup(tagName)
	if !ok {
		return ""
	}
	return value
}
