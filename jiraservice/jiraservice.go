package jirasevice

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	jira "github.com/andygrunwald/go-jira"
	sw "github.com/moznobkin/jira-timesheet/generated/go"
)

type JiraService interface {
	DeleteWorklog(issueID string, worklogID string) (*http.Response, error)
	DeleteWorklogs(wls []sw.WorklogResult) []error
	FindWorklogs(startDate time.Time, endDate time.Time, accountId string) ([]sw.WorklogResult, error)
	AddWorklogs(wls []sw.WorklogResult) ([]*jira.WorklogRecord, []error)
	UserName() string
}
type jiraService struct {
	client   *jira.Client
	userName string
}

func (s *jiraService) DeleteWorklog(issueID string, worklogID string) (*http.Response, error) {
	apiEndpoint := fmt.Sprintf("rest/api/2/issue/%s/worklog/%s", issueID, worklogID)

	req, err := s.client.NewRequestWithContext(context.Background(), "DELETE", apiEndpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req, nil)
	if err != nil {
		return resp.Response, err
	}

	return resp.Response, nil
}

func (s *jiraService) DeleteWorklogs(wls []sw.WorklogResult) []error {
	errs := []error{}
	for _, wl := range wls {
		_, err := s.DeleteWorklog(wl.IssueId, wl.Id)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}
func (s *jiraService) getSelf() (*jira.User, error) {
	u, _, err := s.client.User.GetSelf()
	if err != nil {
		return nil, err
	}
	return u, nil
}
func (s *jiraService) findIssues(startDate time.Time, endDate time.Time) ([]jira.Issue, error) {
	issues, _, err := s.client.Issue.Search(fmt.Sprintf("worklogAuthor='%s' and worklogDate>=%s and worklogDate<=%s", s.userName, startDate.Format("2006-01-02"), endDate.Format("2006-01-02")), &jira.SearchOptions{StartAt: 0, MaxResults: 1000})
	if err != nil {
		return nil, err
	}
	return issues, nil
}
func (s *jiraService) FindWorklogs(startDate time.Time, endDate time.Time, accountId string) ([]sw.WorklogResult, error) {
	issues, err := s.findIssues(startDate, endDate)
	if err != nil {
		return nil, err
	}
	result := []sw.WorklogResult{}
	for _, i := range issues {
		worklog, _, err := s.client.Issue.GetWorklogs(i.ID)
		if err != nil {
			return nil, err
		}
		for _, wl := range worklog.Worklogs {
			started := time.Time(*wl.Started)

			if (startDate.Before(started) || startDate.Equal(started)) && (endDate.After(started) || endDate.Equal(started)) && wl.Author.Name == accountId {
				result = append(result, sw.WorklogResult{
					Id:       wl.ID,
					IssueId:  i.Key,
					Duration: int32(wl.TimeSpentSeconds / 60),
					Date:     started,
				})
			}
		}
	}
	return result, nil
}
func (s *jiraService) AddWorklogs(wls []sw.WorklogResult) ([]*jira.WorklogRecord, []error) {
	result := make([]*jira.WorklogRecord, len(wls))
	errors := []error{}
	for i, wl := range wls {

		jiraTime := jira.Time(wl.Date)
		res, _, err := s.client.Issue.AddWorklogRecord(wl.IssueId, &jira.WorklogRecord{
			Started:          &jiraTime,
			TimeSpentSeconds: int(60 * wl.Duration),
		})
		if err != nil {
			errors = append(errors, err)
		} else {
			result[i] = res
		}
	}
	return result, errors
}
func NewService(token string) (JiraService, error) {
	tp := jira.BearerAuthTransport{
		Token: token,
	}
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client, err := jira.NewClient(tp.Client(), "https://servicedesk.veon.com/")
	if err != nil {
		return nil, err
	}
	service := &jiraService{
		client: client,
	}
	u, err := service.getSelf()
	if err != nil {
		return nil, err
	}
	service.userName = u.Name
	return service, nil
}
func (s *jiraService) UserName() string {
	return s.userName
}
