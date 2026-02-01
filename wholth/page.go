package wholth

// #ifndef WHOLTH_GO_INIT
// #include "wholth/wholth.h"
// #endif
import "C"
import (
	"errors"
	"wholth_go/util"
)

type Page struct {
	Handle *C.wholth_Page
}

func (t *Page) Fetch() error {
	err := C.wholth_pages_fetch(t.Handle)

	if !C.wholth_error_ok(&err) {
		return errors.New(toStr(err.message))
	}

	return nil
}

func (t *Page) Size() uint64 {
	return uint64(C.wholth_pages_array_size(t.Handle))
}

func (t *Page) Close() {
	C.wholth_pages_del(t.Handle)
}

func (t *Page) Pagination() util.Pagination {
	return util.Pagination{
		PageCurrent: uint64(C.wholth_pages_current_page_num(t.Handle)) + 1,
		PageMax:     uint64(C.wholth_pages_max(t.Handle)) + 1,
		Count:       uint64(C.wholth_pages_count(t.Handle)),
	}
}

func (t *Page) SkipTo(to int) {
	C.wholth_pages_skip_to(t.Handle, C.uint64_t(to))
}
