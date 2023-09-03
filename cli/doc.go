/*
Author: Bruce Jagid

**********************************************************************

General Purpose: 

    The cli pkg provides basic, highly opinionated utilities for 
    a command line applcation. So far, tokenizing, 
    parsing, and some validating, is supported. 

    In my(bruce) opinion, there should be no "default" argument to 
    any command. A command in a command line application should be
    like a function call, and the entire command line app like 
    a domain-specific repl, but with some improvements for 
    readability. I like all of my command arguments to have an 
    associated flag, both with a short name for fast convenience
    and a long name for intended verbosity.

    Since this library is so basic, we do not provide any special
    internal documentation in this doc.go - just use the code 
    comments scattered throughout the main code file in this dir.

**********************************************************************

How To Use:  

    In order to use this pkg, one first obtains a ParseCtx struct, 
    which is a config object for the tokenizer, parser, and validator.
    In obtaining a ParseCtx, you must define a few fields:

        1. OpenDelim and CloseDelim

            - These are used for the tokenizer, and they define
              the semantics for grouping characters with whitespace
              into a token. So usually tokenizing "a [  b c]" will
              yield ["a", "[", "b", "c", "]"], but with OpenDelim="["
              and CloseDelim="]", we have tokenization yielding 
              ["a", "  b c"]. These are defined on a per-ctx basis
              since different parse syntax may require different
              escape characters.
    
        2. Commands

            - This is a slice of Command structs which follow 
              the structure below

        3. Command

            - This struct defines what command strings should be 
              matched, what flags they take, and how to handle them.
              The description section adds an optional description,
              which is severely fucked up in the current arch,
              and I'll probably fix that soon. 

              Each command defines a Handler, which takes in a map 
              representing the validated flag to value map. So for
              example if I have a flag with name "n" that takes 1 arg,
              then, assuming the string was a valid command str,
              your handler will be passed a map denoted "in" such that
              in["n"] is a 1 element string slice containing your arg.
              The output of your handler must be a string since the 
              "ParseCtx.Parse" method outputs the result of your 
              handler.

              The last field is "Flags", which is a Flag slice,
              discussed in the next section.

        4. Flag

            - The Flag struct defines a Flag to a command. This
              struct takes in a few fields. The first is a Name,
              intended to be a short name, like "n" or "v". This
              flag can then be defined in a command using syntax,
              "-n" or "-v" respectively. Each flag can also 
              optionally have a LongName field, like "name" which 
              will allow the flag to be defined using "--name". 
              
              Flags also have Descriptions, which are severely 
              screwed up as mentioned with command descriptions.

              Finally, Flags must have a ValidationCtx of type
              FlagValidtionCtx, mentioned in the next section

        5. FlagValidtionCtx

            - The FlagValidationCtx allows some common
              validation steps to be defined once declaritively 
              instead of copied all over impertively. These include
              a MaxArgs and MinArgs field, which define how many
              arguments the flags take. MaxArgs can be set to infinite
              using the cli.InfiniteArgs value. Required is a bool
              which when set to true, will cause the parser to throw
              an error if the flag is missing
 
**********************************************************************
*/
package cli
