package resy 

import (
    "github.com/21Bruce/resolved-server/api"
    "net/http"
    "net/url"
    "encoding/json"
    "io"
    "bytes"
    "strconv"
    "strings"
    "time"
)

type API struct {
    APIKey      string 
}

func isCodeFail(code int) (bool) {
    fst := code / 100
    return (fst != 2)  
}

func byteToJSONString(data []byte) (string, error) {
    var out bytes.Buffer
    err := json.Indent(&out, data, "", " ")

    if err != nil {
        return "", err
    }

    d := out.Bytes()
    return string(d), nil
}

func min(a,b int) (int) {
    if a < b {
        return a
    }
    return b
}

func (a *API) Login(params api.LoginParam) (*api.LoginResponse, error) {
    authUrl := "https://api.resy.com/3/auth/password"
    email := url.QueryEscape(params.Email)
    password := url.QueryEscape(params.Password)
    bodyStr :=`email=` + email + `&password=` + password
    bodyBytes := []byte(bodyStr)

    request, err := http.NewRequest("POST", authUrl, bytes.NewBuffer(bodyBytes))
    if err != nil {
        return nil, err
    }
    
    request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    request.Header.Set("Authorization", `ResyAPI api_key="` + a.APIKey + `"`)

    client := &http.Client{}
    response, err := client.Do(request)

    if err != nil {
        return nil, err
    }

    if response.StatusCode == 419 {
        return nil, api.ErrLoginWrong
    }

    if isCodeFail(response.StatusCode) {
        return nil, api.ErrNetwork
    }

    defer response.Body.Close()

    responseBody, err := io.ReadAll(response.Body)
    if err != nil {
        return nil, err
    }

    var jsonMap map[string]interface{}
    err = json.Unmarshal(responseBody, &jsonMap)
    if err != nil {
        return nil, err
    }

    loginResponse := api.LoginResponse{
        ID:              int64(jsonMap["id"].(float64)),
        FirstName:       jsonMap["first_name"].(string),
        LastName:        jsonMap["last_name"].(string),
        Mobile:          jsonMap["mobile_number"].(string),
        Email:           jsonMap["em_address"].(string),
        PaymentMethodID: int64(jsonMap["payment_method_id"].(float64)),
        AuthToken:       jsonMap["token"].(string),
    }

    return &loginResponse, nil

}

func (a *API) Search(params api.SearchParam) (*api.SearchResponse, error) {
    searchUrl := "https://api.resy.com/3/venuesearch/search"

    bodyStr :=`{"query":"` + params.Name +`"}`
    bodyBytes := []byte(bodyStr)

    request, err := http.NewRequest("POST", searchUrl, bytes.NewBuffer(bodyBytes))
    if err != nil {
        return nil, err
    }
    
    request.Header.Set("Content-Type", "application/json")
    request.Header.Set("Authorization", `ResyAPI api_key="` + a.APIKey + `"`)
    request.Header.Set("Origin", `https://resy.com`)
    request.Header.Set("Referer", `https://resy.com/`)

    client := &http.Client{}
    response, err := client.Do(request)

    if err != nil {
        return nil, err
    }

    if isCodeFail(response.StatusCode) {
        return nil, api.ErrNetwork
    }

    defer response.Body.Close()

    responseBody, err := io.ReadAll(response.Body)
    if err != nil {
        return nil, err
    }
    
    var jsonTopLevelMap map[string]interface{}
    err = json.Unmarshal(responseBody, &jsonTopLevelMap)
    if err != nil {
        return nil, err
    }

    jsonSearchMap := jsonTopLevelMap["search"].(map[string]interface{})
    //numHits := int(jsonSearchMap["nbHits"].(float64))

    jsonHitsMap := jsonSearchMap["hits"].([]interface{}) 
    numHits := len(jsonHitsMap)
    var limit int 
    if params.Limit > 0 {
        limit = min(params.Limit, numHits)
    } else {
        limit = numHits
    }
    searchResults := make([]api.SearchResult, limit, limit)
    for i:=0; i<limit; i++ {
        jsonHitMap := jsonHitsMap[i].(map[string]interface{})
        venueID, err := strconv.ParseInt(jsonHitMap["objectID"].(string), 10, 64)
        if err != nil {
            return nil, err
        }
        searchResults[i] = api.SearchResult{
            VenueID:      venueID,
            Name:         jsonHitMap["name"].(string), 
            Region:       jsonHitMap["region"].(string), 
            Locality:     jsonHitMap["locality"].(string), 
            Neighborhood: jsonHitMap["neighborhood"].(string), 
        }
    }

    searchResponse := api.SearchResponse{
        Results: searchResults,
    }

    return &searchResponse, nil
}

