// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package postgresgen

import (
	"database/sql"
	"time"
)

type Appointment struct {
	AppointmentID  string
	Title          string
	PlaceID        string
	TargetedTo     sql.NullString
	ScheduledBy    string
	ScheduledTime  time.Time
	Notes          sql.NullString
	StatusType     string
	CreateTime     time.Time
	CreateBy       string
	LastUpdateTime time.Time
	LastUpdateBy   string
	RowVersion     int64
	IsDeleted      bool
}

type Employee struct {
	EmployeeID string
	FullName   string
}

type Organization struct {
	OrganizationID string
	Name           string
	CreateTime     time.Time
	CreateBy       string
	LastUpdateTime time.Time
	LastUpdateBy   string
	RowVersion     int64
	IsDeleted      bool
}

type Place struct {
	PlaceID     string
	DisplayName string
	Location    interface{}
}

type PlatformUser struct {
	UserID   string
	FullName string
}
