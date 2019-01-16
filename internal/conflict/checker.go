package conflict

import (
	"github.com/corsc/pagerduty-gcal/internal/gcal"
	"github.com/corsc/pagerduty-gcal/internal/pduty"
)

// CheckerAPI will compare the schedule with the calendar and return any conflicts
type CheckerAPI struct{}

// Check is the main entrypoint for this struct
func (c *CheckerAPI) Check(schedule *pduty.Schedule, calendars map[string]*gcal.Calendar) (map[*pduty.ScheduleEntry]struct{}, error) {
	out := map[*pduty.ScheduleEntry]struct{}{}

	for _, scheduleEntry := range schedule.Entries {
		scheduleUserID := scheduleEntry.User.ID

		calendar := calendars[scheduleUserID]
		if calendar == nil {
			// no scheduled out of office time
			continue
		}

		conflict := c.checkForConflict(scheduleEntry, calendar)
		if conflict {
			out[scheduleEntry] = struct{}{}
		}
	}

	return out, nil
}

func (c *CheckerAPI) checkForConflict(shift *pduty.ScheduleEntry, calendar *gcal.Calendar) bool {
	for _, calendarEntry := range calendar.Items {
		if calendarEntry.Start.Equal(shift.Start) {
			return true
		}

		if calendarEntry.Start.After(shift.Start) {
			if calendarEntry.Start.Before(shift.End) {
				return true
			}
		}

		if calendarEntry.Start.Before(shift.Start) {
			if calendarEntry.End.After(shift.Start) {
				return true
			}
		}
	}
	return false
}
