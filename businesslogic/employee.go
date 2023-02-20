package businesslogic

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/moznobkin/jira-timesheet/arrayutils"
	sw "github.com/moznobkin/jira-timesheet/generated/go"
	j "github.com/moznobkin/jira-timesheet/jiraservice"
	"github.com/pkg/errors"
)

type EmployeeService interface {
	Process(chan Message) error
}
type employeeService struct {
	emp         sw.EmployeeResult
	jiraService j.JiraService
}

type MessageType string

const (
	DeletedExisting   MessageType = "Del"
	RetreivedExisting             = "Ret"
	CreatedNew                    = "Create"
)

type Message struct {
	Type       MessageType
	Count      int
	TotalCount int
}

func (m Message) String() string {
	return fmt.Sprintf("%s count: %d total: %d", m.Type, m.Count, m.TotalCount)
}

func NewEmployeeService(emp sw.EmployeeResult) (EmployeeService, error) {
	js, err := j.NewService(emp.Token)
	if err != nil {
		return nil, err
	}
	return &employeeService{emp: emp, jiraService: js}, nil
}

func compareWorklogs(a sw.WorklogResult, b sw.WorklogResult) int {
	if a.IssueId == b.IssueId {
		if a.Date == b.Date {
			return int(a.Duration - b.Duration)
		} else {
			return int(a.Date.UnixMicro() - b.Date.UnixMicro())
		}
	}
	return strings.Compare(a.IssueId, b.IssueId)
}
func (e *employeeService) Process(c chan Message) error {
	now := time.Now()
	existingWl, err := e.jiraService.FindWorklogs(e.emp.StartDate, e.emp.EndDate, e.jiraService.UserName())
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Could not get existing worklogs for %s", e.emp.Name))
	}
	c <- Message{
		Type:       RetreivedExisting,
		Count:      len(existingWl),
		TotalCount: len(existingWl),
	}
	sort.SliceStable(existingWl, func(a, b int) bool {
		return compareWorklogs(existingWl[a], existingWl[b]) < 0
	})
	sort.SliceStable(e.emp.Worklogs, func(a, b int) bool {
		return compareWorklogs(e.emp.Worklogs[a], e.emp.Worklogs[b]) < 0
	})
	toDelete, toAdd, common := arrayutils.CompareSorted(existingWl, e.emp.Worklogs, compareWorklogs)
	e.jiraService.DeleteWorklogs(toDelete)
	c <- Message{
		Type:       DeletedExisting,
		Count:      len(toDelete),
		TotalCount: len(toDelete),
	}
	e.jiraService.AddWorklogs(toAdd)
	c <- Message{
		Type:       CreatedNew,
		Count:      len(toAdd),
		TotalCount: len(toAdd),
	}
	log.Printf("Processed %s: deleted %d, added %d, existed %d, time spent %f", e.emp.Name, len(toDelete), len(toAdd), len(common), float64((time.Now().UnixMilli()-now.UnixMilli()))/1000)

	return nil
}
