/*
Author: Bruce Jagid
Created On: Aug 12, 2023
*/
package app

import (
    "github.com/21Bruce/resolved-server/api"
    "errors"
    "time"
    "strconv"
    "fmt"
)

var (
    ErrNoOp = errors.New("no operations scheduled")
    ErrCancel = errors.New("operation cancelled")
    ErrNoLogin = errors.New("no login default or login credentials provided")
    ErrNoLogout = errors.New("no login credentials are stored")
    ErrFinOp = errors.New("operation is not in progress")
    ErrCurrOp = errors.New("operation is in progress")
    ErrIdOp = errors.New("no operation has specified id")
    ErrTimeFut = errors.New("provided time has passed")
)

// OperationStatus type is an enum, only use with next const def types
type OperationStatus int

const (
    InProgressStatusType = iota
    SuccessStatusType 
    FailStatusType 
    CancelStatusType 
)

// Hide as much api layer details as permissible 
type LoginParam api.LoginParam

type SearchParam api.SearchParam

type SearchResponse api.SearchResponse

/*
Name: AppCtx
Type: External App Struct
Purpose: The AppCtx struct is the namespace and
state of the core app, in which all App external
functions run
*/
type AppCtx struct {

    // The API to run the app on
    API         api.API

    // List of internal concurrent operations, both completed
    // and running
    operations  []Operation    

    // Login Default storage 
    loginInfo   LoginParam

    // Simple ID generator
    idGen       int64
}

/*
Name: ReserveAtIntervalParam
Type: App api func input parameters
Purpose: Provide a means to make a 
reserve at interval operation by a consumer
*/
type ReserveAtIntervalParam struct {
    Login            LoginParam
    VenueID          int64
    ReservationTimes []time.Time
    PartySize        int
    RepeatInterval   time.Duration
}

/*
Name: ReserveAtTimeParam
Type: App api func input parameters
Purpose: Provide a means to make a 
reserve at time operation by a consumer
*/
type ReserveAtTimeParam struct {
    Login            LoginParam
    VenueID          int64
    ReservationTimes []time.Time
    PartySize        int
    RequestTime      time.Time
}

/*
Name: Timeable 
Type: interface 
Purpose: Provide a common definition
for an operation result
*/
type Timetable interface {
    Time() (time.Time)
}

/*
Name: ReserveAtIntervalResponse 
Type: struct
Purpose: Define the data that should be returned on a
successful reserve at interval response
*/
type ReserveAtIntervalResponse struct {
    ReservationTime time.Time
}

/*
Name: Time 
Type: interface method
Purpose: Satisfy the Timetable interface
*/
func (r ReserveAtIntervalResponse) Time() (time.Time) {
    return r.ReservationTime
}

/*
Name: ReserveAtTimeResponse 
Type: struct
Purpose: Define the data that should be returned on a
successful reserve at time response
*/
type ReserveAtTimeResponse struct {
    ReservationTime time.Time
}

/*
Name: Time 
Type: interface method
Purpose: Satisfy the Timetable interface
*/
func (r ReserveAtTimeResponse) Time() (time.Time) {
    return r.ReservationTime
}

/*
Name: OperationResult 
Type: struct 
Purpose: Define a consistent way of conveying a successful
operation
*/
type OperationResult struct {
    Response    Timetable
    Err         error
}

/*
Name: Operation 
Type: struct
Purpose: Maintain the state associated with a running
go thread operation
*/
type Operation struct{
    ID      int64
    Cancel  chan<- bool
    Output  <-chan OperationResult
    Result  *OperationResult
    Status  OperationStatus
}


/*
Name: findLastTime
Type: Internal Func
Purpose: Out of a list of times, return the latest time
*/
func findLastTime(times []time.Time) (*time.Time, error) {
    if len(times) == 0 {
        return nil, api.ErrTimeNull
    }
    lastTime := times[0]
    for i := 1; i < len(times); i++{
        if times[i].After(lastTime) {
            lastTime = times[i]
        }
    }
    return &lastTime, nil
}

/*
Name: updateOperationResult 
Type: Internal Func
Purpose: Used before querying an operation to 
make sure state is consistent with go thread
*/
func (a *AppCtx) updateOperationResult (id int64) (error) {
    for i, operation := range a.operations {
        if operation.ID == id {
            if operation.Status == InProgressStatusType {
                // if operation is in progess, see if it finished
                // by trying to read output in a nonblocking way
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
                    // if Output is not ready, do nothing
                    return nil
                }
            } 
            return nil
        }
    }
    return ErrIdOp
}

