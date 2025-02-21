package employee

import "time"

type Employee struct {
	id       string
	fullName string
	hiredAt  time.Time
}

func (e Employee) ID() string {
	return e.id
}

func (e Employee) FullName() string {
	return e.fullName
}

func (e Employee) HiredAt() time.Time {
	return e.hiredAt
}
