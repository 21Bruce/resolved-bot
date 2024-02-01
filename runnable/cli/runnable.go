/*
Author: Bruce Jagid
Created On: Aug 12, 2023
*/
package cli

import (
    "bufio"
    "io"
    "fmt"
    "strconv"
    "strings"
    "github.com/21Bruce/resolved-server/app"
    "github.com/21Bruce/resolved-server/api"
    "github.com/21Bruce/resolved-server/cli"
    "os"
    "errors"
    "time"
)

var (
    // Error if we can't parse date properly
    ErrInvDate = errors.New("invalid date format")
    // Error if we can't parse table type properly
    ErrInvTableType = errors.New("invalid table type")
)

/*
Name: ResolvedCLI
Type: External CLI Struct
Purpose: Encapsulate the state
and initial config of the system
*/
type ResolvedCLI struct {
    AppCtx      app.AppCtx 
    In          io.Reader
    Out         io.Writer
    Err         io.Writer
    parseCtx    cli.ParseCtx
}

/*
Name: parseSearch
Type: Internal Func
Purpose: Perform some extra
parsing on top of the cli parser
and return the field for use in the main
handler
*/
func parseSearch(in map[string][]string) (string, int, error){

    name := in["n"][0]
    limit := 0
    if in["l"] != nil {
        limitRes, err := strconv.Atoi(in["l"][0])
        if err != nil {
            return "", 0 , err
        }
        limit = limitRes
    }
    return name, limit, nil

}

/*
Name: handleSearch
Type: Internal Func
Purpose: This function
is the handler for the 'search' command.
It is responsible for taking in the validated
flag args and returning a string
of the search results
*/
func (c *ResolvedCLI) handleSearch(in map[string][]string) (string, error) {
    name, limit, err := parseSearch(in)
    if err != nil {
        return "", err
    }
    searchParams := app.SearchParam{Name: name, Limit: limit}
    resp, err := c.AppCtx.Search(searchParams)
    if err != nil {
        return "", err
    }
    retVal := api.SearchResponse(*resp)
    return retVal.ToString(), nil
}

/*
Name: handleQuit
Type: Internal Func
Purpose: This function
is the handler for the 'exit' and 'quit' commands.
It is responsible for exiting the CLI
*/
func (c *ResolvedCLI) handleQuit(in map[string][]string) (string, error) {
    os.Exit(0)
    return "", nil
}

/*
Name: flagToShortStr 
Type: Internal Func
Purpose: This function assists in
the 'help' command, this all should be moved
to the cli top level
*/
func flagToShortStr(flag cli.Flag) (string) {
    flagStr := " [-" + flag.Name 
    if flag.LongName != "" {
        flagStr += "|--" + flag.LongName
    }
    flagStr += "]"
    return flagStr
}

/*
Name: handleHelp 
Type: Internal Func
Purpose: This function is the handler
for the 'help' command, It is responsible
for printing out helpful info for each
command, it really should be moved to the
cli top level
*/
func (c *ResolvedCLI) handleHelp(in map[string][]string) (string, error) {
    helpStr := "Commands: \n"
    for _, cmd := range c.parseCtx.Commands {
        helpStr += "\t" + cmd.Name 
        for _, flag := range cmd.Flags {
            helpStr += flagToShortStr(flag)
        }
        helpStr += ": "+ cmd.Description + "\n"
        for _, flag := range cmd.Flags {
            helpStr += "\t\t" + flagToShortStr(flag) + ": "  + flag.Description + "\n"
        }
    }

    return helpStr, nil 
}

/*
Name: handleList 
Type: Internal Func
Purpose: This function is the handler
for the 'list' command, It is responsible
for printing out a history of operations
from the AppCtx
*/
func (c *ResolvedCLI) handleList(in map[string][]string) (string, error) {
    return c.AppCtx.OperationsToString()
}

