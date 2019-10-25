package mgogetter

import (
	"git.bingyan.net/lyt99/gqlregister"
	"github.com/globalsign/mgo"
	"reflect"
)

type MgoSessionGetter struct {
	MongoSession *mgo.Session
	DB           string
}

type MgoMongoSession struct {
	MongoSession *mgo.Session
	DB           string
}

func (m MgoMongoSession) Insert(collection string, document interface{}) error {
	c := m.MongoSession.DB(m.DB).C(collection)

	return c.Insert(document)
}

func (m MgoMongoSession) Delete(collection string, query interface{}) error {
	c := m.MongoSession.DB(m.DB).C(collection)

	return c.Remove(query)
}

func (m MgoMongoSession) FindOne(collection string, query interface{}, t reflect.Type) (interface{}, error) {
	ret := reflect.New(t).Interface()
	c := m.MongoSession.DB(m.DB).C(collection)

	err := c.Find(query).One(ret)
	return ret, err
}

func (m MgoMongoSession) FindMany(collection string, query interface{}, t reflect.Type) ([]interface{}, error) {
	var ret []interface{}


	c := m.MongoSession.DB(m.DB).C(collection)

	it := c.Find(query).Iter()

	for v := reflect.New(t).Interface(); it.Next(v); v = reflect.New(t).Interface(){
		ret = append(ret, v)
	}

	return ret, it.Err()
}

func (m MgoMongoSession) UpdateOne(collection string, query interface{}, update interface{}) error {
	c := m.MongoSession.DB(m.DB).C(collection)

	return c.Update(query, update)
}

func (m MgoMongoSession) UpdateMany(collection string, query interface{}, update interface{}) error {
	c := m.MongoSession.DB(m.DB).C(collection)

	_, err := c.UpdateAll(query, update)
	return err
}

func (m MgoMongoSession) Close() {
	m.MongoSession.Close()
}

func (m *MgoSessionGetter) GetSession() gqlregister.MongoSession {
	return MgoMongoSession{
		MongoSession: m.MongoSession.Clone(),
		DB: m.DB,
	}
}
