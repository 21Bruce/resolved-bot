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
    ErrInvDate = errors.New("invalid date format")
)

type ResolvedCLI struct {
    AppCtx      app.AppCtx 
    In          io.Reader
    Out         io.Writer
    Err         io.Writer
    parseCtx    cli.ParseCtx
}

func validateSearch(in map[string][]string) (string, int, error){

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

func (c *ResolvedCLI) handleSearch(in map[string][]string) (string, error) {
    name, limit, err := validateSearch(in)
    if err != nil {
        return "", err
    }
    searchParams := app.SearchParam{Name: name, Limit: limit}
    resp, err := c.AppCtx.Search(searchParams)
    if err != nil {
        return "", err
    }
    return resp.ToString(), nil
}

func (c *ResolvedCLI) handleQuit(in map[string][]string) (string, error) {
    os.Exit(0)
    return "", nil
}

func flagToShortStr(flag cli.Flag) (string) {
    flagStr := " [-" + flag.Name 
    if flag.LongName != "" {
        flagStr += "|--" + flag.LongName
    }
    flagStr += "]"
    return flagStr
}

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

func (c *ResolvedCLI) handleList(in map[string][]string) (string, error) {
    return c.AppCtx.OperationsToString()
}

func (c *ResolvedCLI) parseRats(in map[string][]string) (*app.ReserveAtTimeParam, error) {
    req := app.ReserveAtTimeParam{}
    if in["e"] != nil {
        req.Email = in["e"][0]
    }
    if in["p"] != nil {
        req.Password = in["p"][0]
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
    req.Year = resDaySplt[0] 
    req.Month = resDaySplt[1]
    req.Day = resDaySplt[2]
    req.ReservationTimes = make([]api.Time, len(in["resT"]), len(in["resT"]))
    for i, timeStr := range in["resT"] {
        timeSplt := strings.Split(timeStr, ":")        
        if len(timeSplt) != 2 {
            return nil, ErrInvDate
        }
        req.ReservationTimes[i].Hour = timeSplt[0]
        req.ReservationTimes[i].Minute = timeSplt[1]
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
    yearu, monthu, dayu := timeUTC.Date()
    req.RequestYear = strconv.Itoa(yearu)
    req.RequestMonth = strconv.Itoa(int(monthu))
    req.RequestDay = strconv.Itoa(dayu)
    req.RequestTime.Hour = strconv.Itoa(timeUTC.Hour())
    req.RequestTime.Minute =  strconv.Itoa(timeUTC.Minute())
    return &req, nil
}

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
    retstr := "Successfully started rats operation with ID " + idstr 
    return retstr, nil 
}

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

func (c *ResolvedCLI) initParseCtx() {
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

    ratsCommand := cli.Command{
        Name: "rats",
        Description: "Reserve At Time Scheduler",
        Flags: []cli.Flag{
            cli.Flag{
                Name: "e",
                LongName: "email",
                Description: "This flag is optional. Specifies login email",
                ValidationCtx: cli.FlagValidationCtx{
                    Required: false,
                    MinArgs: 1,
                    MaxArgs: 1,
                },
            },
            cli.Flag{
                Name: "p",
                LongName: "password",
                Description: "This flag is optional. Specifies login password",
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

    listCommand := cli.Command{
        Name: "list",
        Description: "list all operations",
        Flags: []cli.Flag{},
        Handler: c.handleList,
    }

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

    quitCommand := cli.Command{
        Name: "quit",
        Description: "Exits the CLI",
        Flags: []cli.Flag{},
        Handler: c.handleQuit,
    }

    exitCommand := cli.Command{
        Name: "exit",
        Description: "Exits the CLI",
        Flags: []cli.Flag{},
        Handler: c.handleQuit,
    }

    helpCommand := cli.Command{
        Name: "help",
        Description: "Displays helpful info about commands",
        Flags: []cli.Flag{},
        Handler: c.handleHelp,
    }

    c.parseCtx = cli.ParseCtx{
        OpenDelim: "[",
        CloseDelim: "]",
        Commands: []cli.Command{
            searchCommand,
            cancelCommand,
            listCommand,
            ratsCommand,
            quitCommand,
            exitCommand,
            helpCommand,
        },
    }
}

func (c *ResolvedCLI) Run() {
    c.initParseCtx()
    scanner := bufio.NewScanner(c.In)
    fmt.Fprintln(c.Out, "Welcome to the Resolved CLI! For Help type 'help'") 
    for {
        fmt.Fprint(c.Out, "resolved(0.1.0)>> ") 
        scanner.Scan()
        if err := scanner.Err(); err != nil {
            fmt.Fprintln(c.Err, err);
        }
        result, err := c.parseCtx.Parse(scanner.Text()) 
        if err != nil {
            fmt.Fprint(c.Err, "ERROR: ")
            fmt.Fprintln(c.Err, err)
        } else  {
            fmt.Fprintln(c.Out, result) 
        }
    }
}


