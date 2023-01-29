package businesslogic

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/jinzhu/now"

	"github.com/moznobkin/jira-timesheet/arrayutils"
	sw "github.com/moznobkin/jira-timesheet/generated/go"
)

type Iter[T any] interface {
	Next() (T, bool)
}

type DaysIter func() (time.Time, bool)
type iterDummy struct {
	days    []time.Time
	current int
}

func (i *iterDummy) Next() (time.Time, bool) {
	if i.current >= len(i.days) {
		return time.Time{}, false
	}
	res := i.days[i.current]
	i.current++
	return res, true
}

type TS struct {
	sw.MonthlyTimesheet
	workDays []time.Time
}

func (ts *TS) PreprocessMonthlyTimesheet() (*sw.MonthlyTimesheetResult, error) {
	ts.Norm()
	err := ts.Validate()
	ts.preCalculate()
	if err != nil {
		return nil, err
	}
	templatesMap := map[string]sw.WorklogTemplate{}

	for _, t := range ts.WorklogTempates {
		if t.EmployeeName != "" {
			templatesMap[t.EmployeeName] = t
		} else {
			templatesMap[t.Category] = t
		}
	}
	return &sw.MonthlyTimesheetResult{
		StartDate: ts.StartDate,
		EndDate:   ts.EndDate,
		Employees: arrayutils.MapArrays(ts.Employees, func(e *sw.Employee) *sw.EmployeeResult {

			template, ok := templatesMap[e.Name]
			if !ok {
				template = templatesMap[e.Category]
			}
			wls, total := ts.generateWorklogResults(e, template)
			return &sw.EmployeeResult{
				Token:         e.Token,
				Name:          e.Name,
				Category:      e.Category,
				StartDate:     MaxDate(e.StartDate, ts.StartDate),
				EndDate:       MinDate(e.EndDate, ts.EndDate),
				Worklogs:      wls,
				WorklogsTotal: int32(total),
			}

		}),
		WorklogTempates: []sw.WorklogTemplate{},
		AltWorkdays:     []sw.AltWorkday{},
	}, nil
}

func (ts *TS) generateWorklogResults(e *sw.Employee, template sw.WorklogTemplate) ([]sw.WorklogResult, int) {
	var result []sw.WorklogResult
	next := ts.getEmployeeWorkdays(e)
	d, ok := next()
	total := 0
	for ; ok; d, ok = next() {
		for _, wl := range template.Worklogs {
			result = append(result, sw.WorklogResult{
				IssueId:  wl.IssueId,
				Duration: wl.Duration,
				Date:     d,
			})
			total += int(wl.Duration)
		}
	}
	return result, total
}
func (ts *TS) preCalculate() {
	var workdays []time.Time
	altWorkdaysMap := map[time.Time]bool{}
	for _, a := range ts.AltWorkdays {
		if (a.Day == ts.StartDate || ts.StartDate.Before(a.Day)) && (a.Day == ts.EndDate || a.Day.Before(ts.EndDate)) {
			altWorkdaysMap[a.Day] = a.Value
		}
	}

	for d := ts.StartDate; d == ts.EndDate || d.Before(ts.EndDate); d = d.AddDate(0, 0, 1) {
		isWeekEnd := d.Weekday() == time.Sunday || d.Weekday() == time.Saturday
		altValue, isAltWorkday := altWorkdaysMap[d]
		isWorkday := (isAltWorkday && altValue) || (!isAltWorkday && !isWeekEnd)
		if isWorkday {
			workdays = append(workdays, d)
		}
	}
	ts.workDays = workdays
}
func (ts *TS) getEmployeeWorkdays(e *sw.Employee) DaysIter {
	iter := &iterDummy{
		days: ts.workDays,
	}
	i := 0
	vacDays := map[time.Time]bool{}
	for _, v := range e.Vacations {
		if v.EndDate.Before(ts.StartDate) || v.StartDate.After(ts.EndDate) {
			continue
		}
		cur := v.StartDate
		for ; (cur.Before(v.EndDate) || cur.Equal(v.EndDate)) && (cur.Before(ts.EndDate) || cur.Equal(ts.EndDate)); cur = cur.AddDate(0, 0, 1) {
			if cur.After(ts.StartDate) || cur.Equal(ts.StartDate) {
				vacDays[cur] = true
			}
		}
	}
	return func() (time.Time, bool) {
		localIter := iter
		i++
		day, ok := localIter.Next()
		for ; ok; day, ok = localIter.Next() {
			_, isVacation := vacDays[day]
			if !isVacation && (e.StartDate.IsZero() || e.StartDate.Before(day) || e.StartDate.Equal(day)) && (e.EndDate.IsZero() || e.EndDate.After(day) || e.EndDate.Equal(day)) {
				return day, true
			}
		}
		return day, false
	}
}

func (ts *TS) getWorkdays(e *sw.Employee) []time.Time {
	result := []time.Time{}
	for d := ts.StartDate.Truncate(24 * time.Hour); d == ts.EndDate || d.After(ts.EndDate); d = d.AddDate(0, 0, 1) {

		if d.Weekday() != time.Sunday || d.Weekday() != time.Saturday {
			result = append(result, d)
		}
	}
	return result
}

func (ts *TS) Norm() {
	if ts.StartDate.IsZero() {
		ts.StartDate = now.BeginningOfMonth()
	}
	if ts.EndDate.IsZero() {
		ts.EndDate = now.EndOfMonth()
	}
	ts.StartDate = ts.StartDate.Truncate(24 * time.Hour)
}

func (ts *TS) Validate() error {
	var errs []string
	for _, t := range ts.WorklogTempates {
		if t.EmployeeName == "" && t.Category == "" {
			errs = append(errs, "EmployeeName and Category of WorklogTemplate cannot be both empty")
		}
		if t.EmployeeName != "" && t.Category != "" {
			errs = append(errs, "EmployeeName and Category of WorklogTemplate cannot be specified both")
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, ",\n"))
	}
	return nil
}

func PushToJira(tsRes *sw.MonthlyTimesheetResult) {
	for _, e := range tsRes.Employees {
		serv, err := NewEmployeeService(e)
		if err != nil {
			log.Printf("error for %s", e.Name)
		} else {
			serv.Process()
		}
	}
}
