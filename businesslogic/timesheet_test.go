package businesslogic

import (
	"testing"

	sw "github.com/moznobkin/jira-timesheet/generated/go"
	"github.com/stretchr/testify/require"
)

func Test_MonthlyTimesheet_real(t *testing.T) {

	// server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	if r.URL.Path != "/query" {
	// 		t.Errorf("Expected to request '/query', got: %s", r.URL.Path)
	// 	}
	// 	resbody, _ := ReadStruct[sw.MonthlyTimesheet]("../data/json/monthlyTs.json")
	// 	defer r.Body.Close()
	// 	w.Header().Add("Content-Type", "application/json")
	// 	w.Header().Add("Content-Length", strconv.Itoa(len(resbody)))
	// 	w.WriteHeader(http.StatusOK)
	// 	w.Write(resbody)
	// }))
	// defer server.Close()
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

func Test_MonthlyTimesheet_success(t *testing.T) {

	json.RegisterExtension(&timeDecodeExtension{})
	ts, err := ReadStruct[sw.MonthlyTimesheet]("../data/json/monthlyTs1.json")
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
