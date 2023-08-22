package app

import (
    "github.com/21Bruce/resolved-server/api"
    "errors"
    "time"
    "strconv"
)

type OperationStatus int

const (
    InProgressStatusType = iota
    SuccessStatusType 
    FailStatusType 
    CancelStatusType 
)

var (
    ErrNoOp = errors.New("no operations scheduled")
    ErrCancel = errors.New("operation cancelled")
)

type Operation struct{
    ID      int64
    Cancel  chan<- bool
    Status  OperationStatus
}

type LoginParam api.LoginParam

type AppCtx struct {
    API         api.API
    operations  []Operation    
    loginInfo   LoginParam
    idGen       int64
}


type ReserveAtIntervalParam struct {
    Email            string
    Password         string
    VenueID          int64
    Day              string 
    Month            string 
    Year             string 
    ReservationTimes []api.Time
    PartySize        int
    RepeatInterval   api.Time
}

type ReserveAtTimeParam struct {
    Email            string
    Password         string
    VenueID          int64
    Day              string 
    Month            string 
    Year             string 
    ReservationTimes []api.Time
    PartySize        int
    RequestDay       string 
    RequestMonth     string 
    RequestYear      string 
    RequestTime      api.Time
}

type ReserveAtIntervalResponse struct {
    ReservationTime api.Time
}

type ReserveAtTimeResponse struct {
    ReservationTime api.Time
}

func findLastTime(times []api.Time) (*api.Time, error) {
    if len(times) == 0 {
        return nil, api.ErrTimeNull
    }
    lastTime := times[0]
    for i := 1; i < len(times); i++{
        lstHr, err := strconv.ParseInt(lastTime.Hour, 10, 64)
        if err != nil {
            return nil, err
        }
        lstMn, err := strconv.ParseInt(lastTime.Minute, 10, 64)
        if err != nil {
            return nil, err
        }
        thsHr, err := strconv.ParseInt(times[i].Hour, 10, 64)
        if err != nil {
            return nil, err
        }
        thsMn, err := strconv.ParseInt(times[i].Minute, 10, 64)
        if err != nil {
            return nil, err
        }
        if (lstHr < thsHr) || (lstHr == thsHr && lstMn < thsMn) {
           lastTime = times[i] 
        } 
    }
    return &lastTime, nil
}

func dateStringsToInts(in []string) ([]int, error) {
    out := make([]int, len(in), len(in))
    for i, s := range in {
        raw, err := strconv.ParseInt(s, 10, 64)
        if err != nil {
            return nil, err 
        }
        out[i] = int(raw)
    }
    return out, nil
}

func isLastTimeFuture(year, month, day, hour, minute int) bool{
    now := time.Now()
    nowYear, nowMonth, nowDay := now.Date()
    yrCmp := nowYear < year
    yrEq := nowYear == year
    mtCmp := int(nowMonth) < month
    mtEq := int(nowMonth) == month
    dyCmp := nowDay < day
    dyEq := nowDay  == day
    hrCmp := now.Hour() < hour
    hrEq := now.Hour() == hour
    mnCmp := now.Minute() < minute
    cmp := yrCmp || 
    (yrEq && mtCmp) || 
    (yrEq && mtEq && dyCmp) || 
    (yrEq && mtEq && dyEq && hrCmp) ||
    (yrEq && mtEq && dyEq && hrEq && mnCmp) 
    return cmp     
}

func (a *AppCtx) ScheduleReserveAtInterval(params ReserveAtIntervalParam) (string, error) {
    
}

