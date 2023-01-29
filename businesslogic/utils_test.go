package businesslogic

import (
	"testing"
	"time"

	"github.com/moznobkin/jira-timesheet/arrayutils"
	"github.com/stretchr/testify/require"
)

func Test_MinDate(t *testing.T) {
	dates := arrayutils.MapArrays([]string{"2023-04-20", "2023-01-02", "2022-06-02", "2022-03-05"}, func(s *string) *time.Time {
		t, err := time.Parse("2006-01-02", *s)
		if err != nil {
			return nil
		}
		return &t
	})
	dates = append(dates, time.Time{})
	require.Equal(t, time.Date(2022, time.March, 5, 0, 0, 0, 0, time.UTC), MinDate(dates...))
	require.Equal(t, time.Date(2023, time.April, 20, 0, 0, 0, 0, time.UTC), MaxDate(dates...))
}
