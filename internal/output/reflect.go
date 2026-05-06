package output

import "reflect"

type sliceView struct {
	Len int
	at  func(i int) any
}

func (s *sliceView) At(i int) any { return s.at(i) }

func derefSlice(v any) *sliceView {
	if v == nil {
		return nil
	}
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil
	}
	rvCaptured := rv
	return &sliceView{
		Len: rv.Len(),
		at: func(i int) any {
			return rvCaptured.Index(i).Interface()
		},
	}
}
