package outbound

import "errors"

// ErrOutboundLeaseLost is returned when a delivery-state write is rejected because
// another instance holds the mapping lease (or this instance lost its lease mid-flight).
var ErrOutboundLeaseLost = errors.New("outbound delivery lease not held by this instance")
