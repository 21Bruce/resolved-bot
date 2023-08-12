package resy 

import (
    "github.com/21Bruce/resolved-server/api"
    "net/http"
    "net/url"
    "encoding/json"
    "io"
    "bytes"
    "strconv"
)

type API struct {
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
    request.Header.Set("Authorization", `ResyAPI api_key="VbWk7s3L4KiK5fzlO7JD3Q5EYolJI7n5"`)

    client := &http.Client{}
    response, err := client.Do(request)
    if err != nil {
        return nil, err
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
        Token:           jsonMap["token"].(string),
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
    request.Header.Set("Authorization", `ResyAPI api_key="VbWk7s3L4KiK5fzlO7JD3Q5EYolJI7n5"`)
    request.Header.Set("X-Resy-Auth-Token", params.Token)
    request.Header.Set("X-Resy-Universal-Auth-Token", params.Token)

    client := &http.Client{}
    response, err := client.Do(request)
    if err != nil {
        return nil, err
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
    numHits := int(jsonSearchMap["nbHits"].(float64))

    var limit int 
    if params.Limit > 0 {
        limit = min(params.Limit, numHits)
    } else {
        limit = numHits
    }
    searchResults := make([]api.SearchResult, limit, limit)

    jsonHitsMap := jsonSearchMap["hits"].([]interface{}) 
    for i:=0; i<limit; i++ {
        jsonHit := jsonHitsMap[i].(map[string]interface{})
        venueID, err := strconv.ParseInt(jsonHit["objectID"].(string), 10, 64)
        if err != nil {
            return nil, err
        }
        searchResults[i] = api.SearchResult{
            VenueID:      venueID,
            Name:         jsonHit["name"].(string), 
            Region:       jsonHit["region"].(string), 
            Locality:     jsonHit["locality"].(string), 
            Neighborhood: jsonHit["neighborhood"].(string), 
        }
    }

    searchResponse := api.SearchResponse{
        Results: searchResults,
    }

    return &searchResponse, nil
}