/*
Name: parseRats 
Type: Internal Func
Purpose: This function helps with parsing
for the main 'rats' handler function
*/
func (c *ResolvedCLI) parseRats(in map[string][]string) (*app.ReserveAtTimeParam, error) {
    req := app.ReserveAtTimeParam{}
    // if we have login info, overwrite the default
    if in["e"] != nil {
        req.Login.Email = in["e"][0]
    }
    if in["p"] != nil {
        req.Login.Password = in["p"][0]
    }
    if in["t"] != nil {
	    req.TableTypes = make([]api.TableType, len(in["t"]), len(in["t"])) 
	    for i := 0; i < len(in["t"]); i++ {
            currType := strings.ToLower(in["t"][i])
	        if strings.Contains(currType, string(api.DiningRoom)) {
                req.TableTypes[i] = api.DiningRoom
	        } else if strings.Contains(currType, string(api.Indoor)) {
                req.TableTypes[i] = api.Indoor
	        } else if strings.Contains(currType, string(api.Outdoor)) {
                req.TableTypes[i] = api.Outdoor
	        } else if strings.Contains(currType, string(api.Patio)) {
                req.TableTypes[i] = api.Patio
	        } else if strings.Contains(currType, string(api.Bar)) {
                req.TableTypes[i] = api.Bar
	        } else if strings.Contains(currType, string(api.Lounge)) {
                req.TableTypes[i] = api.Lounge
	        } else if strings.Contains(currType, string(api.Booth)) {
                req.TableTypes[i] = api.Booth
	        } else {
                return nil, ErrInvTableType
            }
	    }	
    }
    id, err := strconv.ParseInt(in["v"][0], 10, 64)
    if err != nil {
        return nil, err
    }
    req.VenueID = id
    rawResDay := in["resD"][0]
    resDaySplt := strings.Split(rawResDay, ":")
    if len(resDaySplt) != 3 {
        return nil, ErrInvDate
    }

    reqYear, err := strconv.Atoi(resDaySplt[0])
    if err != nil {
        return nil, err
    }
    reqMonth, err := strconv.Atoi(resDaySplt[1])
    if err != nil {
        return nil, err
    }
    reqDay, err := strconv.Atoi(resDaySplt[2])
    if err != nil {
        return nil, err
    }

    req.ReservationTimes = make([]time.Time, len(in["resT"]), len(in["resT"]))
    for i, timeStr := range in["resT"] {
        timeSplt := strings.Split(timeStr, ":")        
        if len(timeSplt) != 2 {
            return nil, ErrInvDate
        }
        reqHour, err := strconv.Atoi(timeSplt[0])
        if err != nil {
            return nil, err
        }
        reqMin, err := strconv.Atoi(timeSplt[1])
        if err != nil {
            return nil, err
        }
        req.ReservationTimes[i] = time.Date(reqYear, time.Month(reqMonth), reqDay, reqHour, reqMin, 0, 0, time.Local)

    }
    ps, err := strconv.ParseInt(in["ps"][0], 10, 64)
    if err != nil {
        return nil, err
    }
    req.PartySize = int(ps)
    rawReqDate := in["reqD"][0]
    reqDateSplt := strings.Split(rawReqDate, ":")

    if len(reqDateSplt) != 5 {
        return nil, ErrInvDate
    }

    year, err := strconv.ParseInt(reqDateSplt[0], 10, 64) 
    if err != nil {
        return nil, err
    }
    month, err := strconv.ParseInt(reqDateSplt[1], 10, 64) 
    if err != nil {
        return nil, err
    }
    day, err := strconv.ParseInt(reqDateSplt[2], 10, 64) 
    if err != nil {
        return nil, err
    }
    hour, err := strconv.ParseInt(reqDateSplt[3], 10, 64) 
    if err != nil {
        return nil, err
    }
    minute, err := strconv.ParseInt(reqDateSplt[4], 10, 64) 
    if err != nil {
        return nil, err
    }
    timeLoc := time.Date(int(year), time.Month(int(month)), int(day), int(hour), int(minute), 0, 0, time.Local)
    timeUTC := timeLoc.UTC()
    req.RequestTime = timeUTC
    return &req, nil
}

/*
Name: handleRats 
Type: Internal Func
Purpose: This function is the handler
for the 'rats' command. It's goal is to
take the values defined in each flag field
and schedule a reserve at time operation
in the AppCtx
*/
func (c *ResolvedCLI) handleRats(in map[string][]string) (string, error) {
    req, err := c.parseRats(in)
    if err != nil {
        return "", err
    }
    id, err := c.AppCtx.ScheduleReserveAtTimeOperation(*req)
    if err != nil {
        return "", err
    }
    idstr := strconv.FormatInt(id, 10)
    // if successful, tell user which id the operation is linked to
    retstr := "Successfully started rats operation with ID " + idstr 
    return retstr, nil 
}

