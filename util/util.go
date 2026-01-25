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

type AliasableEntity interface {
	EntityAlias() string
}

type EntityAliasAware[T AliasableEntity] struct {
	Dummy T
}

func (p EntityAliasAware[T]) EntityAlias() string {
	return p.Dummy.EntityAlias()
}

type PaginatableList[T AliasableEntity] struct {
	EntityAliasAware[T]
	Pagination
	Values []T
	Q      string
}

type Status struct {
	Alias   string
	Message string
}
