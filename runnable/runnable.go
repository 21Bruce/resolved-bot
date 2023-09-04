package runnable

/*
Name: Runnable
Type: External Runnable Interface
Purpose: Define the minimal expected
behavior of a completed system, with
front-end and back-end behavior
*/
type Runnable interface {
    Run() (error)
    Stop() (error)
}
