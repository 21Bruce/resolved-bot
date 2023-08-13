package golang

import (
    "github.com/21Bruce/resolved-server/api"
    "github.com/21Bruce/resolved-server/app"
    "time"
    "errors"
    "strconv"
    "fmt"
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

func (a *App) ReserveAtInterval(params app.ReserveAtIntervalParam) (*app.ReserveAtIntervalResponse, error){
    lastTime, err := findLastTime(params.ReservationTimes)
    if err != nil {
        return nil, err
    }
    numHours, err := strconv.ParseInt(params.RepeatInterval.Hour, 10, 64)
    if err != nil {
        return nil, err
    }
    numMinutes, err := strconv.ParseInt(params.RepeatInterval.Minute, 10, 64)
    if err != nil {
        return nil, err
    }
    year, err :=  strconv.ParseInt(params.Year, 10, 64)     
    if err != nil {
        return nil, err
    }
    month, err :=  strconv.ParseInt(params.Month, 10, 64)     
    if err != nil {
        return nil, err
    }
    day, err :=  strconv.ParseInt(params.Day, 10, 64)     
    if err != nil {
        return nil, err
    }
    hour, err := strconv.ParseInt(lastTime.Hour, 10, 64)
    if err != nil {
        return nil, err
    }
    minute, err := strconv.ParseInt(lastTime.Minute, 10, 64)
    if err != nil {
        return nil, err
    }

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
            now := time.Now()
            yrCmp := now.Year() <= int(year)
            yrEq := now.Year() == int(year)
            mtCmp := int(now.Month()) <= int(month)
            mtEq := int(now.Month()) == int(month)
            dyCmp := now.YearDay() <= int(day)
            dyEq := now.YearDay() == int(day)
            hrCmp := now.Hour() <= int(hour)
            hrEq := now.Hour() == int(hour)
            mnCmp := now.Minute() <= int(minute)
            cmp := yrCmp || 
            (yrEq && mtCmp) || 
            (yrEq && mtEq && dyCmp) || 
            (yrEq && mtEq && dyEq && hrCmp) ||
            (yrEq && mtEq && dyEq && hrEq && mnCmp) 
            if cmp {
                time.Sleep(time.Hour * time.Duration(int(numHours)) + time.Minute * time.Duration(int(numMinutes)))
                continue
            }
            return nil, app.ErrorPastDate
        }
        return &app.ReserveAtIntervalResponse{ReservationTime: reserveResp.ReservationTime}, nil

    }
}

