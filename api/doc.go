/*
Author: Bruce Jagid

**********************************************************************

General Purpose: 

    The api pkg is the abstraction layer for interacting with 
    external reservation services, such as resy and opentable.
    All services implementing the API interface 
    are sub-pkgs of api

**********************************************************************

API:

    The API interface specifies 3 methods:
    
        Login(params LoginParam) (*LoginResponse, error)
        Reserve(params ReserveParam) (*ReserveResponse, error)
        Search(params SearchParam) (*SearchResponse, error)
    
**********************************************************************

Login:

    The Login function takes in a set of login credentials and returns 
    a response. The login credentials vary by external service, with
    each defining its own set of necessary fields in each sub-pkg.
    The output of the Login function is a LoginResponse, which should
    be used as a token in the input params to a Reserve function 
    call. Login should always be used before a set of reservation
    calls and is only used for the purpose of making reservations.

**********************************************************************

Reserve:

    The Reserve function takes in a set of reserve parameters which 
    specify the date and a priority list of times to try and reserve
    at and produces a response indicating the time made or an error.
    This function's input parameters specifies a 'LoginResp' which
    must be obtained by a 'Login' api function call, though such a
    value only needs to be obtained before a series of Reserve calls.

**********************************************************************   

Search:

    The Search function takes in a set of query parameters which 
    specify the name of a restaurant and limit on responses and
    produces a response with a slice of search results. These results 
    contain necessary and helpful data both for identifying the 
    intended restaurant to reserve at and also for making a 
    reservation request.
    
**********************************************************************   
*/
package api
