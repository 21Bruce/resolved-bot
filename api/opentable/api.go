package opentable

import (
    "github.com/21Bruce/resolved-server/api"
    "bytes"
    "net/http"
    "io"
    "encoding/json"
    "strconv"
)

type API struct {
    XCSRFToken  string
    SearchKey   string
}

func GetDefaultAPI() (API) {
    return API{
        XCSRFToken: "2b167092-25e4-4f0d-a4a5-6f51e18d24e3",
        SearchKey: "3cabca79abcb0db395d3cbebb4d47d41f3ddd69442eba3a57f76b943cceb8cf4",
    }
}

func isCodeFail(code int) bool {
    fst := code / 100
    return (fst != 2)
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
    request.Header.Set("x-query-timeout", "1500")
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
