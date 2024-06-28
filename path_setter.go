/*
 * Copyright (c) 2023 ivfzhou
 * modify-variables-temporarily is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

package modify_variables_temporarily

import (
	"fmt"
	"reflect"
	"strconv"
	"unicode"
	"unsafe"
)

type pathSetter struct {
	Substitute interface{}
	mapKey     reflect.Value
}

func (ps *pathSetter) set(pTarget reflect.Value, target reflect.Value, paths []string) func() {
	if len(paths) == 0 {
		value := reflect.ValueOf(ps.Substitute)
		if target.CanSet() {
			old := reflect.ValueOf(target.Interface())
			target.Set(value.Convert(target.Type()))
			return func() {
				target.Set(old)
			}
		}

		if pTarget.Kind() == reflect.Map {
			pTarget.SetMapIndex(ps.mapKey, value)
			return func() {
				pTarget.SetMapIndex(ps.mapKey, target)
			}
		}

		if target.Kind() == reflect.Ptr || target.Kind() == reflect.Interface {
			target = target.Elem()
			for target.Kind() == reflect.Ptr || target.Kind() == reflect.Interface {
				target = target.Elem()
			}
			if target.CanSet() {
				old := reflect.ValueOf(target.Interface())
				target.Set(value.Convert(target.Type()))
				return func() {
					target.Set(old)
				}
			}
		}

		panic(fmt.Sprintf("the target cannot set: [%s]", target.Type()))
	}

	path := paths[0]
	kind := target.Kind()

	switch kind {
	case reflect.Struct:
		var (
			field reflect.Value
			name  string
		)
		if unicode.IsDigit(rune(path[0])) {
			index, err := strconv.Atoi(path)
			if err != nil {
				panic(fmt.Sprintf("path is unsatisfied, deal with [%s] failure: %v", path, err))
			}
			field = target.Field(index)
			name = target.Type().Field(index).Name
		} else {
			field = target.FieldByName(path)
			name = path
		}
		if !unicode.IsUpper(rune(name[0])) {
			field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
		}

		return ps.set(target, field, paths[1:])

	case reflect.Ptr, reflect.Interface:
		elem := target.Elem()
		for elem.Kind() == reflect.Ptr || elem.Kind() == reflect.Interface {
			elem = elem.Elem()
		}
		return ps.set(target, elem, paths)

	case reflect.Map:
		var key reflect.Value
		if num, err := strconv.Atoi(path); err == nil {
			key = reflect.ValueOf(num)
		} else {
			key = reflect.ValueOf(path)
		}
		ps.mapKey = key
		value := target.MapIndex(key)
		return ps.set(target, value, paths[1:])

	case reflect.Array, reflect.Slice:
		index, err := strconv.Atoi(path)
		if err != nil {
			panic(fmt.Sprintf("path is unsatisfied, deal with [%s] failure: %v", path, err))
		}
		elem := target.Index(index)
		return ps.set(target, elem, paths[1:])

	default:
		panic(fmt.Sprintf("path is unsatisfied: %v", path))
	}
}
