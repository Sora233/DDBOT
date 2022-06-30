package template

import (
	"fmt"
	"reflect"
)

func list(v ...interface{}) []interface{} {
	return v
}

func push(list interface{}, v interface{}) []interface{} {
	l, err := mustPush(list, v)
	if err != nil {
		panic(err)
	}

	return l
}

func mustPush(list interface{}, v interface{}) ([]interface{}, error) {
	tp := reflect.TypeOf(list).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(list)

		l := l2.Len()
		nl := make([]interface{}, l)
		for i := 0; i < l; i++ {
			nl[i] = l2.Index(i).Interface()
		}

		return append(nl, v), nil

	default:
		return nil, fmt.Errorf("Cannot push on type %s", tp)
	}
}

func prepend(list interface{}, v interface{}) []interface{} {
	l, err := mustPrepend(list, v)
	if err != nil {
		panic(err)
	}

	return l
}

func mustPrepend(list interface{}, v interface{}) ([]interface{}, error) {
	//return append([]interface{}{v}, list...)

	tp := reflect.TypeOf(list).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(list)

		l := l2.Len()
		nl := make([]interface{}, l)
		for i := 0; i < l; i++ {
			nl[i] = l2.Index(i).Interface()
		}

		return append([]interface{}{v}, nl...), nil

	default:
		return nil, fmt.Errorf("Cannot prepend on type %s", tp)
	}
}

func concat(lists ...interface{}) interface{} {
	var res []interface{}
	for _, list := range lists {
		tp := reflect.TypeOf(list).Kind()
		switch tp {
		case reflect.Slice, reflect.Array:
			l2 := reflect.ValueOf(list)
			for i := 0; i < l2.Len(); i++ {
				res = append(res, l2.Index(i).Interface())
			}
		default:
			panic(fmt.Sprintf("Cannot concat type %s as list", tp))
		}
	}
	return res
}
