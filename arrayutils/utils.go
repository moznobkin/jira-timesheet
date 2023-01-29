package arrayutils

type mapfunc[itemsType, toType any] func(itemsType) toType

func MapArrays[itemsType any, toType any](items []itemsType, fn mapfunc[*itemsType, *toType]) []toType {
	if items == nil {
		return nil
	}
	result := make([]toType, len(items))
	for i, item := range items {
		res := fn(&item)
		result[i] = *res
	}
	return result
}

func CompareSorted[T any](a []T, b []T, cmpf func(a, b T) int) ([]T, []T, []T) {
	uniqueA, uniqueB, common := []T{}, []T{}, []T{}
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		cmp := cmpf(a[i], b[j])
		switch {
		case cmp < 0:
			uniqueA = append(uniqueA, a[i])
			i++

		case cmp == 0:
			common = append(common, a[i])
			i++
			j++

		case cmp > 0:
			uniqueB = append(uniqueB, b[j])
			j++
		}
	}
	uniqueA = append(uniqueA, a[i:]...)
	uniqueB = append(uniqueB, b[j:]...)
	return uniqueA, uniqueB, common
}
