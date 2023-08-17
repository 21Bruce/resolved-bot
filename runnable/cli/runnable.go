package cli

import (
    "bufio"
    "io"
    "os"
    "fmt"
    "strings"
    "strconv"
    "errors"
    "github.com/21Bruce/resolved-server/api"
)

var (
    ErrNoCmd = errors.New("command does not exist")
    ErrNoFmt = errors.New("improper format")
    ErrNoName = errors.New("name required for search, use syntax '-n name'")
)
type ResolvedCLI struct {
    API     api.API
    In      io.Reader
    Out     io.Writer
    Err     io.Writer
}

var openDelim string  = "["
var closeDelim string = "]"
 
func processHelp() (string, error) {
    helpStr := "\nCommands: \n" 
    helpCmd := "\thelp: displays commands"
    exitCmd := "\texit/quit: leaves command prompt"
    searchCmd := "\tsearch [-n name] [-l limit]: " +
    "searches for 'name' and grabs at most 'limit' responses. " +
    "\n\tIf name is multiple words long, wrap name with square brackets " + 
    openDelim + 
    closeDelim
    helpStr += helpCmd + "\n" + exitCmd + "\n" + searchCmd  + "\n"
    return helpStr, nil
}

func parseNameStr(input []string) (name string, offset int, err error) {
    name = ""
    offset = 0
    err = nil
    args := len(input)
    for i := 0; i < args; i++ {
        offset += 1
        argLen := len(input[i])
        if i == 0  {
            if input[i][0] == openDelim[0] && input[i][argLen-1] != closeDelim[0]{
                name += input[i][1:]
            } else if input[i][0] == openDelim[0] && input[i][argLen-1] == closeDelim[0]{
                name = input[i][1:argLen-1]
                return
            } else {
                name += input[i]
                return    
            }
        }
        if (input[i][argLen-1] == closeDelim[0]) {
            name += input[i][:argLen-1]
            return
        }
        name += " "
    }
    err = ErrNoFmt
    return
}

func parseSearch(input []string) (name string, limit *int, err error) {
    name = ""
    limit = nil 
    err = nil
    args := len(input)
    for i := 1; i < args; i++ {
        switch input[i] {
            case "-n" :
                if (i + 1) > args  || name != ""{
                    err = ErrNoFmt
                    break
                }
                nameRet, offset, errRet:= parseNameStr(input[i+1:])
                if err = errRet; err != nil {
                    return
                }
                name = nameRet
                i += offset
            case "-l" :
                if (i + 1) > args || limit != nil{
                    err = ErrNoFmt
                    break
                }
                i += 1
                limitRaw, err := strconv.ParseInt(input[i], 10, 64)
                if err != nil {
                    break
                }
                limitInt := int(limitRaw)
                limit = &limitInt
            default:
                err = ErrNoFmt
                break
        }

    }
    if name == "" {
        err = ErrNoName
    }
    return name, limit, err
}

func processSearch(a api.API, name string, limit *int) (string, error) {
    var limitInt int
    if limit == nil {
        limitInt = 0
    } else {
        limitInt = *limit
    }
    searchParams := api.SearchParam{Name: name, Limit: limitInt}
    resp, err := a.Search(searchParams)
    if err != nil {
        return "", err
    }
    return resp.ToString(), nil
}

func processCmd(c *ResolvedCLI, input string) (string, error) {
    inputSplit := strings.Split(input, " ")
    var inputParsed []string = make([]string, 0, len(inputSplit))
    for _, e := range inputSplit {
        if e != "" {
            inputParsed = append(inputParsed, e)
        }
    }
    switch inputParsed[0] {
        case "help":
            return processHelp()
        case "quit" :
             os.Exit(0)
        case "exit" :
             os.Exit(0)
        case "search":
            name, limit, err := parseSearch(inputParsed)
            if err != nil {
                return "", err
            }
            return processSearch(c.API, name, limit)
        default:
            return "", ErrNoCmd
    }
    return "", nil
}

func (c *ResolvedCLI) Run() {
    scanner := bufio.NewScanner(c.In)
    fmt.Fprintln(c.Out, "Welcome to the Resolved CLI! For Help type 'help'") 
    for {
        fmt.Fprint(c.Out, "resolved(0.1.0)>> ") 
        scanner.Scan()
        if err := scanner.Err(); err != nil {
            fmt.Fprintln(c.Err, err);
        }
        result, err := processCmd(c, scanner.Text()) 
        if err != nil {
            fmt.Fprint(c.Err, "ERROR: ")
            fmt.Fprintln(c.Err, err)
        } else  {
            fmt.Fprintln(c.Out, result) 
        }
    }
}


