package altirs

import "fmt"

// AltirsError is returned when the Altirs API responds with a non-2xx status.
type AltirsError struct {
	Message   string
	Status    int
	RequestID string
}

func (e *AltirsError) Error() string {
	return fmt.Sprintf("altirs: %s (status=%d, requestId=%s)", e.Message, e.Status, e.RequestID)
}
