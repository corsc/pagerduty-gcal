package conflict

import (
	"github.com/corsc/pagerduty-gcal/internal/gcal"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/corsc/pagerduty-gcal/internal/pduty"
)

func TestSwapAPI_FindSwap(t *testing.T) {
	day2morningUserFoo := &pduty.ScheduleEntry{
		User: &pduty.User{
			ID: "FOO",
		},
		Start: time.Date(2019, 01, 02, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2019, 01, 02, 8, 0, 0, 0, time.UTC),
	}

	day2afternoonUserFoo := &pduty.ScheduleEntry{
		User: &pduty.User{
			ID: "FOO",
		},
		Start: time.Date(2019, 01, 02, 8, 0, 0, 0, time.UTC),
		End:   time.Date(2019, 01, 02, 16, 0, 0, 0, time.UTC),
	}

	day3MorningUserBar := &pduty.ScheduleEntry{
		User: &pduty.User{
			ID: "BAR",
		},
		Start: time.Date(2019, 01, 03, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2019, 01, 03, 8, 0, 0, 0, time.UTC),
	}

	day4morningUserFoo := &pduty.ScheduleEntry{
		User: &pduty.User{
			ID: "FOO",
		},
		Start: time.Date(2019, 01, 02, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2019, 01, 02, 8, 0, 0, 0, time.UTC),
	}

	scenarios := []struct {
		desc             string
		inSchedule       *pduty.Schedule
		inConflict       *pduty.ScheduleEntry
		inCalendars      map[string]*gcal.Calendar
		inAlreadySwapped []*pduty.ScheduleEntry
		expected         *pduty.ScheduleEntry
	}{
		{
			desc: "swap available",
			inSchedule: &pduty.Schedule{
				Entries: []*pduty.ScheduleEntry{
					day2morningUserFoo,
					day3MorningUserBar,
				},
			},
			inConflict: day2morningUserFoo,
			inCalendars: map[string]*gcal.Calendar{
				"FOO": {},
				"BAR": {},
			},
			expected: day3MorningUserBar,
		},
		{
			desc: "already swapped",
			inSchedule: &pduty.Schedule{
				Entries: []*pduty.ScheduleEntry{
					day2morningUserFoo,
					day3MorningUserBar,
				},
			},
			inConflict: day2morningUserFoo,
			inCalendars: map[string]*gcal.Calendar{
				"FOO": {},
				"BAR": {},
			},
			inAlreadySwapped: []*pduty.ScheduleEntry{
				day3MorningUserBar,
			},
			expected: nil,
		},
		{
			desc: "no swaps possible",
			inSchedule: &pduty.Schedule{
				Entries: []*pduty.ScheduleEntry{
					day2morningUserFoo, day2afternoonUserFoo,
				},
			},
			inConflict: day2morningUserFoo,
			inCalendars: map[string]*gcal.Calendar{
				"FU": {
					Items: []*gcal.CalendarItem{
						{
							Start: time.Date(2019, 01, 02, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2019, 01, 02, 8, 0, 0, 0, time.UTC),
						},
					},
				},
				"BAR": {
					Items: []*gcal.CalendarItem{
						{
							Start: time.Date(2019, 01, 02, 0, 0, 0, 0, time.UTC),
							End:   time.Date(2019, 01, 02, 8, 0, 0, 0, time.UTC),
						},
					},
				},
			},
			expected: nil,
		},
		{
			desc: "cannot swap with yourself",
			inSchedule: &pduty.Schedule{
				Entries: []*pduty.ScheduleEntry{
					day2morningUserFoo, day4morningUserFoo,
				},
			},
			inConflict: day2morningUserFoo,
			inCalendars: map[string]*gcal.Calendar{
				"FU": {},
			},
			expected: nil,
		},
	}

	for _, s := range scenarios {
		scenario := s
		t.Run(scenario.desc, func(t *testing.T) {
			// call
			api := &SwapAPI{
				proposedSwaps:scenario.inAlreadySwapped,
			}
			result := api.FindSwap(scenario.inSchedule, scenario.inConflict, scenario.inCalendars)

			// validate
			assert.EqualValues(t, scenario.expected, result, scenario.desc)
		})
	}
}
