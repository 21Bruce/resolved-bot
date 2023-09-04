/*
Author: Bruce Jagid
Created On: Aug 12, 2023

**********************************************************************

General Purpose: 

    The runnable pkg provides a common namespace and interface
    to systems that wrap around the core app. Currently, the only
    working program made with the core app is in the runnable/cli
    directory, although the intent is to expand this to include
    some servers. 

    We hope this abstraction and separation of concerns will achieve
    a few important things. First, since we strive to write all of 
    the code without use of external pkgs, we have to implement a 
    lot of our own front-end libraries, like our own cli parser or
    http server. In this case, it is much better to have separate 
    pkgs where those utilities live and then one common place
    where one can 'attach' there front-end to the core app, which
    is what we want the runnable sub-pkg space to be. We also want a
    common way of running complete systems, such that we may be able
    to select between available systems at runtime through flags.
    
**********************************************************************
*/
package runnable
