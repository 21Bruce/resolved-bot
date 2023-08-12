package api

import (
    "errors"
)

var (
    ErrLoginWrong = errors.New("login parameters error")
    ErrNoTable = errors.New("no table matches reservation")
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


type Time struct {
    Hour            string
    Minute          string
 }

type ReserveParam struct {
    VenueID          int64
    Day              string
    Month            string
    Year             string
    ReservationTimes []Time
    PartySize        int
    Token            string
    PaymentMethodID  int64
}


type ReserveResponse struct {
    ReservationTime Time
}

type API interface {
    Login(params LoginParam) (*LoginResponse, error)
    Search(params SearchParam) (*SearchResponse, error)
    Reserve(params ReserveParam) (*ReserveResponse, error)
}