/*
Name: CancelOperation 
Type: External App Func
Purpose: Used to cancel the operation 
with the specified ID
*/
func (a *AppCtx) CancelOperation(id int64) (error) {
    // update before handling
    err := a.updateOperationResult(id)
    if err != nil {
        return err
    }
    for i, operation := range a.operations {
        if (operation.ID == id){
            // can only cancel ops in progress
            if operation.Status != InProgressStatusType {
                return ErrFinOp 
            }

            // we perform all stateful changes in place,
            // i.e., on a.operations[i] instead of the for loop
            // value 'operation' 
            a.operations[i].Cancel <- true 
            a.operations[i].Status = CancelStatusType
            close(a.operations[i].Cancel)
            return nil
        }
    }
    // if we get here, couldn't find op w/ the given id
    return ErrIdOp
}

/*
Name: ScheduleReserveAtIntervalOperation
Type: External App Func
Purpose: Used to Schedule a reserve at interval operation, returns ID 
*/
func (a *AppCtx) ScheduleReserveAtIntervalOperation(params ReserveAtIntervalParam) (int64, error) {
    // generate a new id
    id := a.idGen
    a.idGen += 1 

    // check if user provided any login overrides 
    if (params.Login.Email == "" || params.Login.Password == "") {
        // check if user did not, and they haven't logged in through
        // login func, return err
        if(a.loginInfo.Email == "" && a.loginInfo.Password == "") {
            return 0, ErrNoLogin
        }
        // override params login info with stored login ijfo 
        params.Login.Email = a.loginInfo.Email
        params.Login.Password = a.loginInfo.Password
    }

    // make cancel and output channels to manage go thread 
    cancel := make(chan bool)
    output := make(chan OperationResult)

    // add op to internal buffer list 
    a.operations = append(a.operations, Operation{
        ID: id,
        Cancel: cancel,
        Output: output,
        Status: InProgressStatusType,
    })
    // run op
    go a.reserveAtInterval(params, cancel, output)
    return id, nil
}

