package gqlregister

import (
	"fmt"
	"github.com/graphql-go/graphql"
	"reflect"
	"strings"
)

const TAG = "bson"

// Pasted from github.com/graphql-go/graphql/util.go
// Changed some details for MonogoDB

// can't take recursive slice type
// e.g
// type Person struct{
//	Friends []Person
// }
// it will throw panic stack-overflow
func BindFields(obj interface{}, el map[string]*graphql.List) graphql.Fields {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	fields := make(map[string]*graphql.Field)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		tag := extractTag(field.Tag)
		if tag == "-" {
			continue
		}

		fieldType := field.Type

		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		var graphType graphql.Output
		if fieldType.Kind() == reflect.Struct {
			structFields := BindFields(v.Field(i).Interface(), el)

			if tag == "" {
				fields = appendFields(fields, structFields)
				continue
			} else {
				graphType = graphql.NewObject(graphql.ObjectConfig{
					Name:   tag,
					Fields: structFields,
				})
			}
		}

		if tag == "" {
			continue
		}

		if graphType == nil {
			graphType = getGraphType(fieldType, el)
		}
		fields[tag] = &graphql.Field{
			Type: graphType,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return extractValue(tag, p.Source), nil
			},
		}
	}
	return fields
}

func getGraphType(tipe reflect.Type, el map[string]*graphql.List) graphql.Output {
	kind := tipe.Kind()
	switch kind {
	case reflect.String:
		return graphql.String
	case reflect.Int, reflect.Int8, reflect.Int32, reflect.Int64:
		return graphql.Int
	case reflect.Float32, reflect.Float64:
		return graphql.Float
	case reflect.Bool:
		return graphql.Boolean
	case reflect.Slice:
		return getGraphList(tipe, el)
	}
	return graphql.String
}

func getGraphList(tipe reflect.Type, el map[string]*graphql.List) *graphql.List {
	if tipe.Kind() == reflect.Slice {
		switch tipe.Elem().Kind() {
		case reflect.Int, reflect.Int8, reflect.Int32, reflect.Int64:
			return graphql.NewList(graphql.Int)
		case reflect.Bool:
			return graphql.NewList(graphql.Boolean)
		case reflect.Float32, reflect.Float64:
			return graphql.NewList(graphql.Float)
		case reflect.String:
			return graphql.NewList(graphql.String)
		}
	}
	// finally bind object
	t := reflect.New(tipe.Elem())
	name := strings.Replace(fmt.Sprint(tipe.Elem()), ".", "_", -1)
	v, ok := el[name]
	if ok {
		return v
	}

	obj := graphql.NewObject(graphql.ObjectConfig{
		Name:   name,
		Fields: BindFields(t.Elem().Interface(), el),
	})

	list := graphql.NewList(obj)

	el[name] = list
	return list
}

func appendFields(dest, origin graphql.Fields) graphql.Fields {
	for key, value := range origin {
		dest[key] = value
	}
	return dest
}

func extractValue(originTag string, obj interface{}) interface{} {
	val := reflect.Indirect(reflect.ValueOf(obj))

	for j := 0; j < val.NumField(); j++ {
		field := val.Type().Field(j)
		if field.Type.Kind() == reflect.Struct {
			res := extractValue(originTag, val.Field(j).Interface())
			if res != nil {
				return res
			}
		}

		if originTag == extractTag(field.Tag) {
			return reflect.Indirect(val.Field(j)).Interface()
		}
	}
	return nil
}

func extractTag(tag reflect.StructTag) string {
	t := tag.Get(TAG)
	if t != "" {
		t = strings.Split(t, ",")[0]
	}
	return t
}

// lazy way of binding args
func BindArg(obj interface{}, el map[string]*graphql.List) graphql.FieldConfigArgument {
	v := reflect.Indirect(reflect.ValueOf(obj))
	var config = make(graphql.FieldConfigArgument)
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)

		mytag := extractTag(field.Tag)
		if mytag == "-" {
			continue
		}

		config[mytag] = &graphql.ArgumentConfig{
			Type: getGraphType(field.Type, el),
		}
	}
	return config
}

//func inArray(slice interface{}, item interface{}) bool {
//	s := reflect.ValueOf(slice)
//	if s.Kind() != reflect.Slice {
//		panic("inArray() given a non-slice type")
//	}
//
//	for i := 0; i < s.Len(); i++ {
//		if reflect.DeepEqual(item, s.Index(i).Interface()) {
//			return true
//		}
//	}
//	return false
//}
