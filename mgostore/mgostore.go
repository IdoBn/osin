//this is storage for oath modules
package mgostore

import (
	"encoding/json"
	"fmt"

	"github.com/idobn/osin"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var _ = fmt.Printf

//some collection name
const (
	CLIENT_COL    = "clients"
	AUTHORIZE_COL = "authorizations"
	ACCESS_COL    = "accesses"
)

const REFRESHTOKEN = "refreshToken"

//keep session to mgo
type OAuthStorage struct {
	dbName  string
	Session *mgo.Session
}

//initialize new storage -- should put global mgo session into
func New(session *mgo.Session, dbName string) *OAuthStorage {
	storage := &OAuthStorage{dbName, session}
	index := mgo.Index{
		Key:        []string{REFRESHTOKEN},
		Unique:     false, // refreshtoken is sometimes empty
		DropDups:   false,
		Background: true,
		Sparse:     true,
	}
	accesses := storage.Session.DB(dbName).C(ACCESS_COL)
	err := accesses.EnsureIndex(index)
	if err != nil {
		panic(err)
	}
	return storage
}

//renew new storage with cloned session
func (s *OAuthStorage) Clone() osin.Storage {
	//clone mgo session and return the storage
	clonedSession := s.Session.Clone()
	newStorage := &OAuthStorage{s.dbName, clonedSession}
	return newStorage
}

//close the session
func (s *OAuthStorage) Close() {
	if s.Session != nil {
		s.Session.Close()
	}
}

func (s *OAuthStorage) GetClient(id string) (osin.Client, error) {
	clients := s.Session.DB(s.dbName).C(CLIENT_COL)
	client := &osin.DefaultClient{}

	err := clients.Find(bson.M{"_id": id}).One(client)
	return client, err
}

func (s *OAuthStorage) GetClients(query bson.M, pageSize, pageNum int) ([]osin.DefaultClient, error) {
	// err = sess.DB("test").C("foo").Find(bson.M{}).Skip(pagesize * (n - 1)).Limit(10)
	clients := s.Session.DB(s.dbName).C(CLIENT_COL)
	clientArr := []osin.DefaultClient{}

	err := clients.Find(query).Skip(pageSize * (pageNum - 1)).Limit(pageSize).All(&clientArr)
	return clientArr, err
}

func (s *OAuthStorage) SetClient(id string, client osin.Client) error {
	clients := s.Session.DB(s.dbName).C(CLIENT_COL)
	_, err := clients.UpsertId(id, client)
	return err
}

// RemoveClient todo
func (s *OAuthStorage) RemoveClient(id string) error {
	clients := s.Session.DB(s.dbName).C(CLIENT_COL)
	err := clients.RemoveId(id)
	return err
}

func (s *OAuthStorage) SaveAuthorize(data *osin.AuthorizeData) error {
	authorizations := s.Session.DB(s.dbName).C(AUTHORIZE_COL)
	_, err := authorizations.UpsertId(data.Code, data)
	return err
}

func (s *OAuthStorage) LoadAuthorize(code string) (*osin.AuthorizeData, error) {
	authorizations := s.Session.DB(s.dbName).C(AUTHORIZE_COL)
	authData := osin.AuthorizeData{Client: &osin.DefaultClient{}}

	genericAuthorizeData := make(map[string]interface{})
	if err := authorizations.FindId(code).One(&genericAuthorizeData); err != nil {
		return &authData, err
	}

	jsonData, err := json.Marshal(&genericAuthorizeData)
	if err != nil {
		return &authData, err
	}

	//then unmarshal again
	if err := json.Unmarshal(jsonData, &authData); err != nil {
		return &authData, err
	}

	//if everything is fine; then redirect directly
	return &authData, nil
}

func (s *OAuthStorage) RemoveAuthorize(code string) error {
	authorizations := s.Session.DB(s.dbName).C(AUTHORIZE_COL)
	return authorizations.RemoveId(code)
}

func (s *OAuthStorage) SaveAccess(data *osin.AccessData) error {
	accesses := s.Session.DB(s.dbName).C(ACCESS_COL)

	//to avoid multiple nested previous record
	if data.AccessData != nil {
		data.AccessData.AccessData = nil
	}

	_, err := accesses.UpsertId(data.AccessToken, data)
	return err
}

func (s *OAuthStorage) LoadAccess(token string) (*osin.AccessData, error) {
	accesses := s.Session.DB(s.dbName).C(ACCESS_COL)

	newClient := osin.DefaultClient{}
	authorizeData := osin.AuthorizeData{
		Client: &newClient,
	}

	prevNewClient := osin.DefaultClient{}

	//TODO: check overhere to avoid infitite recursive -- because client is interface
	accData := osin.AccessData{
		Client:        &newClient,
		AuthorizeData: &authorizeData,
		AccessData: &osin.AccessData{
			Client: &prevNewClient,
			AuthorizeData: &osin.AuthorizeData{
				Client: &prevNewClient,
			},
		},
	}

	genericAccessData := make(map[string]interface{})
	if err := accesses.FindId(token).One(&genericAccessData); err != nil {
		return &accData, err
	}

	jsonData, err := json.Marshal(&genericAccessData)
	if err != nil {
		return &accData, err
	}

	//then unmarshal again
	if err := json.Unmarshal(jsonData, &accData); err != nil {
		return &accData, err
	}

	//if everything is fine; then redirect directly
	return &accData, err
}

func (s *OAuthStorage) RemoveAccess(token string) error {
	accesses := s.Session.DB(s.dbName).C(ACCESS_COL)
	return accesses.RemoveId(token)
}

//loading access data based on refresh token instead
func (s *OAuthStorage) LoadRefresh(token string) (*osin.AccessData, error) {
	accesses := s.Session.DB(s.dbName).C(ACCESS_COL)

	newClient := osin.DefaultClient{}
	authorizeData := osin.AuthorizeData{
		Client: &newClient,
	}

	prevNewClient := osin.DefaultClient{}

	//TODO: check overhere to avoid infitite recursive -- because client is interface
	accData := osin.AccessData{
		Client:        &newClient,
		AuthorizeData: &authorizeData,
		AccessData: &osin.AccessData{
			Client: &prevNewClient,
			AuthorizeData: &osin.AuthorizeData{
				Client: &prevNewClient,
			},
		},
	}

	genericAccessData := make(map[string]interface{})
	if err := accesses.Find(bson.M{REFRESHTOKEN: token}).One(&genericAccessData); err != nil {
		return &accData, err
	}

	jsonData, err := json.Marshal(&genericAccessData)
	if err != nil {
		return &accData, err
	}

	//then unmarshal again
	if err := json.Unmarshal(jsonData, &accData); err != nil {
		return &accData, err
	}

	//if everything is fine; then redirect directly
	return &accData, err
}

func (s *OAuthStorage) RemoveRefresh(token string) error {
	accesses := s.Session.DB(s.dbName).C(ACCESS_COL)
	return accesses.Update(bson.M{REFRESHTOKEN: token}, bson.M{
		"$unset": bson.M{
			REFRESHTOKEN: 1,
		}})
}