/*
Name: parseRais 
Type: Internal Func
Purpose: This function helps with parsing
for the main 'rats' handler function.
This function is very similiar to parseRats
and can probably be merged a little
*/
func (c *ResolvedCLI) parseRais(in map[string][]string) (*app.ReserveAtIntervalParam, error) {
    req := app.ReserveAtIntervalParam{}
    if in["e"] != nil {
        req.Login.Email = in["e"][0]
    }
    if in["p"] != nil {
        req.Login.Password = in["p"][0]
    } 
    if in["t"] != nil {
	    req.TableTypes = make([]api.TableType, len(in["t"]), len(in["t"])) 
	    for i := 0; i < len(in["t"]); i++ {
            currType := in["t"][i]
	        if strings.Contains(currType, string(api.DiningRoom)) {
                req.TableTypes[i] = api.DiningRoom
	        } else if strings.Contains(currType, string(api.Indoor)) {
                req.TableTypes[i] = api.Indoor
	        } else if strings.Contains(currType, string(api.Outdoor)) {
                req.TableTypes[i] = api.Outdoor
	        } else if strings.Contains(currType, string(api.Patio)) {
                req.TableTypes[i] = api.Patio
	        } else if strings.Contains(currType, string(api.Bar)) {
                req.TableTypes[i] = api.Bar
	        } else if strings.Contains(currType, string(api.Lounge)) {
                req.TableTypes[i] = api.Lounge
	        } else if strings.Contains(currType, string(api.Booth)) {
                req.TableTypes[i] = api.Booth
	        } else {
                return nil, ErrInvTableType
            }
	    }	
    }    
    id, err := strconv.ParseInt(in["v"][0], 10, 64)
    if err != nil {
        return nil, err
    }
    req.VenueID = id
    rawResDay := in["resD"][0]
    resDaySplt := strings.Split(rawResDay, ":")
    if len(resDaySplt) != 3 {
        return nil, ErrInvDate
    }
    reqYear, err := strconv.Atoi(resDaySplt[0])
    if err != nil {
        return nil, err
    }
    reqMonth, err := strconv.Atoi(resDaySplt[1])
    if err != nil {
        return nil, err
    }
    reqDay, err := strconv.Atoi(resDaySplt[2])
    if err != nil {
        return nil, err
    }
    req.ReservationTimes = make([]time.Time, len(in["resT"]), len(in["resT"]))
    for i, timeStr := range in["resT"] {
        timeSplt := strings.Split(timeStr, ":")        
        if len(timeSplt) != 2 {
            return nil, ErrInvDate
        }
        reqHour, err := strconv.Atoi(timeSplt[0])
        if err != nil {
            return nil, err
        }
        reqMin, err := strconv.Atoi(timeSplt[1])
        if err != nil {
            return nil, err
        }
        req.ReservationTimes[i] = time.Date(reqYear, time.Month(reqMonth), reqDay, reqHour, reqMin, 0, 0, time.Local)
    }
    ps, err := strconv.ParseInt(in["ps"][0], 10, 64)
    if err != nil {
        return nil, err
    }
    req.PartySize = int(ps)
    rawRepInt := in["i"][0]
    repIntSplt := strings.Split(rawRepInt, ":")

    if len(repIntSplt) != 2 {
        return nil, ErrInvDate
    }

    repHour, err := strconv.Atoi(repIntSplt[0])
    if err != nil {
        return nil, err
    }
    repMin, err := strconv.Atoi(repIntSplt[1])
    if err != nil {
        return nil, err
    }
    req.RepeatInterval = time.Hour * time.Duration(repHour) + time.Minute * time.Duration(repMin)

    return &req, nil
}

/*
Name: handleRais 
Type: Internal Func
Purpose: This function is the handler
for the 'rais' command. Its goal is to
take the values defined in each flag field
and schedule a reserve at interval operation
in the AppCtx
*/
func (c *ResolvedCLI) handleRais(in map[string][]string) (string, error) {
    req, err := c.parseRais(in)
    if err != nil {
        return "", err
    }
    id, err := c.AppCtx.ScheduleReserveAtIntervalOperation(*req)
    if err != nil {
        return "", err
    }
    idstr := strconv.FormatInt(id, 10)
    // if successful, tell user and print ID of new operation
    retstr := "Successfully started rais operation with ID " + idstr 
    return retstr, nil 
}

