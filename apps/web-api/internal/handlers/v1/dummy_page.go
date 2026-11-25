package v1

import "strconv"

type V1Page struct {
	Page  int    `json:"page"`
	Size  int    `json:"size"`
	Total string `json:"total"`
}
type V1List[T any] struct {
	Items      []T    `json:"items"`
	Pagination V1Page `json:"pagination"`
}
type V1DummyResult[T any] struct {
	Result T    `json:"result"`
	Dummy  bool `json:"dummy"`
}

func paginate[T any](items []T, page, size int) ([]T, V1Page) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	total := len(items)
	start := total
	if page-1 <= total/size {
		start = (page - 1) * size
	}
	end := total
	if size < total-start {
		end = start + size
	}
	return items[start:end], V1Page{Page: page, Size: size, Total: strconv.Itoa(total)}
}
