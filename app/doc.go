/*
**********************************************************************

General Purpose: 

    The app pkg builds on top of the api layer, adding functionality
    that represents the core app functions of the bot. These include
    scheduling reserve at time and reserve at interval operations
    and managing these operations concurrently, maintaining a 
    history of operations which can be cleansed, and allowing for the
    cancellation of operations. In this document, we explain both
    the public API and how the internals work.

**********************************************************************

Public API:

    Here we go through the front facing functions of this pkg:

        1. ScheduleReserveAtIntervalOperation(ReserveAtIntervalParam)(int64, error)

            - Description: This func takes in a set of parameters
              specifying the date and times to reserve at, the 
              restaurant to reserve at, party size, and an interval
              to retry the reservation on and returns the id of the
              running operation on success

        2. ScheduleReserveAtTimeOperation(ReserveAtTimeParam)(int64, error)

            - Description: This func takes in a set of parameters
              specifying the date and times to reserve at, the 
              restaurant to reserve at, party size, and a time
              to send the request to the external API at. This time
              must be in UTC. The func returns an id of the running 
              operation on success

        3. CancelOperation(id int64)(error) 

            - Description: This func takes in an id and attempts to
              cancel it.

        4. Login(LoginParam)(error)

            - Description: This func takes in a set of login params 
              and attempts to call the login function of the external 
              api on it. If that call succeeds, it will store the 
              login params and use them as defaults in future requests

        5. Search(SearchParam)(*SearchResponse, error)

            - Description: This func takes in a SearchParam containing
              a name and a potential limit and returns a set of 
              responses

        6. CleanOperation(id int64)(error) 

            - Description: This func takes in an id and on success
              removes its associated operation from the internal
              list history of operations. Can only be done if
              an operation is completed or cancelled.

        7. Logout()(error) 

            - Description: Removes any saved login defaults

        8. OperationsToString()(string, error) 

            - Description: Stringifies the internal list
            of operations

**********************************************************************

App Internals:

    The following serves as a short guide on the internals of the
    app pkg and how to write features in the core app.

**********************************************************************
*/
package app
