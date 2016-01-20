package daemon

// Starter can be started on a given port.
type Starter interface {
	Start(port int) error
}
