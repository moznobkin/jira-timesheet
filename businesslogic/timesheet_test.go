package businesslogic

import (
	"testing"

	sw "github.com/moznobkin/jira-timesheet/generated/go"
	"github.com/stretchr/testify/require"
)

func Test_MonthlyTimesheet_success(t *testing.T) {

	json.RegisterExtension(&timeDecodeExtension{})
	ts, err := ReadStruct[sw.MonthlyTimesheet]("../data/json/monthlyTs.json")
	require.NoError(t, err)

	tsExpected, err := ReadStruct[sw.MonthlyTimesheetResult]("../data/json/monthlyTsResult.json")
	require.NoError(t, err)
	serv := (&TS{
		MonthlyTimesheet: *ts,
	})
	tsRes, err := serv.PreprocessMonthlyTimesheet()
	require.NoError(t, err)
	PushToJira(tsRes)

	require.EqualValues(t, tsExpected, tsRes)

}
