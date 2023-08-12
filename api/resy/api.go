package resy 

import (
    "github.com/21Bruce/resolved-server/api"
    "net/http"
    "net/url"
    "encoding/json"
    "io"
    "bytes"
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

func (a *API)  Login(params api.LoginParam) (*api.LoginResponse, error) {

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


