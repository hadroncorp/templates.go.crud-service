package appointment_test

import (
	"context"
	"testing"
	"time"

	"github.com/hadroncorp/service-template/appointment"
)

func TestNew(t *testing.T) {
	ctx := context.Background()
	scheduleTime := time.Now().Add(time.Hour * 24)
	ap, err := appointment.New(ctx, appointment.NewArgs{
		ID:           "123",
		Title:        "Some title",
		PlaceID:      "some-place-id",
		ScheduledBy:  "some-user",
		ScheduleTime: scheduleTime,
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
	t.Logf("%+v", ap.PullEvents())

	t.Log("\n")

	ap, err = appointment.New(ctx, appointment.NewArgs{
		ID:           "123",
		Title:        "Some title",
		PlaceID:      "some-place-id",
		ScheduledBy:  "some-user",
		ScheduleTime: scheduleTime,
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
	t.Logf("%+v", ap.PullEvents())
}
