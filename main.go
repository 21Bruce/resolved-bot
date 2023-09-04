/*
Author: Bruce Jagid
Created On: Aug 12, 2023
*/
package main

import (
    "github.com/21Bruce/resolved-server/api/resy"
    "github.com/21Bruce/resolved-server/app"
    "github.com/21Bruce/resolved-server/runnable/cli"
    "os"
)

func main() {
    resy_api := resy.GetDefaultAPI()
    appCtx := app.AppCtx{API: &resy_api}
    cli := cli.ResolvedCLI{
        AppCtx: appCtx,
        In: os.Stdin,
        Out: os.Stdout,
        Err: os.Stderr,
    }
    cli.Run()
}
