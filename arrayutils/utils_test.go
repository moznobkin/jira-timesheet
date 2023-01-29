package arrayutils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Compare_Split(t *testing.T) {

	a := []int{1, 3, 5, 6, 9, 44, 55, 65}
	b := []int{2, 5, 6, 7, 8, 9, 46, 55, 76, 77}
	ua, ub, c := CompareSorted(a, b, func(a1, a2 int) int { return a1 - a2 })
	require.EqualValues(t, ua, []int{1, 3, 44, 65})
	require.EqualValues(t, ub, []int{2, 7, 8, 46, 76, 77})
	require.EqualValues(t, c, []int{5, 6, 9, 55})

}
