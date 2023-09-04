/*
Author: Bruce Jagid
Created On: Aug 12, 2023

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
    app pkg and how to write features in the core app. This should
    NOT be read by a consumer of the app layer as a means of juicing
    out more functionality from the app layer. Everything in this
    section is subject to change and vary between releases and this
    document only serves to provide information to help maintainers.

    What is an Operation:

        Each app 'operation' is actually a separate go thread which is
        running code to perform a specific goal, i.e. reserving at a
        specified time. For each of these go threads, we create an
        "Operation" struct and store it in the app context. The 
        operation struct contains the ID for the operation, a 'Cancel'
        channel which allows us to stop the operation, an 'Output' 
        channel which allows us to read from the operation, a 'Result'
        field of type 'OperationResult' which contains a 'Timeable'
        (any struct with a Time() method) and an Err field, which 
        allows each operation to report an error or output successfully
        using one channel.

    Reading State From an Operation:

        In order to query an operation for its status or potential result, 
        one must first run the internal method 'AppCtx.updateOperationResult'
        which takes in an operation id and tries to read its output, and
        updates the operation's 'Status' field. One can then use this to
        make decisions.

    How To Cancel an Operation:

        In order to "cancel" an operation, any time an operation is
        sleeping, this sleeping is actually simulated using a select 
        statement and the "time.After(d Time)" channel type. The 
        "time.After(d time.Duration)" channel is a channel which sends
        a value to its output after "d" time.Duration has passed. We
        put this channel in a select statement with the cancel channel,
        which will cause the thread to block until one of those two
        channels activates. In the cancel case, we report an error of
        cancelled, and in the time.After() case, we continue execution.
    
    Writing Code For App Layer:

        If you are writing an internal function for the app layer, 
        there are a few conventions to be aware of. First off, all 
        successful operations must yield a ReservationTime as it stands.
        Second, operation state changes should always occur in place,
        i.e. once should use 'operations[i]' instead of a copy.

**********************************************************************
*/
package app
