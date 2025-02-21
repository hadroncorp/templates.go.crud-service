package appointment

import "fmt"

type Status uint8

const (
	StatusUnknown = Status(iota)
	StatusScheduled
	StatusCancelled
)

// compile-time assertions
var _ fmt.Stringer = StatusUnknown

var (
	_statusFromPrimitivesMap = map[string]Status{
		"UNKNOWN":   StatusUnknown,
		"SCHEDULED": StatusScheduled,
		"CANCELLED": StatusCancelled,
	}
	_statusToPrimitivesMap = map[Status]string{
		StatusUnknown:   "UNKOWN",
		StatusScheduled: "SCHEDULED",
		StatusCancelled: "CANCELLED",
	}
)

func NewStatus(v string) Status {
	return _statusFromPrimitivesMap[v]
}

func (s Status) String() string {
	return _statusToPrimitivesMap[s]
}
