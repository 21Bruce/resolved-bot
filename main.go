package main

import (
    "github.com/21Bruce/resolved-server/runnable/cli"
//    "github.com/21Bruce/resolved-server/api"
    "github.com/21Bruce/resolved-server/api/resy"
//    "github.com/21Bruce/resolved-server/app/default"
//    "github.com/21Bruce/resolved-server/app"
//    "time"
//    "strconv"
    "os"
)

func main() {

    resy_api := &resy.API{APIKey: "VbWk7s3L4KiK5fzlO7JD3Q5EYolJI7n5"}
    cliWrapper := cli.ResolvedCLI{API: resy_api, In: os.Stdin, Out: os.Stdout, Err: os.Stderr}
    cliWrapper.Run()

//    go_app := default.App{}
//    resp, err := go_app.ReserveAtInterval(
//        app.ReserveAtIntervalParam{
//            Email: "brucejagid@gmail.com",
//            Password: "1Erf236ab1@",
//            API: resy_api,
//            VenueID: 66878,
//            Day: "13", 
//            Month: "08",
//            Year: "2023",
//            ReservationTimes: []api.Time{api.Time{Hour:"12", Minute:"30"}},
//            PartySize: 2,
//            RepeatInterval: api.Time{Hour:"00", Minute:"01"},
//        })
//    date := time.Date(2023, time.Month(8), 13, 12, 11, 0, 0, time.Local)
//    date = date.UTC()
//    hour := strconv.Itoa(date.Hour())
//    minute := strconv.Itoa(date.Minute())
//    year, month, day := date.Date()
//    dayStr := strconv.Itoa(day)
//    monthStr := strconv.Itoa(int(month))
//    yearStr := strconv.Itoa(year)
//    resp, err := go_app.ReserveAtTime(
//        app.ReserveAtTimeParam{
//            Email: "brucejagid@gmail.com",
//            Password: "1Erf236ab1@",
//            API: resy_api,
//            VenueID: 66878,
//            Day: "14", 
//            Month: "08",
//            Year: "2023",
//            ReservationTimes: []api.Time{api.Time{Hour:"14", Minute:"45"}},
//            PartySize: 2,
//            RequestDay: dayStr, 
//            RequestMonth: monthStr,
//            RequestYear: yearStr,
//            RequestTime: api.Time{Hour: hour, Minute: minute},
//        })
//    fmt.Println(resp)
//    fmt.Println(err)

    
}