/*
Name: reserveAtIntervalOperation
Type: Internal App Func
Purpose: This function is intended to run on a separate thread, and tries making
a reservation at a given interval of time
*/
func (a *AppCtx) reserveAtInterval(params ReserveAtIntervalParam, cancel <-chan bool, output chan<- OperationResult){
    // find and store last time from time priority list
    lastTime, err := findLastTime(params.ReservationTimes)

    if err != nil {
        output<-OperationResult{Response: nil, Err: err}     
        close(output)
        return
    }

    for {
        
        // first run pre reservation auth 
        loginResp, err := a.API.Login(api.LoginParam(params.Login))
        
        if err != nil {
            output<-OperationResult{Response: nil, Err: err}     
            close(output)
            return
        }

        // next try reservation 
        reserveResp, err := a.API.Reserve(
            api.ReserveParam{
                LoginResp: *loginResp,
                ReservationTimes: params.ReservationTimes,
                PartySize: params.PartySize,
                VenueID: params.VenueID,
            })

        // if there was an error and it wasn't due to every time being
        // taken, then it's an issue we don't know about
        if err != nil && err != api.ErrNoTable {
            output<-OperationResult{Response: nil, Err: err}     
            close(output)
            return
        }
        if err == api.ErrNoTable {
            // see if last time on list is still in the future,
            // since if it isn't there's no point in trying to reserve it
            if lastTime.After(time.Now()) {
                select {
                case <-time.After(params.RepeatInterval):
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
        // if there's no error, we succeeded
        output<-OperationResult{
            Response: &ReserveAtIntervalResponse{ReservationTime: reserveResp.ReservationTime}, 
            Err: nil,
        }
        close(output)
        return
    }
}

/*
Name: ScheduleReserveAtTimeOperation
Type: External App Func
Purpose: Used to Schedule a reserve at time operation, returns ID 
Note: Most of this logic should probably be merged
with the ScheduleReserveAtIntervalOperation func since it's similar logic
*/
func (a *AppCtx) ScheduleReserveAtTimeOperation(params ReserveAtTimeParam) (int64, error) {
    id := a.idGen
    a.idGen += 1 
    if (params.Login.Email == "" || params.Login.Password == "") {
        if(a.loginInfo.Email == "" && a.loginInfo.Password == "") {
            return 0, ErrNoLogin
        }
        params.Login.Email = a.loginInfo.Email
        params.Login.Password = a.loginInfo.Password
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

/*
Name: reserveAtTimeOperation
Type: Internal App Func
Purpose: This function is intended to run on a separate thread, and tries making
a reservation at a given time
*/
func (a *AppCtx) reserveAtTime(params ReserveAtTimeParam, cancel <-chan bool, output chan<- OperationResult) {
 
    // if this date is not in the future, err 
    if params.RequestTime.Before(time.Now().UTC()) {
        output <- OperationResult{Response: nil, Err: ErrTimeFut}     
        close(output)
        return
    }

    minAuthTime := a.API.AuthMinExpire()
    authDate := params.RequestTime.Add(-1 * minAuthTime)
    if (!authDate.Before(time.Now().UTC())) {
        select {
        case <-time.After(time.Until(authDate)):
            break
        case <-cancel:
            output<- OperationResult{Response: nil, Err:ErrCancel}
            close(output)
            return
        }
    }

    loginResp, err := a.API.Login(api.LoginParam(params.Login))

    if err != nil {
       output<- OperationResult{Response: nil, Err:err}
       close(output)
       return
    }
 
    // sleep with ability to cancel 
    select {
    case <-time.After(time.Until(params.RequestTime)):
    case <-cancel:
        output<- OperationResult{Response: nil, Err:ErrCancel}
        close(output)
        return
    }

    // reserve 
    reserveResp, err := a.API.Reserve(
        api.ReserveParam{
            LoginResp: *loginResp,
            ReservationTimes: params.ReservationTimes,
            PartySize: params.PartySize,
            VenueID: params.VenueID,
        })

    if err != nil {
        output<- OperationResult{Response: nil, Err:err}
        close(output)
        return
    }

    
    // return value if succeeded 
    returnValue := ReserveAtTimeResponse{ ReservationTime: reserveResp.ReservationTime }
    output<- OperationResult{Response: returnValue, Err:nil}
    close(output)
    return
}

/*
Name: Login 
Type: External App Func
Purpose: This function stores loginParams in the
app Ctx if they pass the Login method
*/
func (a *AppCtx) Login(params LoginParam) (error) {
    reqParams := api.LoginParam(params)
    _, err :=  a.API.Login(reqParams)
    if err != nil {
        return err
    }
    a.loginInfo = params
    return nil
}

/*
Name: Search 
Type: External App Func
Purpose: This function performs a search using the underlying 
api
*/
func (a *AppCtx) Search(params SearchParam) (*SearchResponse, error) {
    reqParams := api.SearchParam(params)
    resp, err :=  a.API.Search(reqParams)
    if err != nil {
        return nil, err
    }
    returnValue := SearchResponse(*resp)
    return &returnValue, nil
}

/*
Name: CleanOperation 
Type: External App Func
Purpose: This function removes an operation
from the internal slice of ops, but only does
this if the op is not in progress
*/
func (a *AppCtx) CleanOperation(id int64) (error) {
    for i, operation := range a.operations {
        if operation.ID == id {
            // update before handling
            err := a.updateOperationResult(operation.ID)
            if err != nil {
                return err
            }
            // if op in progress, fail 
            if a.operations[i].Status == InProgressStatusType {
                return ErrCurrOp
            }
            // remove op from op list
            a.operations = append(a.operations[:i], a.operations[i+1:]...)
            return nil
        }
    }
    return ErrIdOp
}

/*
Name: Logout 
Type: External App Func
Purpose: This function removes login info
from the AppCtx if saved from a Login call
*/
func (a *AppCtx) Logout() (error) {
    if (a.loginInfo.Email == "") && (a.loginInfo.Password == "") {
        return ErrNoLogout
    }
    params := LoginParam{}
    a.loginInfo = params
    return nil
}

/*
Name: OperationsToString
Type: External App Func
Purpose: This function stringifies operations
in a use independent manner
*/
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
        // stringify based on op type
        opLstStr += "\tID: " + strconv.FormatInt(operation.ID, 10) + "\n" 
        opLstStr += "\tStatus: " 
        switch operation.Status {
            case InProgressStatusType:
                opLstStr += "In Progress"
            case SuccessStatusType:
                time := operation.Result.Response.Time()
                opLstStr += "Succeeded\n"
                opLstStr += fmt.Sprintf("\tResult: %02d:%02d", time.Hour(), time.Minute())
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

/*
Name: OperationStatus
Type: External App Func
Purpose: This function returns the status of an op 
*/
func (a *AppCtx) OperationStatus(id int64) (OperationStatus, error) {
    for i, operation := range a.operations {
        if operation.ID == id {
            // update before handling
            err := a.updateOperationResult(operation.ID)
            if err != nil {
                return InProgressStatusType, err
            }
            return a.operations[i].Status, nil
        }
    }
    return InProgressStatusType, ErrIdOp
}
