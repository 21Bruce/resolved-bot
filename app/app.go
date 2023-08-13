package app

import (
    "github.com/21Bruce/resolved-server/api"
    "errors"
)

var (
    ErrorPastDate = errors.New("no more reservations possible")
)

type ReserveAtIntervalParam struct {
    Email    string
    Password string
    VenueID          int64
    Day              string 
    Month            string 
    Year             string 
    ReservationTimes []api.Time
    PartySize        int
    API api.API
    RepeatInterval api.Time
}

type ReserveAtTimeParam struct {
}

type ReserveAtIntervalResponse struct {
    ReservationTime api.Time
}

type ReserveAtTimeResponse struct {
    ReservationTime api.Time
}

type App interface {
    ReserveAtInterval(params ReserveAtIntervalParam) (*ReserveAtIntervalResponse, error)
    ReserveAtTime(params ReserveAtTimeParam) (*ReserveAtTimeResponse, error)
}
