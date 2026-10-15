package worker

import (
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type emailWorkerMeta interface {
	meta() (eventType, name string)
}

func TestNewEmailWorkersRegistration(t *testing.T) {
	Convey("Given every email worker returned by NewEmailWorkers", t, func() {
		workers := NewEmailWorkers(nil, nil, testutil.NewTestLogger(), nil, testBaseURL)

		So(workers, ShouldHaveLength, 11)

		eventTypes := map[string]bool{}
		names := map[string]bool{}
		for _, w := range workers {
			m, ok := w.(emailWorkerMeta)
			So(ok, ShouldBeTrue)
			eventType, name := m.meta()

			Convey(name+" consumes a known, unique event type under a unique name", func() {
				So(events.IsKnownEventType(events.EventType(eventType)), ShouldBeTrue)
				So(eventType, ShouldNotBeEmpty)
				So(name, ShouldNotBeEmpty)
			})

			So(eventTypes[eventType], ShouldBeFalse)
			So(names[name], ShouldBeFalse)
			eventTypes[eventType], names[name] = true, true
		}
	})
}