func (a *API) Reserve(params api.ReserveParam) (*api.ReserveResponse, error) {
    
    date := params.Year + "-" + params.Month + "-" + params.Day
    dayField := `day=` + date
    authField := `x-resy-auth-token=` + params.AuthToken
    latField := `lat=0`
    longField := `long=0`
    venueIDField := `venue_id=` + strconv.FormatInt(params.VenueID, 10)
    partySizeField := `party_size=` + strconv.Itoa(params.PartySize)
    fields := []string{dayField, authField, latField, longField, venueIDField, partySizeField}

    findUrl := `https://api.resy.com/4/find?` + strings.Join(fields, "&")
    

    request, err := http.NewRequest("GET", findUrl, bytes.NewBuffer([]byte{}))
    if err != nil {
        return nil, err
    }
    
    request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    request.Header.Set("Authorization", `ResyAPI api_key="` + a.APIKey + `"`)
    request.Header.Set("X-Resy-Auth-Token", params.AuthToken)
    request.Header.Set("X-Resy-Universal-Auth-Token", params.AuthToken)
    request.Header.Set("Referer", "https://resy.com/")


    client := &http.Client{}
    response, err := client.Do(request)
    if err != nil {
        return nil, err
    }

    if isCodeFail(response.StatusCode) {
        return nil, api.ErrNetwork
    }

    defer response.Body.Close()

    responseBody, err := io.ReadAll(response.Body)
    if err != nil {
        return nil, err
    }

    var jsonTopLevelMap map[string]interface{}
    err = json.Unmarshal(responseBody, &jsonTopLevelMap)
    if err != nil {
        return nil, err
    }


    jsonResultsMap := jsonTopLevelMap["results"].(map[string]interface{}) 
    jsonVenuesList := jsonResultsMap["venues"].([]interface{}) 
    jsonVenueMap := jsonVenuesList[0].(map[string]interface{})
    jsonSlotsList := jsonVenueMap["slots"].([]interface{}) 
    for i:=0; i < len(params.ReservationTimes); i++ {
        currentTime := params.ReservationTimes[i]
        for j:=0; j < len(jsonSlotsList); j++ {
            jsonSlotMap := jsonSlotsList[j].(map[string]interface{})
            jsonDateMap:= jsonSlotMap["date"].(map[string]interface{})
            startRaw := jsonDateMap["start"].(string)
            startFields := strings.Split(startRaw, " ")
            timeFields := strings.Split(startFields[1], ":")
            if timeFields[0] == currentTime.Hour && timeFields[1] == currentTime.Minute {
                jsonConfigMap := jsonSlotMap["config"].(map[string]interface{})
                configToken := jsonConfigMap["token"].(string)
                configIDField := `config_id=` + url.QueryEscape(configToken)
                fields = []string{dayField, partySizeField, authField, venueIDField, configIDField}
                detailUrl := "https://api.resy.com/3/details?" + strings.Join(fields, "&") 
                requestDetail, err := http.NewRequest("GET", detailUrl, bytes.NewBuffer([]byte{}))
                if err != nil {
                    continue 
                }

                requestDetail.Header.Set("Authorization", `ResyAPI api_key="` + a.APIKey + `"`)
                requestDetail.Header.Set("Host", `api.resy.com`)
                requestDetail.Header.Set("X-Resy-Auth-Token", params.AuthToken)
                requestDetail.Header.Set("X-Resy-Universal-Auth-Token", params.AuthToken)
            
                responseDetail, err := client.Do(requestDetail)
                if err != nil {
                    continue
                }

                if isCodeFail(responseDetail.StatusCode) {
                    return nil, api.ErrNetwork
                }


                defer responseDetail.Body.Close()

                responseDetailBody, err := io.ReadAll(responseDetail.Body)
                if err != nil {
                    continue
                }

                var jsonTopLevelMap map[string]interface{}
                err = json.Unmarshal(responseDetailBody, &jsonTopLevelMap)
                if err != nil {
                    return nil, err
                }

                bookUrl := "https://api.resy.com/3/book?" + strings.Join(fields, "&") 


                jsonBookTokenMap := jsonTopLevelMap["book_token"].(map[string]interface{}) 
                bookToken := jsonBookTokenMap["value"].(string)
                bookField := "book_token=" + url.QueryEscape(bookToken)
                paymentMethodStr := `{"id":` + strconv.FormatInt(params.PaymentMethodID, 10) + `}`
                paymentMethodField := "struct_payment_method=" + url.QueryEscape(paymentMethodStr)
                requestBookBodyStr := bookField + "&" + paymentMethodField + "&" + "source_id=resy.com-venue-details"
                requestBook, err := http.NewRequest("POST", bookUrl, bytes.NewBuffer([]byte(requestBookBodyStr)))
                requestBook.Header.Set("Authorization", `ResyAPI api_key="` + a.APIKey + `"`)
                requestBook.Header.Set("Content-Type", `application/x-www-form-urlencoded`)
                requestBook.Header.Set("Host", `api.resy.com`)
                requestBook.Header.Set("X-Resy-Auth-Token", params.AuthToken)
                requestBook.Header.Set("X-Resy-Universal-Auth-Token", params.AuthToken)
                requestBook.Header.Set("Referer", "https://resy.com/")
                responseBook, err := client.Do(requestBook)
                if err != nil {
                   continue 
                }

                if isCodeFail(responseBook.StatusCode) {
                    return nil, api.ErrNetwork
                }

                responseBookBody, err := io.ReadAll(responseBook.Body)
                if err != nil {
                    continue
                }

                err = json.Unmarshal(responseBookBody, &jsonTopLevelMap)
                if err != nil {
                    continue
                }

                resp := api.ReserveResponse{
                    ReservationTime: currentTime,
                    ResyToken: jsonTopLevelMap["resy_token"].(string),
                }

                return &resp, nil

            }
        }
         
    }
    
    return nil, api.ErrNoTable 
}

