package runnable

import (
    "errors"
)

type Runnable interface {
    Run() (error)
    Stop() (error)
}
