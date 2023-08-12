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
    ID              int64  `json: id`
    FirstName       string `json: first_name`
    LastName        string `json: last_name`
    Mobile          string `json: mobile_number`
    Email           string `json: em_address`
    PaymentMethodID int64  `json: payment_method_id`
    Token           string `json: token`
}

type API interface {
    Login(params LoginParam) (LoginResponse, error)
}
