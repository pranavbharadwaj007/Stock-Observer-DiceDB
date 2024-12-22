package observer

import "fmt"

var (
	ErrDuplicateObserver = fmt.Errorf("observer already exists")
	ErrObserverNotFound  = fmt.Errorf("observer not found")
	ErrInvalidPrice      = fmt.Errorf("invalid price")
)