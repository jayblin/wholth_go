package util

func ArrayFilter(ss []string, test func(string) bool) (ret []string) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}

type Pagination struct {
	PageCurrent uint64
	PageMax     uint64
	Count       uint64
}

type Status struct {
	Alias  string
	Message string
}
