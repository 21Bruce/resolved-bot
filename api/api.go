/*
Author: Bruce Jagid
Created On: Aug 21, 2023
*/
package api

import (
    "errors"
    "strconv"
    "time"
)

var (
    ErrLoginWrong = errors.New("invalid login credentials")
    ErrNoTable = errors.New("no tables available matching reservation requests")
    ErrNetwork = errors.New("unknown network error")
    ErrPastDate = errors.New("latest reservation time has passed")
    ErrTimeNull = errors.New("times list empty")
    ErrNoOffer = errors.New("table is not offered on given date")
    ErrNoPayInfo = errors.New("no payment info on account")
)


/*
Name: LoginParam
Type: API Func Input Struct
Purpose: Input parameters for the api function 'Login'
Note: LoginParam is meant to hide login details from the app layer,
but each individual external service has different login requirements.

Field Requirements for Resy:
    - Email: string 
    - Password: string 

Field Requirements for Opentable:
    - FirstName: string 
    - LastName: string
    - Email: string
    - Mobile: string, omitting dashes and region indicator(i.e. the +1 for US)

*/
type LoginParam struct {
    FirstName       string 
    LastName        string 
    Mobile          string 
    Email           string
    Password        string
}

/*
Name: LoginResponse
Type: API Func Output Struct
Purpose: Output information for the api function 'Login'
Note: LoginResponse is only meant to be used as an input to the 'Reserve' api function,
and its internals are subject to change with any update, so no code should be written on
another layer relying on the fields of this data structure
*/
type LoginResponse struct {
    ID              int64  
    FirstName       string 
    LastName        string 
    Mobile          string 
    Email           string 
    PaymentMethodID int64  
    AuthToken       string 
}

/*
Name: SeachParam 
Type: API Func Input Struct
Purpose: Input information to the 'Search' api function 
*/
type SearchParam struct {
    Name            string
    Limit           int
}

/*
Name: SeachResponse
Type: API Func Output Struct
Purpose: Output information from 'Search' api function 
*/
type SearchResponse struct {
    Results []SearchResult
}

/*
Name: SeachResult
Type: API Output Struct
Purpose: Output specific results from 'Search' api function 
*/
type SearchResult struct {
    VenueID         int64 
    Name            string
    Region          string
    Locality        string
    Neighborhood    string
}

/*
Name: LongTime
Type: API Input Struct
Purpose: Provide a go indepent struct for representing time
at a long scale(i.e. years + months + days)
*/
/*type LongTime struct {
    Year            string
    Month           string
    Day             string
    Hour            string
    Minute          string
}*/

type TableType int64

const (
    Empty TableType = iota
    DiningRoom
    Indoor
    Outdoor
    Patio
    Bar
    Lounge
)

/*
Name: ReserveParam
Type: API Func Input Struct
Purpose: Input information to the 'Reserve' api function 
*/
type ReserveParam struct {
    VenueID          int64
    ReservationTimes []time.Time
    PartySize        int
    Table            TableType
    LoginResp        LoginResponse
}

/*
Name: ReserveResponse
Type: API Func Output Struct
Purpose: Output information from the 'Reserve' api function 
*/
type ReserveResponse struct {
    ReservationTime time.Time
}

/*
Name: API 
Type: Interface 
Purpose: Provide a minimal enough abstraction of common behavior
among external reservation services to allow cross-platform
application production
*/
type API interface {
    Login(params LoginParam) (*LoginResponse, error)
    Search(params SearchParam) (*SearchResponse, error)
    Reserve(params ReserveParam) (*ReserveResponse, error)
    AuthMinExpire() (time.Duration)
}

/*
Name: SearchResponse.ToString 
Type: Stringify Func
Purpose: Provide a default string representation of search
responses amongst consumers of this layer 
*/
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
