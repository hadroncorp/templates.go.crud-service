package appointment

import (
    "time"

    "github.com/hadroncorp/service-template/employee"
    "github.com/hadroncorp/service-template/place"
    userplatform "github.com/hadroncorp/service-template/user"
)

// --> Read <--

type ReadModel struct {
    ID           string
    Title        string
    Place        place.Place
    TargetedTo   employee.Employee
    ScheduledBy  userplatform.User
    ScheduleTime time.Time
}

type ListUserReadModel struct {
    ID           string
    Title        string
    Place        place.Place
    TargetedTo   employee.Employee
    ScheduleTime time.Time
}

type ListPlaceReadModel struct {
    ID           string
    Title        string
    TargetedTo   employee.Employee
    ScheduledBy  userplatform.User
    ScheduleTime time.Time
}
