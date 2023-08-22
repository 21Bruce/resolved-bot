package api

import (
    "errors"
    "strconv"
)

var (
    ErrLoginWrong = errors.New("invalid login credentials")
    ErrNoTable = errors.New("no tables available matching reservation requests")
    ErrNetwork = errors.New("unknown network error")
    ErrPastDate = errors.New("latest reservation time has passed")
    ErrTimeNull = errors.New("times list empty")
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

func (sr *SearchResponse) ToString() (string) {
    respStr := "\nResponses:"
    for _, e := range sr.Results {
        respStr += "\n"
        respStr += "\tName: " + e.Name + "\n"
        respStr += "\t\tVenueID: " + strconv.FormatInt(e.VenueID, 10) + "\n"
        respStr += "\t\tRegion: " + e.Region + "\n"
        respStr += "\t\tLocality: " + e.Locality + "\n"
        respStr += "\t\tNeighborhood: " + e.Neighborhood +"\n"
    }
    return respStr
}
