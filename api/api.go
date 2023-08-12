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
    ID              int64  
    FirstName       string 
    LastName        string 
    Mobile          string 
    Email           string 
    PaymentMethodID int64  
    Token           string 
}

type SearchParam struct {
    Token           string    
    Name            string
    Limit           int
}

type SearchResponse struct {
    Results []SearchResult
}

type SearchResult struct {
    VenueID         int64 
    Name            string
    Region          string
    Locality        string
    Neighborhood    string
}


type API interface {
    Login(params LoginParam) (LoginResponse, error)
    Search(params SearchParam) (SearchResponse, error)
}
