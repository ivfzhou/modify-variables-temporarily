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
	"sync/atomic"
	"unicode"
	"unsafe"
)

// OutValue 函数返回值。
type OutValue []interface{}

// OutValues 多组函数返回值。
type OutValues []struct {
	Values OutValue
	// Times 每组返回值返回次数。
	Times int
}

func setVal(target reflect.Value, value reflect.Value) func() {
	old := reflect.ValueOf(target.Interface())
	target.Set(value)
	return func() {
		target.Set(old)
	}
}

func setFieldByName(target reflect.Value, name string, value reflect.Value) func() {
	field := target.FieldByName(name)
	if !field.IsValid() {
		panic(fmt.Sprintf("字段名错误 %s %T", name, target.Interface()))
	}
	if !unicode.IsUpper(rune(name[0])) {
		field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
	}
	old := reflect.ValueOf(field.Interface())
	field.Set(value)
	return func() {
		field.Set(old)
	}
}

func setField(target reflect.Value, index int, value reflect.Value) func() {
	field := target.Field(index)
	if !field.IsValid() {
		panic(fmt.Sprintf("字段序号错误 %d %T", index, target.Interface()))
	}
	fieldType := target.Type().Field(index)
	if !unicode.IsUpper(rune(fieldType.Name[0])) {
		field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
	}
	old := reflect.ValueOf(field.Interface())
	field.Set(value)
	return func() {
		field.Set(old)
	}
}

func setElem(target reflect.Value, index int, value reflect.Value) func() {
	elem := target.Index(index)
	old := reflect.ValueOf(elem.Interface())
	elem.Set(value)
	return func() {
		elem.Set(old)
	}
}

func setMap(target reflect.Value, key reflect.Value, value reflect.Value) func() {
	old := reflect.ValueOf(target.MapIndex(key).Interface())
	target.SetMapIndex(key, value)
	return func() {
		target.SetMapIndex(key, old)
	}
}

func setFuncOuts(target reflect.Value, outs OutValues) func() {
	old := reflect.ValueOf(target.Interface())
	target.Set(makeFunc(target, old, outs))
	return func() {
		target.Set(old)
	}
}

func makeFunc(target reflect.Value, old reflect.Value, outs OutValues) reflect.Value {
	outValues := generateOutValues(old, outs)
	length := int64(len(outValues))
	count := int64(-1)

	return reflect.MakeFunc(target.Type(), func(in []reflect.Value) []reflect.Value {
		index := atomic.AddInt64(&count, 1)
		if index < length {
			return outValues[index]
		} else {
			return old.Call(in)
		}
	})
}

func generateOutValues(fn reflect.Value, outs OutValues) [][]reflect.Value {
	res := make([][]reflect.Value, 0)

	for _, v := range outs {
		oneOut := make([]reflect.Value, 0, len(v.Values))
		for index, out := range v.Values {
			typ := fn.Type().Out(index)
			if out == nil {
				oneOut = append(oneOut, reflect.Zero(typ))
			} else {
				tmp := reflect.New(typ)
				tmp.Elem().Set(reflect.ValueOf(out))
				oneOut = append(oneOut, tmp.Elem())
			}
		}

		if v.Times <= 0 {
			v.Times = 1
		}

		for i := 0; i < v.Times; i++ {
			res = append(res, oneOut)
		}
	}

	return res
}
