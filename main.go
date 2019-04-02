package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/corsc/pagerduty-gcal/internal/conflict"
	"github.com/corsc/pagerduty-gcal/internal/gcal"
	"github.com/corsc/pagerduty-gcal/internal/pduty"
)

// NOTES:
// * For this tool to work it requires a public calendar event with the word "out" in the title.
// 	 Using the "out of office" event type in the google calendar UI will achieve this.

var (
	scheduleID        string
	startAsString     string
	days              int64
	daysBetweenShifts int64
)

func main() {
	// these are inputs that should come from command line or env
	apiKey, found := os.LookupEnv("PD_API_KEY")
	if !found {
		panic("PD_API_KEY must be set")
	}
	flag.StringVar(&scheduleID, "schedule", "", "schedule id (see README.md) for more info")
	flag.StringVar(&startAsString, "start", "", "start of the schedule")
	flag.Int64Var(&days, "days", 30, "days to add to start to define the schedule")
	flag.Int64Var(&daysBetweenShifts, "between", 3, "minimum number of days between shifts")
	flag.Parse()

	start, err := time.Parse("2006-01-02", startAsString)
	if err != nil {
		panic("failed to parse start with err: %s" + err.Error())
	}

	now := time.Now()
	if start.Before(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)) {
		panic("sorry you cannot re-write the past")
	}

	end := start.Add(time.Duration(daysBetweenShifts) * 24 * time.Hour)
	credentialsFile := "credentials.json"
	tokenFile := "token.json"

	// actual logic
	fmt.Printf("Loading schedule for %s to %s\n", start.Format(time.RFC3339), end.Format(time.RFC3339))
	scheduleStart := start.Add(time.Duration(-daysBetweenShifts * 24) * time.Hour)
	schedule, err := (&pduty.ScheduleAPI{}).GetSchedule(apiKey, scheduleID, scheduleStart, end)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Loading scheduled user details\n")
	participants, err := (&pduty.UserAPI{}).GetUsers(apiKey, schedule.Entries)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Loading calendars for scheduled users\n")
	calendars, err := (&gcal.CalendarAPI{}).GetCalendars(credentialsFile, tokenFile, participants, start, end)
	if err != nil {
		panic(err)
	}

	conflicts := checkForConflicts(schedule, calendars, daysBetweenShifts)
	if len(conflicts) == 0 {
		return
	}

	_ = findSwaps(schedule, conflicts, calendars)
}

func checkForConflicts(schedule *pduty.Schedule, calendars map[string]*gcal.Calendar, daysBetweenShifts int64) map[*pduty.ScheduleEntry]struct{} {
	fmt.Printf("Checking for conflicts\n")
	conflicts, err := (&conflict.CheckerAPI{}).Check(schedule, calendars, daysBetweenShifts)
	if err != nil {
		panic(err)
	}

	// output result
	if len(conflicts) == 0 {
		log.Printf("No conflicts found")
		return nil
	}

	fmt.Printf("Conflict (slot : user)\n")
	for scheduleEntry := range conflicts {
		fmt.Printf("%s to %s : %s\n", scheduleEntry.Start, scheduleEntry.End, scheduleEntry.User.Name)
	}

	return conflicts
}

func findSwaps(schedule *pduty.Schedule, conflicts map[*pduty.ScheduleEntry]struct{}, calendars map[string]*gcal.Calendar) map[*pduty.ScheduleEntry]*pduty.ScheduleEntry {
	fmt.Printf("\nPotential Swaps (slot - user -> slot - user)\n")
	swapAPI := &conflict.SwapAPI{}
	swaps := map[*pduty.ScheduleEntry]*pduty.ScheduleEntry{}
	for conf := range conflicts {
		swap := swapAPI.FindSwap(schedule, conf, calendars)
		if swap != nil {
			fmt.Printf("%s - %s - %s", conf.Start, conf.End, conf.User.Name)
			fmt.Printf(" -> %s - %s - %s\n", swap.Start, swap.End, swap.User.Name)
			swaps[conf] = swap
			continue
		}

		fmt.Fprintf(os.Stderr, "\n ==> SWAP NOT FOUND FOR %s - %s - %s <==\n\n", conf.Start, conf.End, conf.User.Name)
	}

	return swaps
}