/*
Name: handleLogin 
Type: Internal Func
Purpose: This function is the handler
for the 'login' command, its goal is to
save the login info on the appctx if its
valid
*/
func (c *ResolvedCLI) handleLogin(in map[string][]string) (string, error) {
    req := app.LoginParam{
        Email: in["e"][0],
        Password: in["p"][0],
    }
    err := c.AppCtx.Login(req)
    if err != nil {
        return "", err
    }
    // if successful, tell user
    return "Successfully Logged In", nil
}

/*
Name: handleLogout
Type: Internal Func
Purpose: This function is the handler
for the 'logout' command, its goal is to
erase login info from the appctx
*/
func (c *ResolvedCLI) handleLogout(in map[string][]string) (string, error) {
    err := c.AppCtx.Logout()
    if err != nil {
        return "", err
    }
    // if successful, tell user
    return "Successfully Logged Out", nil
}

/*
Name: handleCancel
Type: Internal Func
Purpose: This function is the handler
for the 'cancel' command, its goal is to
cancel all operations given the id list
in the -i field. We only cancel all or no 
operations, so we check before if they are
valid to be cancelled
*/
func (c *ResolvedCLI) handleCancel(in map[string][]string) (string, error) {
    for _, idStr := range in["i"] {
        id, err := strconv.ParseInt(idStr, 10, 64)
        if err != nil {
            return "", err
        }
        stat, err := c.AppCtx.OperationStatus(id)
        if err != nil {
            return "", err
        }
        if stat != app.InProgressStatusType {
            return "", app.ErrFinOp
        }
    }
    for _, idStr := range in["i"] {
        // errs checked above
        id, _ := strconv.ParseInt(idStr, 10, 64)
        c.AppCtx.CancelOperation(id)
    }
    return "Cancelled Operations Successfully", nil 
}

/*
Name: handleClean
Type: Internal Func
Purpose: This function is the handler
for the 'clean' command, its goal is to
cancel all operations given the id list
in the -i field. We only clean all or no 
operations, so we check before if they are
valid to be cleaned
*/
func (c *ResolvedCLI) handleClean(in map[string][]string) (string, error) {
    for _, idStr := range in["i"] {
        id, err := strconv.ParseInt(idStr, 10, 64)
        if err != nil {
            return "", err
        }
        stat, err := c.AppCtx.OperationStatus(id)
        if err != nil {
            return "", err
        }
        if stat == app.InProgressStatusType {
            return "", app.ErrCurrOp
        }
    }
    for _, idStr := range in["i"] {
        // errs checked above
        id, _ := strconv.ParseInt(idStr, 10, 64)
        c.AppCtx.CleanOperation(id)
    }
    return "Cleaned Operations Successfully", nil 
}