func (a *API) Cancel(params api.CancelParam) (*api.CancelResponse, error) {
    cancelUrl := `https://api.resy.com/3/cancel` 
    resyToken := url.QueryEscape(params.ResyToken)
    requestBodyStr := "resy_token=" + resyToken
    request, err := http.NewRequest("POST", cancelUrl, bytes.NewBuffer([]byte(requestBodyStr)))
    if err != nil {
        return nil, err
    }
    
    request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    request.Header.Set("Authorization", `ResyAPI api_key="` + a.APIKey + `"`)
    request.Header.Set("X-Resy-Auth-Token", params.AuthToken)
    request.Header.Set("X-Resy-Universal-Auth-Token", params.AuthToken)
    request.Header.Set("Referer", "https://resy.com/")
    request.Header.Set("Origin", "https://resy.com")


    client := &http.Client{}
    response, err := client.Do(request)
    if err != nil {
        return nil, err
    }

    if isCodeFail(response.StatusCode) {
        return nil, api.ErrNetwork
    }

    responseBody, err := io.ReadAll(response.Body)
    if err != nil {
        return nil, err 
    }

    defer response.Body.Close()
    var jsonTopLevelMap map[string]interface{}
    err = json.Unmarshal(responseBody, &jsonTopLevelMap)
    if err != nil {
        return nil, err
    }

    jsonPaymentMap := jsonTopLevelMap["payment"].(map[string]interface{})
    jsonTransactionMap := jsonPaymentMap["transaction"].(map[string]interface{})
    refund := jsonTransactionMap["refund"].(int) == 1
    return &api.CancelResponse{Refund: refund}, nil
}

func findLastTime(times []api.Time) (*api.Time, error) {
    if len(times) == 0 {
        return nil, api.ErrTimeNull
    }
    lastTime := times[0]
    for i := 1; i < len(times); i++{
        lstHr, err := strconv.ParseInt(lastTime.Hour, 10, 64)
        if err != nil {
            return nil, err
        }
        lstMn, err := strconv.ParseInt(lastTime.Minute, 10, 64)
        if err != nil {
            return nil, err
        }
        thsHr, err := strconv.ParseInt(times[i].Hour, 10, 64)
        if err != nil {
            return nil, err
        }
        thsMn, err := strconv.ParseInt(times[i].Minute, 10, 64)
        if err != nil {
            return nil, err
        }
        if (lstHr < thsHr) || (lstHr == thsHr && lstMn < thsMn) {
           lastTime = times[i] 
        } 
    }
    return &lastTime, nil
}

