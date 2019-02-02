package controllers

import (
	"context"
	"encoding/json"
	"github.com/hwangseonu/goBackend/common/functions"
	"github.com/hwangseonu/goBackend/common/jwt"
	"github.com/hwangseonu/goBackend/common/models"
	"github.com/hwangseonu/goBackend/users/requests"
	"github.com/hwangseonu/goBackend/users/responses"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"regexp"
)

type UserController struct {
	http.Handler
}

func (c *UserController) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	path := []byte(req.URL.Path)

	if regexp.MustCompile("^/users$").Match(path) && req.Method == "POST" {
		c.signUp(res, req)
	} else if regexp.MustCompile("^/users$").Match(path) && req.Method == "GET" {
		c.getUserData(res, req)
	}
}

func (c UserController) signUp(res http.ResponseWriter, req *http.Request) {
	var request requests.SignUpRequest
	err := functions.Request(res, req, &request)

	if err != nil {
		return
	}

	err = models.User{
		Id: bson.NewObjectId(),
		Username: request.Username,
		Password: request.Password,
		Nickname: request.Nickname,
		Email:    request.Email,
	}.Save()

	if err != nil {
		if err.Error() == "user already exists" {
			*req = *req.WithContext(context.WithValue(req.Context(), "statusCode", 409))
			res.WriteHeader(409)
			res.Write([]byte(`{}`))
			return
		} else {
			*req = *req.WithContext(context.WithValue(req.Context(), "statusCode", 500))
			res.WriteHeader(500)
			res.Write([]byte(`{"message": `+err.Error()+`}`))
			return
		}
	}

	*req = *req.WithContext(context.WithValue(req.Context(), "statusCode", 201))
	res.WriteHeader(201)
	res.Write([]byte(`{}`))
	return
}

func (c UserController) getUserData(res http.ResponseWriter, req *http.Request) {
	claims := jwt.AuthRequire(res, req, "access")
	if claims == nil {
		return
	}

	user := new(models.User)
	user.FindByUsername(claims.Identity)

	response := responses.GetUserResponse{Username: user.Username, Nickname: user.Nickname, Email: user.Email}
	b, _ := json.MarshalIndent(response, "", "  ")

	*req = *req.WithContext(context.WithValue(req.Context(), "statusCode", 200))
	res.WriteHeader(200)
	res.Write(b)
	return
}