package conflict

import (
	"time"

	"github.com/corsc/pagerduty-gcal/internal/gcal"
	"github.com/corsc/pagerduty-gcal/internal/pduty"
)

// SwapAPI will attempt to find a swap in the schedule
type SwapAPI struct {
	checker *CheckerAPI
}

// FindSwap attempts to find a swap for the supplied conflict
func (s *SwapAPI) FindSwap(schedule *pduty.Schedule, conflict *pduty.ScheduleEntry, calendars map[string]*gcal.Calendar) *pduty.ScheduleEntry {
	if s.checker == nil {
		s.checker = &CheckerAPI{}
	}

	for _, potentialSwap := range schedule.Entries {
		if potentialSwap.Start.Equal(conflict.Start) && potentialSwap.End.Equal(conflict.End) {
			// cant swap with yourself
			continue
		}

		if s.timeEqual(potentialSwap.Start, conflict.Start) && s.timeEqual(potentialSwap.End, conflict.End) {
			// check for conflicts
			if !s.checker.checkForConflict(conflict, calendars[potentialSwap.User.ID]) {
				return potentialSwap
			}
		}

	}

	return nil
}

// compare the hour and minute only
func (s *SwapAPI) timeEqual(a time.Time, b time.Time) bool {
	return a.Hour() == b.Hour() && a.Minute() == b.Minute()
}
