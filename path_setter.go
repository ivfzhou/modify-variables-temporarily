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
	"reflect"
	"strconv"
	"unicode"
	"unsafe"
)

type pathSet struct {
	Substitute interface{}

	curKey  reflect.Value
	setFunc []func()
	restore []func()
}

func (ps *pathSet) set(pTarget reflect.Value, target reflect.Value, paths []string) []func() {
	if len(paths) == 0 {
		value := reflect.ValueOf(ps.Substitute)
		if pTarget.Kind() == reflect.Map {
			ps.restore = append(ps.restore, setMap(pTarget, ps.curKey, value))
		} else {
			ps.restore = append(ps.restore, setVal(target, value))
		}

		funcs := ps.setFunc
		length := len(funcs)
		for i := length - 1; i >= 0; i-- {
			funcs[i]()
		}

		return ps.restore
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
			index, _ := strconv.Atoi(path)
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

	case reflect.Ptr:
		elem := target.Elem()
		return ps.set(target, elem, paths)

	case reflect.Interface:
		value := target.Elem()
		elemKind := value.Kind()
		if elemKind != reflect.Ptr && elemKind != reflect.Map && elemKind != reflect.Slice && elemKind != reflect.Interface {
			newVal := reflect.New(value.Type()).Elem()
			newVal.Set(value)
			ps.setFunc = append(ps.setFunc, func() {
				target.Set(newVal)
			})
			ps.restore = append(ps.restore, func() {
				target.Set(value)
			})
			return ps.set(target, newVal, paths)
		}
		return ps.set(target, value, paths)

	case reflect.Map:
		var key reflect.Value
		if num, err := strconv.Atoi(path); err == nil {
			key = reflect.ValueOf(num)
		} else {
			key = reflect.ValueOf(path)
		}
		value := target.MapIndex(key)
		newVal := reflect.New(value.Type()).Elem()
		newVal.Set(value)
		ps.setFunc = append(ps.setFunc, func() {
			target.SetMapIndex(key, newVal)
		})
		ps.restore = append(ps.restore, func() {
			target.SetMapIndex(key, value)
		})
		ps.curKey = key
		return ps.set(target, newVal, paths[1:])

	case reflect.Array, reflect.Slice:
		index, err := strconv.Atoi(path)
		if err != nil {
			panic("path is unsatisfied: " + err.Error())
		}
		elem := target.Index(index)
		return ps.set(target, elem, paths[1:])

	default:
		panic("path 不合理")
	}
}
