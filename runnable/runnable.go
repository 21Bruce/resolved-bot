package runnable

type Runnable interface {
    Run() (error)
    Stop() (error)
}