func dateStringsToInts(in []string) ([]int, error) {
    out := make([]int, len(in), len(in))
    for i, s := range in {
        raw, err := strconv.ParseInt(s, 10, 64)
        if err != nil {
            return nil, err 
        }
        out[i] = int(raw)
    }
    return out, nil
}

func isLastTimeFuture(year, month, day, hour, minute int) bool{
    now := time.Now()
    nowYear, nowMonth, nowDay := now.Date()
    yrCmp := nowYear < year
    yrEq := nowYear == year
    mtCmp := int(nowMonth) < month
    mtEq := int(nowMonth) == month
    dyCmp := nowDay < day
    dyEq := nowDay  == day
    hrCmp := now.Hour() < hour
    hrEq := now.Hour() == hour
    mnCmp := now.Minute() < minute
    cmp := yrCmp || 
    (yrEq && mtCmp) || 
    (yrEq && mtEq && dyCmp) || 
    (yrEq && mtEq && dyEq && hrCmp) ||
    (yrEq && mtEq && dyEq && hrEq && mnCmp) 
    return cmp     
}

func (a *API) ReserveAtInterval(params api.ReserveAtIntervalParam) (*api.ReserveAtIntervalResponse, error){
    lastTime, err := findLastTime(params.ReservationTimes)
    if err != nil {
        return nil, err
    }
    dateInts, err := dateStringsToInts([]string{ 
        params.RepeatInterval.Hour,
        params.RepeatInterval.Minute,
        params.Year,
        params.Month,
        params.Day,
        lastTime.Hour,
        lastTime.Minute,
    })
    if err != nil {
        return nil, err
    }
    numHrs := dateInts[0]
    numMns := dateInts[1] 
    year := dateInts[2]
    month := dateInts[3] 
    day := dateInts[4]
    hour := dateInts[5] 
    minute := dateInts[6] 

    repeatInterval := time.Hour * time.Duration(numHrs) + time.Minute * time.Duration(numMns)
    for {
        loginResp, err := a.Login(
            api.LoginParam{
                Email: params.Email,
                Password: params.Password,
            })
        
        if err != nil {
            return nil, err
        }
        reserveResp, err := a.Reserve(
            api.ReserveParam{
                AuthToken: loginResp.AuthToken,
                Day: params.Day,
                Month: params.Month,
                Year: params.Year,
                ReservationTimes: params.ReservationTimes,
                PartySize: params.PartySize,
                PaymentMethodID: loginResp.PaymentMethodID,
                VenueID: params.VenueID,
            })
        if err != nil && err != api.ErrNoTable {
            return nil, err
        }
        if err == api.ErrNoTable {
           cmp := isLastTimeFuture(year, month, day, hour, minute)
           if cmp {
                time.Sleep(repeatInterval)
                continue
            }
            return nil, api.ErrPastDate
        }
        return &api.ReserveAtIntervalResponse{ReservationTime: reserveResp.ReservationTime}, nil

    }
}

func (a *API) ReserveAtTime(params api.ReserveAtTimeParam) (*api.ReserveAtTimeResponse, error){

    dateInts, err := dateStringsToInts([]string{ 
        params.RequestTime.Hour,
        params.RequestTime.Minute,
        params.RequestYear,
        params.RequestMonth,
        params.RequestDay,
    })

    if err != nil {
        return nil, err
    }
    hour := dateInts[0]
    minute := dateInts[1] 
    year := dateInts[2]
    month := dateInts[3] 
    day := dateInts[4]
    requestTime :=  time.Date(year, time.Month(month), day, hour, minute, 0, 0, time.UTC)
    time.Sleep(time.Until(requestTime))

    loginResp, err := a.Login(
        api.LoginParam{
            Email: params.Email,
            Password: params.Password,
        })
    
    if err != nil {
        return nil, err
    }
    reserveResp, err := a.Reserve(
        api.ReserveParam{
            AuthToken: loginResp.AuthToken,
            Day: params.Day,
            Month: params.Month,
            Year: params.Year,
            ReservationTimes: params.ReservationTimes,
            PartySize: params.PartySize,
            PaymentMethodID: loginResp.PaymentMethodID,
            VenueID: params.VenueID,
        })
    if err != nil {
        return nil, err
    }
    
    returnValue := api.ReserveAtTimeResponse{ ReservationTime: reserveResp.ReservationTime }
    return &returnValue, nil
    
}

