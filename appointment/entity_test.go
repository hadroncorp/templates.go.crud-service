package appointment_test

import (
	"testing"
	"time"

	"github.com/hadroncorp/service-template/appointment"
)

func TestNew(t *testing.T) {
	ap, err := appointment.New(appointment.NewArgs{
		ID:           "123",
		Title:        "Some title",
		PlaceID:      "some-place-id",
		ScheduledBy:  "some-user",
		ScheduleTime: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ap.ID())
	t.Log(ap.Title())
	t.Log(ap.PlaceID())
	t.Log(ap.TargetedTo())
	t.Log(ap.ScheduledBy())
	t.Log(ap.ScheduleTime().String())

	t.Log("\n")

	ap, err = appointment.New(appointment.NewArgs{
		ID:           "123",
		Title:        "Some title",
		PlaceID:      "some-place-id",
		ScheduledBy:  "some-user",
		ScheduleTime: time.Now(),
	}, appointment.WithTargetNew("some-employee"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ap.ID())
	t.Log(ap.Title())
	t.Log(ap.PlaceID())
	t.Log(ap.TargetedTo())
	t.Log(ap.ScheduledBy())
	t.Log(ap.ScheduleTime().String())
}
