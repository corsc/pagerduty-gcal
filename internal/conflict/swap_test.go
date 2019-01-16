package conflict

import (
	"testing"
	"time"

	"github.com/corsc/pagerduty-gcal/internal/gcal"
	"github.com/corsc/pagerduty-gcal/internal/pduty"
	"github.com/stretchr/testify/assert"
)

func TestSwapAPI_FindSwap(t *testing.T) {
	scenarios := []struct {
		desc        string
		inSchedule  *pduty.Schedule
		inConflict  *pduty.ScheduleEntry
		inCalendars map[string]*gcal.Calendar
		expected    *pduty.ScheduleEntry
	}{
		{
			desc: "no swap",
			inSchedule: &pduty.Schedule{
				Entries: []*pduty.ScheduleEntry{
					{
						User: &pduty.User{
							ID: "FOO",
						},
						Start: time.Date(2019, 01, 02, 0, 0, 0, 0, time.UTC),
						End:   time.Date(2019, 01, 02, 8, 0, 0, 0, time.UTC),
					},
					{
						User: &pduty.User{
							ID: "BAR",
						},
						Start: time.Date(2019, 01, 03, 0, 0, 0, 0, time.UTC),
						End:   time.Date(2019, 01, 03, 8, 0, 0, 0, time.UTC),
					},
				},
			},
			inConflict: &pduty.ScheduleEntry{
				User: &pduty.User{
					ID: "FOO",
				},
				Start: time.Date(2019, 01, 02, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2019, 01, 02, 8, 0, 0, 0, time.UTC),
			},
			inCalendars: map[string]*gcal.Calendar{
				"BAR": {},
			},
			expected: &pduty.ScheduleEntry{
				User: &pduty.User{
					ID: "BAR",
				},
				Start: time.Date(2019, 01, 03, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2019, 01, 03, 8, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, s := range scenarios {
		scenario := s
		t.Run(scenario.desc, func(t *testing.T) {
			// call
			api := &SwapAPI{}
			result := api.FindSwap(scenario.inSchedule, scenario.inConflict, scenario.inCalendars)

			// validate
			assert.EqualValues(t, scenario.expected, result, scenario.desc)
		})
	}

}
