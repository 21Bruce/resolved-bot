/*
**********************************************************************

General Purpose: 

    The resy pkg is an implementation of the 'API' interface using resy
    as an external reservation service. This documentation dives into
    the specifics on how we implement that api spec with Resy.

**********************************************************************

General Overview of Resy API:

    Resy, being a closed-source API, is still a black-box to us and 
    the information provided here is only an inference based on 
    studying the message structure as observed on the Resy web UI.

    The Resy REST API communicates via url-encoded form data and JSON
    over HTTP. Each HTTP Request must have an authorization header with 
    a Resy API key, which is a string of characters that Resy recognizes.
    The only known key to this project is provided in the API struct 
    returned from resy.GetDefaultAPI(), but we leave this field of the
    API struct public in case other apps want to expose this as a
    setting. The specific format is discussed in the 'APIKey' section 
    of this doc.go.

    The Login functionality of Resy requires only one request message,
    and generally takes an account email and password as input. On 
    success, we are given a token to use in future requests, along
    with account data such as a name and phone numbers. The specific
    message structure of a Login operation is discussed in the 'Login'
    section of this document.

    The Search functionality of Resy requires only one request message,
    and generally takes a query name  as input. On success, we are
    given a JSON of restaurant data entries matching the query  in
    some way. The specific message structure of a Search operation is
    discussed in the 'Search' section of this document.

    The Reservation functionality of Resy requires three
    different types of request messages, each requiring different
    inputs. One message is used to get the available slots at a
    restaurant along with metadata on each slot, another is used
    to signal to the server that a reservation on that slot is in 
    progress which returns a token used to identify the ongoing
    reservation attempt in future requests. The last message sends
    the book token and attempts to finalize the reservation. 
    The specific message structure of a Reserve operation is
    discussed in the 'Reserve' section of this document.
    
**********************************************************************

APIKey: 

    The format of the header is provided here with the
    token ###KEY### representing where a key should go:

        Authorization: ResyAPI api_key="###KEY###"

**********************************************************************

Login: 
        
    The Login function of the Resy REST API is relatively simple.
    The request encapsulates data in the form of url-
    encoded form data. It is a POST. The URL for authentication is:

        https://api.resy.com/3/auth/password

    Then, the user email and password is inserted into the body in
    the following manner. We will use ###UEEMAIL### to represent a 
    url-encoded email and ###UEPSW### to represent a url-encoded 
    password:

        Body:

            email=###UEEMAIL###&password=###UEPSW###

    The HTTP response to this data is a large complex of Resy metadata 
    and relevant information packaged in a JSON structure. In the
    following, we use ###UID### to represent the user id, ###FST### 
    to represent the first name of the user, ###lst### to represent 
    the last name of the user, ###MOB### to represent the mobile phone
    number of the user, ###EM### to represent the email of the user,
    ###TOK### to represent the authorization token,
    and ###PID### to represent the payment method id:

        Response Body:
            
            {
                ...
                "id": "###UID###",
                "first_name": "###FST###",
                "last_name": "###LST###",
                "mobile_number": "###MOB###",
                "em_address": "###EM###",
                "payment_method_id": "###PID###",
                "token": "###TOK###",
                ...
            }

    The relevant info for future requests are the ###TOK###  and
    ###PID### values. The ###TOK### value must be provided in any  
    user specific HTTP request, as the value for two header fields:

        Headers:

            X-Resy-Auth-Token: ###TOK###
            X-Resy-Universal-Auth-Token: ###TOK###

    The ###PID### value is used in the reservation process.

**********************************************************************

Search: 

    The Search function of the Resy REST API is relatively simple.
    The request, which is a POST, encapsulates data in the form of 
    JSON. The URL for Searching is:

        https://api.resy.com/3/venuesearch/search

    Then, the input parameter query parameter, denoted here by the
    token ###NAME### is inserted into the body like so:

        Body:
            
            {
                "query": "###NAME###"
            }

    We also include in this request's headers the follwoing:

        Headers:

            Origin: https://resy.com        
            Referer: https://resy.com/ 

    Which are standard HTTP header fields. The body of the HTTP 
    response to this information is again a confused mass of 
    Metadata and relevant facts, so we only specify the structure
    and relevant fields of the response. It comes in a JSON structure.
    We use the tokens ###NAME### to mean the name of a restaurant, the tokens
    ###REG###, ###LOC###, ###NEIG### to be the region, locality, and
    neighborhood, respectfully, that a restaurant is situated in and 
    ###ID### to be the VenueID of the restuarant:

        Body: 

            {
                ...
                "search": 
                    {
                        ...
                        "hits":
                            [
                                ...
                                {
                                    ...
                                    "objectID": "###ID###",
                                    "name": "###NAME###",
                                    "region": "###REG###",
                                    "locality": "###LOC###",
                                    "neighborhood": "###NEIG###",
                                    ...
                                },
                                ...
                            ],
                        ...
                    },
                ...
            }

    The ###ID### token is relevant to later requests regarding reservations 
    at the specified venue. 

**********************************************************************

Reserve: 

    The Reserve function of the Resy REST API is complex. It is composed
    of three different requests and responses, each containing information
    pertinent to the next stage.

    The first request, referred to here as the 'find' operation
    is sent at a dynamically generated URL, since it includes 
    query values. We send a GET request. The URL is provided below, with tokens
    ###YEAR#### to represent the year number, ###MONTH### to represent
    the 2-digit month number, ###DAY### to represent the 2-digit day 
    number, ###TOK### to represent the login auth token, ###ID### to 
    represent the venue id, ###PS### to represent the party size:

        https://api.resy.com/4/find?day=###YEAR###-###MONTH###-###DAY###&x-resy-auth-token=###TOK###&lat=0&long=0&venue_id=###ID###&party_size=###PS###

    Where we fix the latitude and longitude to 0, although these can be
    used to make the search more precise. This request has no body associated, 
    but should use an Content Type header of 'application/x-www-form-urlencoded'. 
    The uses the Login fields described in the 'Login' section and the authorization
    field described in the 'APIKey' section of this document, along with the extra
    header:

        Referer: https://resy.com/

    Which is a standard HTTP request header. The reponse body is a JSON object
    with the following relevant structure:

        Body:
        
            {
                ...
                "results": 
                    {
                        ...
                        "venues":
                            [
                                ...
                                {
                                    ...
                                    "slots":
                                        [
                                            ...
                                            {
                                                ...
                                                "date": 
                                                    {
                                                        ...
                                                        "start": "###YR###:###MT###:###DY### ###HR###:###MN###",
                                                        ...
                                                    },
                                                "config": 
                                                    {
                                                        ...
                                                        "token": "###CONFTOKEN###",
                                                        ...
                                                    },
                                                ...
                                            },
                                            ...
                                        ],
                                    ...
                                },
                                ...
                            ],
                        ...
                    },
                ...
            }

    Where ###YR###, ###MT###, ###DY###, ###HR###, and ###MN###, are
    the year, month, day, hour, and minute respectfully of an open slot in
    military time format, and the ###CONFTOKEN### is an identifier for the
    slot used in the next request-response interaction.

    The next pair of HTTP messages is informally referred to as the 'config' 
    step. We send a GET request with no body. The request has a dynamic URL:

        https://api.resy.com/3/detail?day=###YEAR###-###MONTH###-###DAY###&x-resy-auth-token=###TOK###&lat=0&long=0&venue_id=###ID###&party_size=###PS###&config_id=###UECONFTOKEN###

    Where the first few tokens have the same definition as the URL from the 
    find step and ###UECONFTOKEN### is the identifier from the response from 
    the find step after url-encoding. This request uses the standard APIKey and Login headers.
    The server response looks as follows:

        Body:

            {
                ...
                "book_token":
                    {
                        ...
                        "value": "###BTOKEN###",
                        ...
                    },
                ...
            }

    Where ###BTOKEN### is an identifier used in the next step.

    The final step, denoted 'reserve', is where the reservation curated in the 
    past 2 steps is finalized and made persistent on Resy servers. It is a POST
    request, and uses the following dynamic URL:

        https://api.resy.com/3/book?day=###YEAR###-###MONTH###-###DAY###&x-resy-auth-token=###TOK###&lat=0&long=0&venue_id=###ID###&party_size=###PS###&config_id=###UECONFTOKEN###

    
    Where each token is used from the description of the previous step's URL. We send url form
    encoded data in the body that follows this structure:

        Body:

            book_token=###UEBTOKEN###&struct_payment_method=###UESTCPID###&source_id=resy.com-venue-details

    Where ###UEBTOKEN### is the url encoded version of the ###BTOKEN### value found 
    in the last step and ###UESTCPID### is the url encoded version of the following text:

        {"id":###PID###} 
    
    Where ###PID### is the payment method id from the login api function response. We use login
    headers as well as the following:

        Headers:

            Referer: https://resy.com/

    If the server response is any 200 code, the reservation has been made.    

**********************************************************************
*/
package resy
