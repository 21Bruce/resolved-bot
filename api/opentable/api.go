package opentable

import (
    "github.com/21Bruce/resolved-server/api"
    "bytes"
    "net/http"
    "io"
    "encoding/json"
    "strconv"
    "errors"
)

var (
    ErrTimeFalse = errors.New("time does not work")
    ErrBadData = errors.New("venue id does not work")
)

type API struct {
    XCSRFToken  string
    SearchKey   string
    FindKey     string
}

func GetDefaultAPI() (API) {
    return API{
        XCSRFToken: "2b167092-25e4-4f0d-a4a5-6f51e18d24e3",
        SearchKey: "3cabca79abcb0db395d3cbebb4d47d41f3ddd69442eba3a57f76b943cceb8cf4",
        FindKey: "e6b87021ed6e865a7778aa39d35d09864c1be29c683c707602dd3de43c854d86",
    }
}

func isCodeFail(code int) bool {
    fst := code / 100
    return (fst != 2)
}

// Since we don't need to login to opentable to reserve, just store 'login' 
// values in the LoginResp, which goes into the reserve function
func (a *API) Login(params api.LoginParam) (*api.LoginResponse, error) {
    return &api.LoginResponse{
        FirstName: params.FirstName,
        LastName: params.LastName,
        Email: params.Email,
        Mobile: params.Mobile,
    }, nil
}

func (a *API) getSlotMetadata(params api.ReserveParam, time api.Time) (*string, *string, error) {
    findUrl := "https://www.opentable.com/dapi/fe/gql?optype=query&opname=RestaurantsAvailability"
    dateStr := params.Year + "-" + params.Month + "-" + params.Day
    timeStr := time.Hour + ":" + time.Minute
    venueId := strconv.FormatInt(params.VenueID, 10)
    partySize := strconv.Itoa(params.PartySize)
    variableStr := `"variables": {"onlyPop": false, "forwardDays": 0,` +
    `"requireTimes": false, "requireTypes": [], "restaurantIds": [` + venueId + `],` +
    `"date":"` + dateStr + `", "time":"` + timeStr + `", "partySize":` + partySize +
    `,"databaseRegion": "NA"}`
    extensionStr := `"extensions": {"persistedQuery": {"version": 1, "sha256Hash": "` +
    a.FindKey + `"}}`
    bodyStr := `{"operationName": "RestaurantsAvailability", ` + variableStr + `, `+ extensionStr + `}`
    bodyBytes := []byte(bodyStr)
    request, err := http.NewRequest("POST", findUrl, bytes.NewBuffer(bodyBytes))

    if err != nil {
        return nil, nil, err
    }
    
    request.Header.Set("Content-Type", "application/json")
    request.Header.Set("Host", "www.opentable.com")
    request.Header.Set("x-csrf-token", a.XCSRFToken)
    request.Header.Set("accept", "*/*")
    request.Header.Set("Connection", "keep-alive")
    request.Header.Set("Origin", "https://www.opentable.com")
    request.Header.Set("user-agent", "Resolved-Server")

    client := &http.Client{}

    response, err := client.Do(request)

    if err != nil {
        return nil, nil, err
    }

    if isCodeFail(response.StatusCode) {
        return nil, nil, api.ErrNetwork
    }

    defer response.Body.Close()

    responseBody, err := io.ReadAll(response.Body)
    if err != nil {
        return nil, nil, err
    }

    var jsonTopLevelMap map[string]interface{}
    err = json.Unmarshal(responseBody, &jsonTopLevelMap)
    if err != nil {
        return nil, nil, err
    }

    jsonDataMap := jsonTopLevelMap["data"].(map[string]interface{})
    if jsonDataMap["availability"] == nil {
        return nil, nil, ErrBadData
    }

    jsonAvailabilityMap := jsonDataMap["availability"].([]interface{})[0].(map[string]interface{})
    jsonAvailabilityDaysMap := jsonAvailabilityMap["availabilityDays"].([]interface{})[0].(map[string]interface{})
    jsonHitsMap := jsonAvailabilityDaysMap["slots"].([]interface{})
    for i := 0; i < len(jsonHitsMap); i++ {
        jsonHitMap := jsonHitsMap[i].(map[string]interface{})
        if jsonHitMap["isAvailable"].(bool) != true {
            continue
        }
        if int(jsonHitMap["timeOffsetMinutes"].(float64)) != 0 {
            continue
        }
        slotHash := jsonHitMap["slotHash"].(string)
        slotToken := jsonHitMap["slotAvailabilityToken"].(string)
        return &slotHash, &slotToken, nil

    }
    return nil, nil, api.ErrNoTable 

}

