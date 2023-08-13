package api

import (
    "errors"
)

var (
    ErrLoginWrong = errors.New("invalid login credentials")
    ErrNoTable = errors.New("no tables available matching reservation requests")
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
    AuthToken       string 
}

type SearchParam struct {
    AuthToken       string    
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
    AuthToken        string
    PaymentMethodID  int64
}


type ReserveResponse struct {
    ReservationTime Time
    ResyToken       string
}

type CancelParam struct {
    ResyToken       string
    AuthToken       string
}

type CancelResponse struct {
    Refund          bool
}

type API interface {
    Login(params LoginParam) (*LoginResponse, error)
    Search(params SearchParam) (*SearchResponse, error)
    Reserve(params ReserveParam) (*ReserveResponse, error)
    Cancel(params CancelParam) (*CancelResponse, error)
}