/*
Name: initParseCtx 
Type: Internal Func
Purpose: This function initializes
the parse ctx with the above handlers
and command info
*/
func (c *ResolvedCLI) initParseCtx() {
    // 'search' command
    searchCommand := cli.Command{
        Name: "search",
        Description: "Finds restaurant info",
        Flags: []cli.Flag{
            cli.Flag{
                Name: "n",
                LongName: "name",
                Description: "This flag is required. It takes one text input, the name of the restaurant",
                ValidationCtx: cli.FlagValidationCtx{
                    Required: true,
                    MinArgs: 1,
                    MaxArgs: 1,
                },
            },
            cli.Flag{
                Name: "l",
                LongName: "limit",
                Description: "This flag is optional. It takes one number input, the max amount of results to return",
                ValidationCtx: cli.FlagValidationCtx{
                    Required: false,
                    MinArgs: 1,
                    MaxArgs: 1,
                },
            },
        },
        Handler: c.handleSearch,
    }

    // 'rats' command
    ratsCommand := cli.Command{
        Name: "rats",
        Description: "Reserve At Time Scheduler",
        Flags: []cli.Flag{
            cli.Flag{
                Name: "e",
                LongName: "email",
                Description: "This flag is optional if already logged in using Login command. Specifies login email",
                ValidationCtx: cli.FlagValidationCtx{
                    Required: false,
                    MinArgs: 1,
                    MaxArgs: 1,
                },
            },
            cli.Flag{
                Name: "p",
                LongName: "password",
                Description: "This flag is optional if already logged in using Login command. Specifies login password",
                ValidationCtx: cli.FlagValidationCtx{
                    Required: false,
                    MinArgs: 1,
                    MaxArgs: 1,
                },
            },
            cli.Flag{
                Name: "v",
                LongName: "venue-id",
                Description: "This flag is required. Specifies the venueu id(use search to find by name)",
                ValidationCtx: cli.FlagValidationCtx{
                    Required: true,
                    MinArgs: 1,
                    MaxArgs: 1,
                },
            },
	        cli.Flag{
		        Name: "t",
		        LongName: "table",
		        Description: "This flag is optional. Used to set the type of table in order of preference",
		        ValidationCtx: cli.FlagValidationCtx{
		            Required: false,
		            MinArgs: 1,
		            MaxArgs: cli.InfiniteArgs, 
		        },
	        },
            cli.Flag{
                Name: "resD",
                LongName: "reservation-day",
                Description: "This flag is required. Specifies the day for the reservation in yyyy:mm:dd format",
                ValidationCtx: cli.FlagValidationCtx{
                    Required: true,
                    MinArgs: 1,
                    MaxArgs: 1,
                },
            },
            cli.Flag{
                Name: "resT",
                LongName: "reservation-times",
                Description: "This flag is required. Specifies the priority time list for the reservation in hh:mm format",
                ValidationCtx: cli.FlagValidationCtx{
                    Required: true,
                    MinArgs: 1,
                    MaxArgs: cli.InfiniteArgs,
                },
            },
            cli.Flag{
                Name: "reqD",
                LongName: "request-date",
                Description: "This flag is required. Specifies the date to send request in yyyy:mm:dd:hh:mm format",
                ValidationCtx: cli.FlagValidationCtx{
                    Required: true,
                    MinArgs: 1,
                    MaxArgs: 1,
                },
            },
            cli.Flag{
                Name: "ps",
                LongName: "party-size",
                Description: "This flag is required. Specifies the size of party",
                ValidationCtx: cli.FlagValidationCtx{
                    Required: true,
                    MinArgs: 1,
                    MaxArgs: 1,
                },
            },
 
        },
        Handler: c.handleRats,
    }

    // 'rais' command
    raisCommand := cli.Command{
        Name: "rais",
        Description: "Reserve At Interval Scheduler",
        Flags: []cli.Flag{
            cli.Flag{
                Name: "e",
                LongName: "email",
                Description: "This flag is optional if already logged in using Login command. Specifies login email",
                ValidationCtx: cli.FlagValidationCtx{
                    Required: false,
                    MinArgs: 1,
                    MaxArgs: 1,
                },
            },
            cli.Flag{
                Name: "p",
                LongName: "password",
                Description: "This flag is optional if already logged in using Login command. Specifies login password",
                ValidationCtx: cli.FlagValidationCtx{
                    Required: false,
                    MinArgs: 1,
                    MaxArgs: 1,
                },
            },
            cli.Flag{
                Name: "v",
                LongName: "venue-id",
                Description: "This flag is required. Specifies the venueu id(use search to find by name)",
                ValidationCtx: cli.FlagValidationCtx{
                    Required: true,
                    MinArgs: 1,
                    MaxArgs: 1,
                },
            },
	        cli.Flag{
		        Name: "t",
		        LongName: "table",
		        Description: "This flag is optional. Used to set the type of table in order of preference",
		        ValidationCtx: cli.FlagValidationCtx{
		            Required: false,
		            MinArgs: 1,
		            MaxArgs: cli.InfiniteArgs, 
		        },
	        },
            cli.Flag{
                Name: "resD",
                LongName: "reservation-day",
                Description: "This flag is required. Specifies the day for the reservation in yyyy:mm:dd format",
                ValidationCtx: cli.FlagValidationCtx{
                    Required: true,
                    MinArgs: 1,
                    MaxArgs: 1,
                },
            },
            cli.Flag{
                Name: "resT",
                LongName: "reservation-times",
                Description: "This flag is required. Specifies the priority time list for the reservation in hh:mm format",
                ValidationCtx: cli.FlagValidationCtx{
                    Required: true,
                    MinArgs: 1,
                    MaxArgs: cli.InfiniteArgs,
                },
            },
            cli.Flag{
                Name: "i",
                LongName: "interval",
                Description: "This flag is required. Specifies the interval to send request on in hh:mm format",
                ValidationCtx: cli.FlagValidationCtx{
                    Required: true,
                    MinArgs: 1,
                    MaxArgs: 1,
                },
            },
            cli.Flag{
                Name: "ps",
                LongName: "party-size",
                Description: "This flag is required. Specifies the size of party",
                ValidationCtx: cli.FlagValidationCtx{
                    Required: true,
                    MinArgs: 1,
                    MaxArgs: 1,
                },
            },
 
        },
        Handler: c.handleRais,
    }


    // 'list' command
    listCommand := cli.Command{
        Name: "list",
        Description: "List all operations",
        Flags: []cli.Flag{},
        Handler: c.handleList,
    }

    // 'login' command
    loginCommand := cli.Command{
        Name: "login",
        Description: "Set login defaults",
        Flags: []cli.Flag{
            cli.Flag{
                Name: "e",
                LongName: "email",
                Description: "This flag is required. Provides login email",
                ValidationCtx: cli.FlagValidationCtx{
                    Required: true,
                    MaxArgs: 1,
                    MinArgs: 1,
                },
            },
            cli.Flag{
                Name: "p",
                LongName: "password",
                Description: "This flag is required. Provides login password",
                ValidationCtx: cli.FlagValidationCtx{
                    Required: true,
                    MaxArgs: 1,
                    MinArgs: 1,
                },
            },
        },
        Handler: c.handleLogin,
    }

    // 'logout' command
    logoutCommand := cli.Command{
        Name: "logout",
        Description: "Clear default login credentials",
        Flags: []cli.Flag{},
        Handler: c.handleLogout,
    }

    // 'cancel' command
    cancelCommand := cli.Command{
        Name: "cancel",
        Description: "Cancel operations given ids",
        Flags: []cli.Flag{
            cli.Flag{
                Name: "i",
                LongName: "id",
                Description: "This flag is required. It takes one to unmeasured number inputs, the ids of operations",
                ValidationCtx: cli.FlagValidationCtx{
                    Required: true,
                    MinArgs: 1,
                    MaxArgs: cli.InfiniteArgs,
                },
            },
        },
        Handler: c.handleCancel,
    }

    // 'clean' command
    cleanCommand := cli.Command{
        Name: "clean",
        Description: "Clean operations given ids",
        Flags: []cli.Flag{
            cli.Flag{
                Name: "i",
                LongName: "id",
                Description: "This flag is required. It takes one to unmeasured number inputs, the ids of operations",
                ValidationCtx: cli.FlagValidationCtx{
                    Required: true,
                    MinArgs: 1,
                    MaxArgs: cli.InfiniteArgs,
                },
            },
        },
        Handler: c.handleClean,
    }

    // 'quit' command
    quitCommand := cli.Command{
        Name: "quit",
        Description: "Exits the CLI",
        Flags: []cli.Flag{},
        Handler: c.handleQuit,
    }

    // 'exit' command, same handler as 'quit' command
    exitCommand := cli.Command{
        Name: "exit",
        Description: "Exits the CLI",
        Flags: []cli.Flag{},
        Handler: c.handleQuit,
    }

    // 'help' command
    helpCommand := cli.Command{
        Name: "help",
        Description: "Displays helpful info about commands",
        Flags: []cli.Flag{},
        Handler: c.handleHelp,
    }

    // we init the parseCtx with the above info
    c.parseCtx = cli.ParseCtx{
        OpenDelim: "[",
        CloseDelim: "]",
        Commands: []cli.Command{
            searchCommand,
            cancelCommand,
            cleanCommand,
            listCommand,
            loginCommand,
            logoutCommand,
            ratsCommand,
            raisCommand,
            quitCommand,
            exitCommand,
            helpCommand,
        },
    }
}

/*
Name: Run 
Type: External Func
Purpose: This function inits the 
parse ctx and enters into an infinite
loop, starting the REPL
*/
func (c *ResolvedCLI) Run() (error) {
    // init the parse ctx w/the above handler
    c.initParseCtx()
    scanner := bufio.NewScanner(c.In)
    // print welcome msg 
    fmt.Fprintln(c.Out, "Welcome to the Resolved CLI! For Help type 'help'") 
    for {
        // print prompt 
        fmt.Fprint(c.Out, "resolved(0.1.0)>> ") 
        scanner.Scan()
        if err := scanner.Err(); err != nil {
            fmt.Fprintln(c.Err, err);
        }
        // parse input 
        result, err := c.parseCtx.Parse(scanner.Text()) 
        if err != nil {
            fmt.Fprint(c.Err, "ERROR: ")
            fmt.Fprintln(c.Err, err)
        } else  {
            fmt.Fprintln(c.Out, result) 
        }
    }
    return nil
}


