package main

import (
    "github.com/21Bruce/resolved-server/api/resy"
    "github.com/21Bruce/resolved-server/app"
    "github.com/21Bruce/resolved-server/runnable/cli"
    "os"
)

func main() {

    resy_api := &resy.API{APIKey: "VbWk7s3L4KiK5fzlO7JD3Q5EYolJI7n5"}
    appCtx := app.AppCtx{API: resy_api}
    cli := cli.ResolvedCLI{
        AppCtx: appCtx,
        In: os.Stdin,
        Out: os.Stdout,
        Err: os.Stderr,
    }
    cli.Run()
}
