package app

import (
    "github.com/21Bruce/resolved-server/api"
    "errors"
    "time"
    "strconv"
)

var (
    ErrNoOp = errors.New("no operations scheduled")
    ErrCancel = errors.New("operation cancelled")
    ErrNoLogin = errors.New("no login default or login credentials provided")
    ErrNoLogout = errors.New("no login credentials are stored")
    ErrFinOp = errors.New("operation is not in progress")
    ErrIdOp = errors.New("no operation has specified id")
    ErrTimeFut = errors.New("provided time has passed")
)

type OperationStatus int

const (
    InProgressStatusType = iota
    SuccessStatusType 
    FailStatusType 
    CancelStatusType 
)

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

type Timeable interface {
    Time() (api.Time)
}

type ReserveAtIntervalResponse struct {
    ReservationTime api.Time
}

func (r ReserveAtIntervalResponse) Time() (api.Time) {
    return r.ReservationTime
}

type ReserveAtTimeResponse struct {
    ReservationTime api.Time
}

func (r ReserveAtTimeResponse) Time() (api.Time) {
    return r.ReservationTime
}

type OperationResult struct {
    Response    Timeable    
    Err         error
}

type Operation struct{
    ID      int64
    Cancel  chan<- bool
    Output  <-chan OperationResult
    Result  *OperationResult
    Status  OperationStatus
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

func isTimeUTCFuture(year, month, day, hour, minute int) bool{
    now := time.Now().UTC()
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

func isTimeLocalFuture(year, month, day, hour, minute int) bool{
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


func (a *AppCtx) updateOperationResult (id int64) (error) {
    for i, operation := range a.operations {
        if operation.ID == id {
            if operation.Status == InProgressStatusType {
                select {
                case opRes := <-a.operations[i].Output:
                    a.operations[i].Result = &opRes
                    if opRes.Err != nil {
                        a.operations[i].Status = FailStatusType
                    } else {
                        a.operations[i].Status = SuccessStatusType
                    }
                    return nil
                default:
                    return nil
                }
            } 
            return nil
        }
    }
    return ErrIdOp
}


func (a *AppCtx) CancelOperation(id int64) (error) {
    // update before handling
    err := a.updateOperationResult(id)
    if err != nil {
        return err
    }
    for i,operation := range a.operations {
        if (operation.ID == id){
            if operation.Status != InProgressStatusType {
                return ErrFinOp 
            }
            a.operations[i].Cancel <- true 
            a.operations[i].Status = CancelStatusType
            close(a.operations[i].Cancel)
            return nil
        }
    }
    return ErrIdOp
}

func (a *AppCtx) ScheduleReserveAtIntervalOperation(params ReserveAtIntervalParam) (int64, error) {
    id := a.idGen
    a.idGen += 1 
    if (params.Email == "" || params.Password == "") {
        if(a.loginInfo.Email == "" && a.loginInfo.Password == "") {
            return 0, ErrNoLogin
        }
        params.Email = a.loginInfo.Email
        params.Password = a.loginInfo.Password
    }
    cancel := make(chan bool)
    output := make(chan OperationResult)
    a.operations = append(a.operations, Operation{
        ID: id,
        Cancel: cancel,
        Output: output,
        Status: InProgressStatusType,
    })
    go a.reserveAtInterval(params, cancel, output)
    return id, nil
}


func (a *AppCtx) reserveAtInterval(params ReserveAtIntervalParam, cancel <-chan bool, output chan<- OperationResult){
    lastTime, err := findLastTime(params.ReservationTimes)
    if err != nil {
        output<-OperationResult{Response: nil, Err: err}     
        close(output)
        return
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
        output<-OperationResult{Response: nil, Err: err}     
        close(output)
        return
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
            output<-OperationResult{Response: nil, Err: err}     
            close(output)
            return
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
            output<-OperationResult{Response: nil, Err: err}     
            close(output)
            return
        }
        if err == api.ErrNoTable {
           cmp := isTimeLocalFuture(year, month, day, hour, minute)
           if cmp {
                select {
                case <-time.After(repeatInterval):
                    continue
                case <-cancel:
                    output<-OperationResult{Response: nil, Err: ErrCancel}     
                    close(output)
                    return
                }
            }
            output<-OperationResult{Response: nil, Err: api.ErrPastDate}     
            close(output)
            return
        }
        output<-OperationResult{
            Response: &ReserveAtIntervalResponse{ReservationTime: reserveResp.ReservationTime}, 
            Err: nil,
        }
        close(output)
        return
    }
}

func (a *AppCtx) ScheduleReserveAtTimeOperation(params ReserveAtTimeParam) (int64, error) {
    id := a.idGen
    a.idGen += 1 
    if (params.Email == "" || params.Password == "") {
        if(a.loginInfo.Email == "" && a.loginInfo.Password == "") {
            return 0, ErrNoLogin
        }
        params.Email = a.loginInfo.Email
        params.Password = a.loginInfo.Password
    }
    cancel := make(chan bool)
    output := make(chan OperationResult)
    a.operations = append(a.operations, Operation{
        ID: id,
        Cancel: cancel,
        Output: output,
        Status: InProgressStatusType,
    })
    go a.reserveAtTime(params, cancel, output)
    return id, nil
}

func (a *AppCtx) reserveAtTime(params ReserveAtTimeParam, cancel <-chan bool, output chan<- OperationResult) {
    dateInts, err := dateStringsToInts([]string{ 
        params.RequestTime.Hour,
        params.RequestTime.Minute,
        params.RequestYear,
        params.RequestMonth,
        params.RequestDay,
    })

    if err != nil {
        output <- OperationResult{Response: nil, Err:err}
        close(output)
        return
    }
    hour := dateInts[0]
    minute := dateInts[1] 
    year := dateInts[2]
    month := dateInts[3] 
    day := dateInts[4]
    if !isTimeUTCFuture(year, month, day, hour, minute) {
        output <- OperationResult{Response: nil, Err: ErrTimeFut}     
        close(output)
        return
    }
    requestTime :=  time.Date(year, time.Month(month), day, hour, minute, 0, 0, time.UTC)
    select {
    case <-time.After(time.Until(requestTime)):
    case <-cancel:
        output<- OperationResult{Response: nil, Err:ErrCancel}
        close(output)
        return
    }
    loginResp, err := a.API.Login(
        api.LoginParam{
            Email: params.Email,
            Password: params.Password,
        })
    
    if err != nil {
        output<- OperationResult{Response: nil, Err:err}
        close(output)
        return
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
        output<- OperationResult{Response: nil, Err:err}
        close(output)
        return
    }
    
    returnValue := ReserveAtTimeResponse{ ReservationTime: reserveResp.ReservationTime }
    output<- OperationResult{Response: returnValue, Err:nil}
    close(output)
    return
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

func (a *AppCtx) Logout() (error) {
    if (a.loginInfo.Email == "") && (a.loginInfo.Password == "") {
        return ErrNoLogout
    }
    params := LoginParam{
        Email:      "",
        Password:   "",
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
        // update before handling
        err := a.updateOperationResult(operation.ID)
        if err != nil {
            return "", err
        }
        opLstStr += "\tID: " + strconv.FormatInt(operation.ID, 10) + "\n" 
        opLstStr += "\tStatus: " 
        switch operation.Status {
            case InProgressStatusType:
                opLstStr += "In Progress"
            case SuccessStatusType:
                time := operation.Result.Response.Time()
                opLstStr += "Succeeded\n"
                opLstStr += "\tResult: " + time.Hour + ":" + time.Minute 
            case FailStatusType:
                err := operation.Result.Err.Error()
                opLstStr += "Failed\n"
                opLstStr += "\tResult: " + err 
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

func (a *AppCtx) OperationStatus(id int64) (OperationStatus, error) {
    for i, operation := range a.operations {
        if operation.ID == id {
            err := a.updateOperationResult(operation.ID)
            if err != nil {
                return InProgressStatusType, err
            }
            return a.operations[i].Status, nil
        }
    }
    return InProgressStatusType, ErrIdOp
}

