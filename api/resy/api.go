/*
Author: Bruce Jagid
Created On: Aug 12, 2023
*/
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

/*
Name: API
Type: API interface struct
Purpose: This struct acts as the resy implementation of the 
api interface. 
Note: The only known working APIKey value can be located and
defaulted using the GetDefaultAPI function, but we leave
it exposed so front-facing wrappers may expose it as a
setting
*/
type API struct {
    APIKey      string 
}

/*
Name: isCodeFail 
Type: Internal Func 
Purpose: Function which takes in an HTTP code and returns
true if it is not a success code and false otherwise
*/
func isCodeFail(code int) (bool) {
    fst := code / 100
    return (fst != 2)  
}

/*
Name: byteToJSONString 
Type: Internal Func 
Purpose: Function which takes in a byte sequence 
representing a JSON struct and returns a string 
or error. Useful for debugging
*/
func byteToJSONString(data []byte) (string, error) {
    var out bytes.Buffer
    err := json.Indent(&out, data, "", " ")

    if err != nil {
        return "", err
    }

    d := out.Bytes()
    return string(d), nil
}

/*
Name: min 
Type: Internal Func 
Purpose: Function that determins the min of two ints
*/
func min(a,b int) (int) {
    if a < b {
        return a
    }
    return b
}

/*
Name: GetDefaultAPI 
Type: External Func 
Purpose: Function that provides an out of the box
working API struct
*/
func GetDefaultAPI() (API){
    return API{
        APIKey: "VbWk7s3L4KiK5fzlO7JD3Q5EYolJI7n5",
    }
}

