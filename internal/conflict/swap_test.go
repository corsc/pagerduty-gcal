package conflict

import (
	"github.com/corsc/pagerduty-gcal/internal/gcal"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/corsc/pagerduty-gcal/internal/pduty"
)


var (
	sourceUserID = "FOO"
	destinationUserID = "BAR"

	day2Morning= time.Date(2019, 01, 02, 0, 0, 0, 0, time.UTC)
	day2Afternoon = time.Date(2019, 01, 02, 8, 0, 0, 0, time.UTC)
	day2Evening = time.Date(2019, 01, 02, 16, 0, 0, 0, time.UTC)

	day3Morning = day2Morning.Add(24 * time.Hour)
	day3Afternoon = day2Afternoon.Add(24 * time.Hour)

	day2MorningSource = &pduty.ScheduleEntry{
		User: &pduty.User{
			ID: sourceUserID,
		},
		Start: day2Morning,
		End:   day2Afternoon,
	}

	day2AfternoonDestination = &pduty.ScheduleEntry{
		User: &pduty.User{
			ID: sourceUserID,
		},
		Start: day2Afternoon,
		End:   day2Evening,
	}

	day3MorningDestination = &pduty.ScheduleEntry{
		User: &pduty.User{
			ID: destinationUserID,
		},
		Start: day3Morning,
		End:   day3Afternoon,
	}

	day4MorningUserFoo = &pduty.ScheduleEntry{
		User: &pduty.User{
			ID: sourceUserID,
		},
		Start: day2Morning,
		End:   day2Afternoon,
	}
)

func TestSwapAPI_FindSwap(t *testing.T) {
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
					day2MorningSource,
					day3MorningDestination,
				},
			},
			inConflict: day2MorningSource,
			inCalendars: map[string]*gcal.Calendar{
				sourceUserID: {},
				destinationUserID: {},
			},
			expected: day3MorningDestination,
		},
		{
			desc: "already swapped",
			inSchedule: &pduty.Schedule{
				Entries: []*pduty.ScheduleEntry{
					day2MorningSource,
					day3MorningDestination,
				},
			},
			inConflict: day2MorningSource,
			inCalendars: map[string]*gcal.Calendar{
				sourceUserID: {},
				destinationUserID: {},
			},
			inAlreadySwapped: []*pduty.ScheduleEntry{
				day3MorningDestination,
			},
			expected: nil,
		},
		{
			desc: "swap not possible because destination is not available",
			inSchedule: &pduty.Schedule{
				Entries: []*pduty.ScheduleEntry{
					day2MorningSource,
					day3MorningDestination,
				},
			},
			inConflict: day2MorningSource,
			inCalendars: map[string]*gcal.Calendar{
				sourceUserID: {},
				destinationUserID: {
					Items: []*gcal.CalendarItem{
						{
							Start: day2Morning,
							End:   day2Afternoon,
						},
					},
				},
			},
			expected: nil,
		},
		{
			desc: "swap not possible because original user cannot take the replacement's shift",
			inSchedule: &pduty.Schedule{
				Entries: []*pduty.ScheduleEntry{
					day2MorningSource,
					day3MorningDestination,
				},
			},
			inConflict: day2MorningSource,
			inCalendars: map[string]*gcal.Calendar{
				sourceUserID: {
					Items: []*gcal.CalendarItem{
						{
							Start: day3Morning,
							End:   day3Afternoon,
						},
					},
				},
				destinationUserID: {},
			},
			expected: nil,
		},
		{
			desc: "cannot swap with yourself",
			inSchedule: &pduty.Schedule{
				Entries: []*pduty.ScheduleEntry{
					day2MorningSource, day4MorningUserFoo,
				},
			},
			inConflict: day2MorningSource,
			inCalendars: map[string]*gcal.Calendar{
				sourceUserID: {},
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
