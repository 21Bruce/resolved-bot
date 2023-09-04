/*
Author: Bruce Jagid

**********************************************************************

General Purpose: 

    The runnable/cli pkg is a complete program representing the combo
    of the core app with a command line interface. This cli is a REPL,
    and acts as a running environment which is not supposed to crash.

    This pkg has two very big dependencies: The app pkg(back-end) and
    the cli pkg(front-end), so understanding internals can be learned
    best from those studying those pkgs. We provide only an 
    explanation of the public facing app here.

**********************************************************************

How To Use(In Code): 

    The functionality in the runnable/cli pkg can be used via 
    obtaining a ResolvedCLI struct. This struct requires a few
    fields. It takes in an io.Reader object to read input from
    in the In field, and io.Writer objects for the Out and Err
    fields, where it will report valid output or errors respectively.
    Finally, the Resolved CLI takes in an AppCtx, with the intent
    being that this CLI pkg can be easily repurposed between external
    APIs. Although the opentable go API is not complete yet, its 
    partial development has already yielded difficulties for this pkg.

**********************************************************************

How to Use(On Computer):

    The ResolvedCLI allows an end-user to access all of the core
    app features for the Resy(Opentable soon) implementation
    via a simple CLI. We explain each command here:

        1. login [-e email] [-p password] 

            This command checks if the given 
            login information link to a valid Resy
            account and if they do will save them
            for use in future commands
 
        2. logout 

            This command removes any saved login
            info if present

        3. search [-n name] [-l limit] 
            
            This command searches resy for the
            given name in the -n field and will
            limit the return results with the -l
            field, although this field is optional.
            The resy API returns valuable information
            about each found result such as locality 
            and region information so an end-user can
            verify that the returned restaurant is
            correct and the API yields the 'VenueID',
            an internal identifier used by resy to 
            specify restaurants and a piece of data
            that must be sent in a reservation command

        4. rats [-v venue-id] [-ps party-size] [-resD reservation-day] [-resT reservation-times] [-reqD request-date]
            
            This command sends a reservation request
            at a specified date down to the minute.
            The res is for a venue specified by the id in the 
            -v field, party size specified by the
            -ps field, reservation day specified
            in the -resD field(in YYYY:MM:DD format
            relative to the restaurant locale), 
            priority list of reservation times
            specified in the -resT field(each in HH:MM 
            military time format relative to the 
            restaurant locale), and the date to send
            the request to resy in the -reqD field
            (in YYYY:MM:DD:HH:MM miliatry time format
            relative to the local locale).

        5. rais [-v venue-id] [-ps party-size] [-resD reservation-day] [-resT reservation-times] [-i interval] 
            
            This command sends a reservation request
            on a repeated interval until a time is
            acquired or all possible times are past.
            The res is for a venue specified by the id in the 
            -v field, party size specified by the
            -ps field, reservation day specified
            in the -resD field(in YYYY:MM:DD format
            relative to the restaurant locale), 
            priority list of reservation times
            specified in the -resT field(each in HH:MM 
            military time format relative to the 
            restaurant locale), and the interval to send
            the request to resy in the -i field
            (in HH:MM format).

        6. list
            
            This command lists a history of operations, their IDs,
            and statuses

        7. cancel [-i id]
            
            This command will attempt to cancel the operations with
            ids specified in the -i field. Operations can only be
            cancelled if they are in progress 

        8. clean [-i id]
            
            This command will attempt to remove the operation
            from the history displayed by the list command. This
            will only work on operations that are not in progress.
            
        9. help 

            Display helpful info about commands    

        10. exit/quit 
            
            Leave the CLI environment 
 
**********************************************************************
*/
package cli
