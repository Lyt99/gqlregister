package gqlregister

import (
	"errors"
	"fmt"
	"github.com/graphql-go/graphql"
	"reflect"
)

var (
	ErrorNotStruct = errors.New("schema is not a struct")
	ErrorColExists = errors.New("collection name has existed")
)

type MongoSession interface {
	Insert(collection string, document interface{}) error
	Delete(collection string, query interface{}) error
	FindOne(collection string, query interface{}, t reflect.Type) (interface{}, error)
	FindMany(collection string, query interface{}, t reflect.Type) ([]interface{}, error)
	UpdateOne(collection string, query interface{}, update interface{}) error
	UpdateMany(collection string, query interface{}, update interface{}) error
	Close()
}

type SessionGetter interface {
	GetSession() MongoSession
}

type TObjectPair struct {
	T reflect.Type
	Object *graphql.Object
	Args graphql.FieldConfigArgument
}
type GraphqlMongoRegister struct {
	collectionType map[string]TObjectPair
	sessionGetter  SessionGetter
	query          *graphql.Object
	existingLists  map[string]*graphql.List
}

func New(sessionGetter SessionGetter) *GraphqlMongoRegister {
	ret :=  &GraphqlMongoRegister{
		collectionType: make(map[string]TObjectPair),
		sessionGetter:  sessionGetter,
		query:          &graphql.Object{},
		existingLists: make(map[string]*graphql.List),
	}

	return ret
}

func (m *GraphqlMongoRegister) makeQuery() {
	fields := graphql.Fields{}

	for k, v := range m.collectionType {
		fields[k] = &graphql.Field{
			Name:              k,
			Type:              graphql.NewList(v.Object),
			Args:              v.Args,
			Resolve: func(p graphql.ResolveParams) (i interface{}, err error) {
				session := m.sessionGetter.GetSession()
				return session.FindMany(k, p.Args, v.T)
			},
			DeprecationReason: "",
			Description:       v.Object.Description(),
		}
	}

	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}

	m.query = graphql.NewObject(rootQuery)
}

func (m *GraphqlMongoRegister) makeMutation() {
	// TODO
}

func (m *GraphqlMongoRegister) GetSchema() (*graphql.Schema, error) {
	m.makeQuery()

	schemaConfig := graphql.SchemaConfig{
		Query:        m.query,
		Mutation:     nil,
		Subscription: nil,
		Types:        nil,
		Directives:   nil,
		Extensions:   nil,
	}

	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		return nil, err
	}

	return &schema, nil
}

func (m *GraphqlMongoRegister) Register(colName string, schema interface{}) (err error) {
	defer func() {
		if e := recover(); err != nil {
			if err1, ok := e.(error); ok {
				err = err1
			} else {
				panic(e)
			}
		}
	}()

	// colName是否已经存在
	if _, exists := m.collectionType[colName]; exists {
		return ErrorColExists
	}

	m.collectionType[colName] = TObjectPair{
		T:      reflect.TypeOf(schema),
		Object: graphql.NewObject(graphql.ObjectConfig{
			Name:        colName,
			Interfaces:  nil,
			Fields:      BindFields(schema, m.existingLists),
			IsTypeOf:    nil,
			Description: fmt.Sprintf("Struct in collection %s", colName),
		}),
		Args: BindArg(schema, m.existingLists),
	}


	return nil
}
