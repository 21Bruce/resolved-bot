/*
Author: Bruce Jagid
Created On: Aug 12, 2023
*/
package cli

import (
    "strings"
    "errors"
)

// For use in the FlagValidationCtx field MaxArgs
const (
    InfiniteArgs = -1
)

var (
    ErrNoCmd = errors.New("command unrecognized")
    ErrNoFlg = errors.New("flag unrecognized")
    ErrRpFlg = errors.New("flag repeated")
    ErrNstGrp = errors.New("found nested group attempt")
    ErrNoGrp = errors.New("unclosed group")
    ErrMissReq = errors.New("missing required flag")
    ErrMulArg = errors.New("too many arguments for a flag")
    ErrNoArg = errors.New("too few arguments for a flag")
)

/*
Name: FlagValidationCtx
Type: External CLI struct 
Purpose: Define some common validation
rules for flags
*/
type FlagValidationCtx struct{
    MaxArgs     int    
    MinArgs     int
    Required    bool
}

/*
Name: Flag
Type: External CLI struct 
Purpose: Define a representation
of a flag to be used in the ParseCtx
*/
type Flag struct {
    Name            string
    LongName        string
    Description     string
    ValidationCtx   FlagValidationCtx
}

/*
Name: Command 
Type: External CLI struct 
Purpose: Define a representation of
a command that can be used in ParseCtx
*/
type Command struct {
    Name            string
    Description     string
    Flags           []Flag 
    Handler         func(in map[string][]string)(string, error)
}

/*
Name: ParseCtx 
Type: External CLI struct 
Purpose: Define a representation a
command syntax ruleset
*/
type ParseCtx struct {
    Commands    []Command
    OpenDelim   string
    CloseDelim  string
}

/*
Name: splitOn 
Type: Internal func
Purpose: Same as split but leaves delim as its own "token". 
For example splitOn("a/b/c", "/") == []string{"a" "/" "b" "/" "c"}
*/
func splitOn(in string, delim string) []string {
    inDelim := strings.Split(in, delim)
    inProc := make([]string, 0, len(inDelim) * 2 - 1)
    for i, v := range inDelim {
        inProc = append(inProc, v)    
        if i != len(inDelim) - 1 {
            inProc = append(inProc, delim)
        }
    }
    return inProc
}

/*
Name: Tokenize 
Type: External CLI func
Purpose: Take the input command
string and split it into tokens
with grouped tokens defined by
pc.OpenDelim and pc.CloseDelim
*/
func (pc *ParseCtx) Tokenize (in string) ([]string, error) {
    // creates a tokenized slice using input string on open delim
    inOpenDelimProc := splitOn(in, pc.OpenDelim)
    // alloc a slice to process close delim tokens
    inCloseDelimProc := make([]string, 0,len(inOpenDelimProc))
    for _,v := range inOpenDelimProc {
        // check each open delim recognized token for close delims
        closeSplitted := splitOn(v, pc.CloseDelim)
        inCloseDelimProc = append(inCloseDelimProc, closeSplitted...)
    }
    // isGrp is a flag indicating we are processing a group token
    isGrp := false
    // store group token in grpToken
    grpToken := ""
    // Arbitrary capacity
    spaceProc := make([]string, 0, 3 * len(inOpenDelimProc))
    for _, dSpltToken := range inCloseDelimProc {
        // if we find an open delim and we arent processing 
        // a group, then start processing a group
        if dSpltToken == pc.OpenDelim && !isGrp {
            isGrp = true
            grpToken = ""
            continue
        }
        // if we find a close delim and we arent processing 
        // a group, then append the recorded group token and
        // stop processing a group
        if dSpltToken == pc.CloseDelim && isGrp {
            isGrp = false
            spaceProc = append(spaceProc, grpToken)
            continue
        }
        // if we arent processing a group, split spaces and add
        // each token, ignoring emptry strs when found
        if !isGrp {
            spaceSplitted := strings.Split(dSpltToken, " ") 
            for _, spSpltToken := range spaceSplitted {
                if spSpltToken == "" {
                    continue
                }
                spaceProc = append(spaceProc, spSpltToken)
            }
        } else {
            // if we are in a group, just add the whole token
            // to the group token
            grpToken += dSpltToken
        }

    }

    if isGrp {
        return nil, ErrNoGrp
    }

    return spaceProc, nil
}

/*
Name: parseFlags 
Type: Internal CLI func
Purpose: Parse logic for flags
*/
func (pc *ParseCtx) parseFlags(cmd Command, tokens []string) (string, error) {
    out := make(map[string][]string)
    currFlg := ""
    for _, token := range tokens {
        if len(token) > 2 && string(token[0:2]) == "--" {
            // check if longname exists
            didFnd := false
            for _, flag := range cmd.Flags {
                if flag.LongName != "" &&  flag.LongName == string(token[2:]) {
                    if out[flag.Name] != nil {
                        return "", ErrRpFlg
                    }
                    currFlg = flag.Name
                    out[currFlg] = make([]string, 0)
                    didFnd = true
                    break
                }
            }
            if didFnd {
                continue 
            }
        } else if len(token) > 1 && string(token[0]) == "-"{
            didFnd := false
            for _, flag := range cmd.Flags {
                if flag.Name == string(token[1:]) {
                    if out[flag.Name] != nil {
                        return "", ErrRpFlg
                    }
                    currFlg = flag.Name
                    out[currFlg] = make([]string, 0)
                    didFnd = true
                    break
                }
            }
            if didFnd {
                continue 
            }
        }
        if currFlg == "" {
            return "", ErrNoFlg
        }
        out[currFlg] = append(out[currFlg], token)
    }

    // perform validation
    err := pc.validation(cmd, out)

    if err != nil {
        return "", err
    }

    return cmd.Handler(out)
}


/*
Name: validation 
Type: Internal CLI func
Purpose: Parse logic for flags
*/
func (pc *ParseCtx) validation(cmd Command, in map[string][]string) (error){
    for _, flag := range cmd.Flags {
        if flag.ValidationCtx.Required && in[flag.Name] == nil {
            return ErrMissReq
        }
        if in[flag.Name] != nil {
            if flag.ValidationCtx.MaxArgs != InfiniteArgs && len(in[flag.Name]) > flag.ValidationCtx.MaxArgs {
                return ErrMulArg
            }
            if len(in[flag.Name]) < flag.ValidationCtx.MinArgs {
                return ErrNoArg
            }
        }
    }
    return nil
}

/*
Name: Parse 
Type: External CLI func
Purpose: Parse input str
*/
func (pc *ParseCtx) Parse(in string) (string, error) {
    tokens, err := pc.Tokenize(in)

    if err != nil {
        return "", err
    }

    if len(tokens) == 0 {
        return "", ErrNoCmd
    }

    for _, cmd := range pc.Commands {
        if cmd.Name == tokens[0] {
            return pc.parseFlags(cmd, tokens[1:])
        }
    }

    return "", ErrNoCmd
}