func (a *API) finalizeReservation(hash string, token string, time api.Time, params api.ReserveParam) (*api.ReserveResponse, error) {
    resUrl := "https://www.opentable.com/dapi/booking/make-reservation"
    dateStr := params.Year + "-" + params.Month + "-" + params.Day
    timeStr := time.Hour + ":" + time.Minute
    dateTimeStr := dateStr + "T" + timeStr
    venueId := strconv.FormatInt(params.VenueID, 10)
    partySize := strconv.Itoa(params.PartySize)
    bodyStr := `{"restaurantId": ` + venueId + `,` +
    `"slotAvailabilityToken": "` + token + `",` +
    `"slotHash": "` + hash + `",` +
    `"isModify": false,` + 
    `"reservationDateTime": "` + dateTimeStr + `",` + 
    `"partySize": ` +  partySize + `,` + 
    `"firstName": "` +  params.LoginResp.FirstName + `",` + 
    `"lastName": "` +  params.LoginResp.LastName + `",` + 
    `"email": "` +  params.LoginResp.Email + `",` + 
    `"phoneNumber": "` +  params.LoginResp.Mobile + `",` + 
    `"phoneNumberCountryId": "US",` + 
    `"country": "US",` + 
    `"reservationType": "Standard",` + 
    `"reservationAttribute": "highTop",` + 
    `"pointsType": "Standard",` + 
    `"diningAreaId": 1,` + 
    `"optInEmailRestaurant": false}`  
    bodyBytes := []byte(bodyStr)
    request, err := http.NewRequest("POST", resUrl, bytes.NewBuffer(bodyBytes))

    if err != nil {
        return nil, err
    }
    
    request.Header.Set("Content-Type", "application/json")
    request.Header.Set("Host", "www.opentable.com")
    request.Header.Set("x-csrf-token", a.XCSRFToken)
    request.Header.Set("accept", "*/*")
    request.Header.Set("Connection", "keep-alive")
    request.Header.Set("Origin", "https://www.opentable.com")
    request.Header.Set("user-agent", "Resolved-Server")

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

    if jsonTopLevelMap["success"].(bool) {
        return &api.ReserveResponse{
            ReservationTime: time,
        }, nil
    }

    return nil, api.ErrNoTable


}

func (a *API) Reserve(params api.ReserveParam) (*api.ReserveResponse, error) {
    for _, v := range params.ReservationTimes {
        hash, token, err := a.getSlotMetadata(params, v)
        if err != nil {
            continue
        } 
        res, err := a.finalizeReservation(*hash, *token, v, params)
        if err != nil {
            continue
        }
        return res, nil
    }
    return nil, api.ErrNoTable 
}

func (a *API) Search(params api.SearchParam) (*api.SearchResponse, error) {
    searchUrl := "https://www.opentable.com/dapi/fe/gql?optype=query&opname=Autocomplete"

    variableStr := `"variables": {"term": "` + params.Name +`", "latitude": 1, "longitude": 1, "useNewVersion": true}`
    extensionStr := `"extensions": {"persistedQuery": {"version": 1, "sha256Hash":"` + a.SearchKey + `"}}`
    bodyStr :=`{"operationName": "Autocomplete",` + variableStr + `,` + extensionStr + `}`
    bodyBytes := []byte(bodyStr)

    request, err := http.NewRequest("POST", searchUrl, bytes.NewBuffer(bodyBytes))

    if err != nil {
        return nil, err
    }
    
    request.Header.Set("Content-Type", "application/json")
    request.Header.Set("Host", "www.opentable.com")
    request.Header.Set("x-csrf-token", a.XCSRFToken)
    request.Header.Set("accept", "*/*")
    request.Header.Set("Connection", "keep-alive")
    request.Header.Set("Origin", "https://www.opentable.com")
    request.Header.Set("user-agent", "Resolved-Server")

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

    jsonDataMap := jsonTopLevelMap["data"].(map[string]interface{})
    jsonAutocompleteMap := jsonDataMap["autocomplete"].(map[string]interface{})
    jsonResultsMap := jsonAutocompleteMap["autocompleteResults"].([]interface{})
    numPossibleHits := len(jsonResultsMap)
    var limit int
    if params.Limit > 0 {
        limit = params.Limit
    } else {
        limit = numPossibleHits
    }
    
    searchResults := make([]api.SearchResult, 0, limit)
    for i:=0; i<numPossibleHits; i++ {
        jsonHitMap := jsonResultsMap[i].(map[string]interface{})
        if len(searchResults) == limit {
            break
        }
        if jsonHitMap["type"].(string) != "Restaurant" {
            continue
        }
        venueID, err := strconv.ParseInt(jsonHitMap["id"].(string), 10, 64)
        if err != nil {
            return nil, err
        }
        searchResults = append(searchResults, api.SearchResult{
            VenueID:      venueID,
            Name:         jsonHitMap["name"].(string), 
            Region:       jsonHitMap["country"].(string), 
            Locality:     jsonHitMap["metroName"].(string), 
            Neighborhood: jsonHitMap["neighborhoodName"].(string), 
        })
    }

    searchResponse := api.SearchResponse{
        Results: searchResults,
    }

    return &searchResponse, nil
}
