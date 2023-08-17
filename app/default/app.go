package default 

import (
    "github.com/21Bruce/resolved-server/api"
    "github.com/21Bruce/resolved-server/app"
    "time"
    "errors"
    "strconv"
)


type App struct {
}

func findLastTime(times []api.Time) (*api.Time, error) {
    if len(times) == 0 {
        return nil, errors.New("times list empty")
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

func (a *App) ReserveAtInterval(params app.ReserveAtIntervalParam) (*app.ReserveAtIntervalResponse, error){
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
        loginResp, err := params.API.Login(
            api.LoginParam{
                Email: params.Email,
                Password: params.Password,
            })
        
        if err != nil {
            return nil, err
        }
        reserveResp, err := params.API.Reserve(
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
                time.Sleep(repeatInterval)
                continue
            }
            return nil, app.ErrorPastDate
        }
        return &app.ReserveAtIntervalResponse{ReservationTime: reserveResp.ReservationTime}, nil

    }
}

func (a *App) ReserveAtTime(params app.ReserveAtTimeParam) (*app.ReserveAtTimeResponse, error){

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

    loginResp, err := params.API.Login(
        api.LoginParam{
            Email: params.Email,
            Password: params.Password,
        })
    
    if err != nil {
        return nil, err
    }
    reserveResp, err := params.API.Reserve(
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
    
    returnValue := app.ReserveAtTimeResponse{ ReservationTime: reserveResp.ReservationTime }
    return &returnValue, nil
    
}

