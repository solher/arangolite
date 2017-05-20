package requests

import "encoding/json"

// JWTAuth authenticates the given user.
type JWTAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (r *JWTAuth) Path() string {
	return "/_open/auth"
}

func (r *JWTAuth) Method() string {
	return "POST"
}

func (r *JWTAuth) Generate() []byte {
	m, _ := json.Marshal(r)
	return m
}
