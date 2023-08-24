package cli

import (
    "strings"
    "errors"
)

const (
    InfiniteArgs = -1
)

var (
    ErrNoCmd = errors.New("command unrecognized")
    ErrNoFlg = errors.New("flag unrecognized")
    ErrRpFlg = errors.New("flag repeated")
    ErrNstGrp = errors.New("found nested group attempt")
    ErrNoGrp = errors.New("unclosed group")
)

type FlagValidationCtx struct{
    MaxArgs     int    
    MinArgs     int
    Required    bool
}

type Flag struct {
    Name            string
    LongName        string
    Description     string
    ValidationCtx   FlagValidationCtx
}

type Command struct {
    Name            string
    Description     string
    Flags           []Flag 
    Handler         func(in map[string][]string)(string, error)
}

type ParseCtx struct {
    Commands    []Command
    OpenDelim   string
    CloseDelim  string
}

// Same as split but leaves delim as its own "token". For example
// splitOn("a/b/c", "/") == []string{"a" "/" "b" "/" "c"}
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
        if dSpltToken == pc.OpenDelim && !isGrp {
            isGrp = true
            grpToken = ""
            continue
        }
        if dSpltToken == pc.CloseDelim && isGrp {
            isGrp = false
            spaceProc = append(spaceProc, grpToken)
            continue
        }
        if !isGrp {
            spaceSplitted := strings.Split(dSpltToken, " ") 
            for _, spSpltToken := range spaceSplitted {
                if spSpltToken != "" {
                    spaceProc = append(spaceProc, spSpltToken)
                }
            }
        } else {
            grpToken += dSpltToken
        }

    }

    if isGrp {
        return nil, ErrNoGrp
    }

    return spaceProc, nil
}

func (pc *ParseCtx) parseFlags(cmd Command, tokens []string) (string, error) {
    out := make(map[string][]string)
    currFlg := ""
    for _, token := range tokens {
        if len(token) > 2 && string(token[0:2]) == "--" {
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

//    err := pc.validation(out)
//
//    if err != nil {
//        return "", err
//    }

    return cmd.Handler(out)
}

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