/*
Name: Login 
Type: API Func 
Purpose: Resy implementation of the Login api func
Note: The only required login fields for this func 
are Email and Password.
*/
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

    // Resy servers return a 419 is the auth parameters were invalid
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

    if jsonMap["payment_method_id"] == nil {
        return nil, api.ErrNoPayInfo
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

/*
Name: Search 
Type: API Func 
Purpose: Resy implementation of the Search api func
*/
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

    jsonHitsMap := jsonSearchMap["hits"].([]interface{}) 
    numHits := len(jsonHitsMap)

    // if input param limit is nonnegative, limit the search loop
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

/*
Name: Reserve
Type: API Func 
Purpose: Resy implementation of the Reserve api func
*/
func (a *API) Reserve(params api.ReserveParam) (*api.ReserveResponse, error) {
    
    // converting fields to url query format
    year := strconv.Itoa(params.ReservationTimes[0].Year())
    month := strconv.Itoa(int(params.ReservationTimes[0].Month()))
    day := strconv.Itoa(params.ReservationTimes[0].Day())
    date := year + "-" + month + "-" + day
    dayField := `day=` + date
    authField := `x-resy-auth-token=` + params.LoginResp.AuthToken
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
    request.Header.Set("X-Resy-Auth-Token", params.LoginResp.AuthToken)
    request.Header.Set("X-Resy-Universal-Auth-Token", params.LoginResp.AuthToken)
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


    // JSON structure is complicated here, see api/resy/doc.go for full explanation
    jsonResultsMap := jsonTopLevelMap["results"].(map[string]interface{}) 
    jsonVenuesList := jsonResultsMap["venues"].([]interface{})
    if len(jsonVenuesList) == 0 {
        return nil, api.ErrNoOffer
    }
    jsonVenueMap := jsonVenuesList[0].(map[string]interface{})
    jsonSlotsList := jsonVenueMap["slots"].([]interface{}) 
    for k := 0; k < len(params.TableTypes) || (len(params.TableTypes) == 0 && k == 0) ; k++ {
        // table type to search for, we decide this early on since its the least important thing
        
        currentTableType := api.DiningRoom
        if len(params.TableTypes) != 0 {
            currentTableType = params.TableTypes[k]
        }
        
        for i := 0; i < len(params.ReservationTimes); i++ {

            currentTime := params.ReservationTimes[i]
            for j:=0; j < len(jsonSlotsList); j++ {
                // if any errs appear, we just move to next time on list(i.e. continue)

                jsonSlotMap := jsonSlotsList[j].(map[string]interface{})
                jsonDateMap:= jsonSlotMap["date"].(map[string]interface{})

                // start contains the date for this slot in format "YrYrYrYr-MoMo-DyDy HrHr:MnMn"
                startRaw := jsonDateMap["start"].(string)
                // split to get ["YrYrYrYr-MoMo-DyDy", "HrHr:MnMn"]
                startFields := strings.Split(startRaw, " ")
                // isolate time field and split to get ["HrHr","MnMn"]
                timeFields := strings.Split(startFields[1], ":")
                // if time field matches of slot matches current selected ResTime, move to config step

                hourFieldInt, err := strconv.Atoi(timeFields[0])
                if err != nil {
                    return nil, err
                }

                minFieldInt, err := strconv.Atoi(timeFields[1])
                if err != nil {
                    return nil, err
                }

                jsonConfigMap := jsonSlotMap["config"].(map[string]interface{})
                // get table type of this slot for comparison
                tableType := strings.ToLower(jsonConfigMap["type"].(string))
                if hourFieldInt == currentTime.Hour() && minFieldInt == currentTime.Minute() && (len(params.TableTypes) == 0 || strings.Contains(string(currentTableType), tableType)){
                    configToken := jsonConfigMap["token"].(string)
                    configIDField := `config_id=` + url.QueryEscape(configToken)
                    // Reuse same fields from def of findUrl(see api/resy/doc.go)
                    fields = []string{dayField, partySizeField, authField, venueIDField, configIDField}

                    detailUrl := "https://api.resy.com/3/details?" + strings.Join(fields, "&") 

                    requestDetail, err := http.NewRequest("GET", detailUrl, bytes.NewBuffer([]byte{}))
                    if err != nil {
                        continue 
                    }

                    requestDetail.Header.Set("Authorization", `ResyAPI api_key="` + a.APIKey + `"`)
                    requestDetail.Header.Set("Host", `api.resy.com`)
                    requestDetail.Header.Set("X-Resy-Auth-Token", params.LoginResp.AuthToken)
                    requestDetail.Header.Set("X-Resy-Universal-Auth-Token", params.LoginResp.AuthToken)
                
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
                    jsonBookTokenMap := jsonTopLevelMap["book_token"].(map[string]interface{}) 
                    bookToken := jsonBookTokenMap["value"].(string)
 
                    // if config step yielded a book token, move to 'reserve' step
                    bookUrl := "https://api.resy.com/3/book?" + strings.Join(fields, "&") 

                    bookField := "book_token=" + url.QueryEscape(bookToken)
                    paymentMethodStr := `{"id":` + strconv.FormatInt(params.LoginResp.PaymentMethodID, 10) + `}`
                    paymentMethodField := "struct_payment_method=" + url.QueryEscape(paymentMethodStr)
                    requestBookBodyStr := bookField + "&" + paymentMethodField + "&" + "source_id=resy.com-venue-details"
                    requestBook, err := http.NewRequest("POST", bookUrl, bytes.NewBuffer([]byte(requestBookBodyStr)))
                    requestBook.Header.Set("Authorization", `ResyAPI api_key="` + a.APIKey + `"`)
                    requestBook.Header.Set("Content-Type", `application/x-www-form-urlencoded`)
                    requestBook.Header.Set("Host", `api.resy.com`)
                    requestBook.Header.Set("X-Resy-Auth-Token", params.LoginResp.AuthToken)
                    requestBook.Header.Set("X-Resy-Universal-Auth-Token", params.LoginResp.AuthToken)
                    requestBook.Header.Set("Referer", "https://resy.com/")
                    responseBook, err := client.Do(requestBook)
                    if err != nil {
                       continue 
                    }

                    if isCodeFail(responseBook.StatusCode) {
                        continue
                    }

                    responseBookBody, err := io.ReadAll(responseBook.Body)
                    if err != nil {
                        continue
                    }

                    err = json.Unmarshal(responseBookBody, &jsonTopLevelMap)
                    if err != nil {
                        continue
                    }

                    // if everything worked out, return time
                    resp := api.ReserveResponse{
                        ReservationTime: currentTime,
                    }

                    return &resp, nil

                }
            }
        }
    }
   
    // we only reach here if every time failed, meaning no table
    return nil, api.ErrNoTable 
}

/*
Name: AuthMinExpire 
Type: API Func 
Purpose: Resy implementation of the AuthMinExpire api func.
The largest minimum validity time is 6 days.
*/
func (a *API) AuthMinExpire() (time.Duration) {
    /* 6 days */
    var d time.Duration = time.Hour * 24 * 6
    return d
}

//func (a *API) Cancel(params api.CancelParam) (*api.CancelResponse, error) {
//    cancelUrl := `https://api.resy.com/3/cancel` 
//    resyToken := url.QueryEscape(params.ResyToken)
//    requestBodyStr := "resy_token=" + resyToken
//    request, err := http.NewRequest("POST", cancelUrl, bytes.NewBuffer([]byte(requestBodyStr)))
//    if err != nil {
//        return nil, err
//    }
//    
//    request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
//    request.Header.Set("Authorization", `ResyAPI api_key="` + a.APIKey + `"`)
//    request.Header.Set("X-Resy-Auth-Token", params.AuthToken)
//    request.Header.Set("X-Resy-Universal-Auth-Token", params.AuthToken)
//    request.Header.Set("Referer", "https://resy.com/")
//    request.Header.Set("Origin", "https://resy.com")
//
//
//    client := &http.Client{}
//    response, err := client.Do(request)
//    if err != nil {
//        return nil, err
//    }
//
//    if isCodeFail(response.StatusCode) {
//        return nil, api.ErrNetwork
//    }
//
//    responseBody, err := io.ReadAll(response.Body)
//    if err != nil {
//        return nil, err 
//    }
//
//    defer response.Body.Close()
//    var jsonTopLevelMap map[string]interface{}
//    err = json.Unmarshal(responseBody, &jsonTopLevelMap)
//    if err != nil {
//        return nil, err
//    }
//
//    jsonPaymentMap := jsonTopLevelMap["payment"].(map[string]interface{})
//    jsonTransactionMap := jsonPaymentMap["transaction"].(map[string]interface{})
//    refund := jsonTransactionMap["refund"].(int) == 1
//    return &api.CancelResponse{Refund: refund}, nil
//}
//
