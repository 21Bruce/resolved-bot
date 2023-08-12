package api

import (
    "errors"
)

var (
    ErrLoginWrong = errors.New("login parameters error")
)

type LoginParam struct {
    Email string
    Password string
}

type LoginResponse struct {
    ID int64
    First string 
    Last string 
    Mobile string
    Email string
    Payment_method_id int64
    Token string
}

type API interface {
    Login(params LoginParam) (LoginResponse, error)
}
