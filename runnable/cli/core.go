package cli

import (
    "strings"
    "errors"
)

var (
    ErrNoCm = errors.New("command unrecognized")
    ErrNoFlg = errors.New("flag unrecognized")
    ErrRpFlg = errors.New("flag repeated")
    ErrNstGrp = errors.New("found nested group attempt")
    ErrNoGrp = errors.New("unclosed group")
)

type Flag struct {
    Name    string
}

type Command struct {
    Name        string
    Flags       []Flag 
    Handler     func(in map[string][]string)(string, error)
}

type CLIWrapper struct {
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

func (c *CLIWrapper) Tokenize (in string) ([]string, error) {
    inOpenDelimProc := splitOn(in, c.OpenDelim)
    inCloseDelimProc := make([]string, 0,len(inOpenDelimProc))
    isGrp := false
    grpToken := ""
    for _,v := range inOpenDelimProc {
        closeSplitted := splitOn(v, c.CloseDelim)
        inCloseDelimProc = append(inCloseDelimProc, closeSplitted...)
    }
    // Arbitrary capacity
    spaceProc := make([]string, 0, 3 * len(inOpenDelimProc))
    for _, dSpltToken := range inCloseDelimProc {
        if dSpltToken == c.OpenDelim && !isGrp {
            isGrp = true
            grpToken = ""
            continue
        }
        if dSpltToken == c.CloseDelim && isGrp {
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

func (c *CLIWrapper) parseFlags(cmd Command, tokens []string) (string, error) {
    out := make(map[string][]string)
    currFlg := ""
    for _, token := range tokens {
        if len(token) > 1 && string(token[0]) == "-"{
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

    return cmd.Handler(out)
}

func (c *CLIWrapper) Parse(in string) (string, error) {
    tokens, err := c.Tokenize(in)

    if err != nil {
        return "", err
    }

    if len(tokens) == 0 {
        return "", ErrNoCm
    }

    for _, cmd := range c.Commands {
        if cmd.Name == tokens[0] {
            return c.parseFlags(cmd, tokens[1:])
        }
    }

    return "", ErrNoCm
}