func (a *AppCtx) reserveAtInterval(params ReserveAtIntervalParam, cancel <-chan bool) (*ReserveAtIntervalResponse, error){
    lastTime, err := findLastTime(params.ReservationTimes)
    if err != nil {
        return nil, err
    }
    dateInts, err := dateStringsToInts([]string{ 
        params.RepeatInterval.Hour,
        params.RepeatInterval.Minute,
        params.Year,
        params.Month,
        params.Day,
        lastTime.Hour,
        lastTime.Minute,
    })
    if err != nil {
        return nil, err
    }
    numHrs := dateInts[0]
    numMns := dateInts[1] 
    year := dateInts[2]
    month := dateInts[3] 
    day := dateInts[4]
    hour := dateInts[5] 
    minute := dateInts[6] 

    repeatInterval := time.Hour * time.Duration(numHrs) + time.Minute * time.Duration(numMns)
    for {
        loginResp, err := a.API.Login(
            api.LoginParam{
                Email: params.Email,
                Password: params.Password,
            })
        
        if err != nil {
            return nil, err
        }
        reserveResp, err := a.API.Reserve(
            api.ReserveParam{
                AuthToken: loginResp.AuthToken,
                Day: params.Day,
                Month: params.Month,
                Year: params.Year,
                ReservationTimes: params.ReservationTimes,
                PartySize: params.PartySize,
                PaymentMethodID: loginResp.PaymentMethodID,
                VenueID: params.VenueID,
            })
        if err != nil && err != api.ErrNoTable {
            return nil, err
        }
        if err == api.ErrNoTable {
           cmp := isLastTimeFuture(year, month, day, hour, minute)
           if cmp {
                select {
                case <-time.After(repeatInterval):
                    continue
                case <-cancel:
                    return nil, ErrCancel
                }
            }
            return nil, api.ErrPastDate
        }
        return &ReserveAtIntervalResponse{ReservationTime: reserveResp.ReservationTime}, nil

    }
}

func (a *AppCtx) reserveAtTime(params ReserveAtTimeParam, cancel <-chan bool) (*ReserveAtTimeResponse, error){

    dateInts, err := dateStringsToInts([]string{ 
        params.RequestTime.Hour,
        params.RequestTime.Minute,
        params.RequestYear,
        params.RequestMonth,
        params.RequestDay,
    })

    if err != nil {
        return nil, err
    }
    hour := dateInts[0]
    minute := dateInts[1] 
    year := dateInts[2]
    month := dateInts[3] 
    day := dateInts[4]
    requestTime :=  time.Date(year, time.Month(month), day, hour, minute, 0, 0, time.UTC)
    time.Sleep(time.Until(requestTime))
    select {
    case <-time.After(time.Until(requestTime)):
    case <-cancel:
        return nil, ErrCancel
    }
    loginResp, err := a.API.Login(
        api.LoginParam{
            Email: params.Email,
            Password: params.Password,
        })
    
    if err != nil {
        return nil, err
    }
    reserveResp, err := a.API.Reserve(
        api.ReserveParam{
            AuthToken: loginResp.AuthToken,
            Day: params.Day,
            Month: params.Month,
            Year: params.Year,
            ReservationTimes: params.ReservationTimes,
            PartySize: params.PartySize,
            PaymentMethodID: loginResp.PaymentMethodID,
            VenueID: params.VenueID,
        })
    if err != nil {
        return nil, err
    }
    
    returnValue := ReserveAtTimeResponse{ ReservationTime: reserveResp.ReservationTime }
    return &returnValue, nil
    
}


func (a *AppCtx) Login(params LoginParam) (error) {
    reqParams := api.LoginParam{
        Email:      params.Email,
        Password:   params.Password,
    }
    _, err :=  a.API.Login(reqParams)
    if err != nil {
        return err
    }

    a.loginInfo = params
    return nil
}

func (a *AppCtx) OperationsToString() (string, error) {
    if len(a.operations) == 0 {
        return "", ErrNoOp
    }
    opLstStr := "Operations: \n\n"
    for i, operation := range a.operations {
        opLstStr += "\tID: " + strconv.FormatInt(operation.ID, 10) + "\n" 
        opLstStr += "\tStatus: " 
        switch operation.Status {
            case InProgressStatusType:
                opLstStr += "In Progress"
            case SuccessStatusType:
                opLstStr += "Succeeded"
            case FailStatusType:
                opLstStr += "Failed"
            case CancelStatusType:
                opLstStr += "Cancelled"
        }
        opLstStr += "\n"
        if i != (len(opLstStr) - 1) {
            opLstStr += "\n"
        }
    }
    return opLstStr, nil
}


