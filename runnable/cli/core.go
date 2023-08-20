package cli

import (
    "strings"
    "fmt"
)

type Flag struct {
    FlagName    string
    Optional    bool
    Listable    bool
    Flags       []Flag
}

type Command struct {
    Name        string
    Flags       []Flag 
    Validator   func(in map[string][]string)(error)
    Handler     func(in map[string][]string)(string, error)
}

type CLIParser struct {
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

func (c *CLIParser) Tokenize (in string) []string {
    inOpenDelimProc := splitOn(in, c.OpenDelim)
    inCloseDelimProc := make([]string, 0,len(inOpenDelimProc))
    for _,v := range inOpenDelimProc {
        closeSplitted := splitOn(v, c.CloseDelim)
        inCloseDelimProc = append(inCloseDelimProc, closeSplitted...)
    }
    // Arbitrary capacity
    spaceProc := make([]string, 0, 3 * len(inOpenDelimProc))
    for _, outv := range inCloseDelimProc {
        spaceSplitted := strings.Split(outv, " ") 
        for _, inv := range spaceSplitted {
            if inv != "" {
                spaceProc = append(spaceProc, inv)
            }
        }
    }
    return spaceProc
}

func (c *CLIParser) Parse(in string) (string, error) {
    tokens := c.tokenize(in)
    fmt.Println(len(tokens))
    fmt.Println(tokens)
    return "", nil
}


