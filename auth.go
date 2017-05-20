package arangolite

import (
	"context"
	"net/http"

	"github.com/solher/arangolite/requests"
)

type authentication interface {
	Setup(ctx context.Context, db *Database) error
	Apply(req *http.Request) error
}

type basicAuth struct {
	username, password string
}

func (a *basicAuth) Setup(ctx context.Context, db *Database) error {
	return nil
}

func (a *basicAuth) Apply(req *http.Request) error {
	req.SetBasicAuth(a.username, a.password)
	return nil
}

type jwtAuth struct {
	username, password string
	jwt                string
}

func (a *jwtAuth) Setup(ctx context.Context, db *Database) error {
	res, err := db.Send(ctx, &requests.JWTAuth{Username: a.username, Password: a.password})
	if err != nil {
		return err
	}
	jwtRes := struct {
		JWT string `json:"jwt"`
	}{}
	if err := res.Unmarshal(&jwtRes); err != nil {
		return err
	}
	a.jwt = jwtRes.JWT
	return nil
}

func (a *jwtAuth) Apply(req *http.Request) error {
	req.Header.Set("Authorization", "bearer "+a.jwt)
	return nil
}
