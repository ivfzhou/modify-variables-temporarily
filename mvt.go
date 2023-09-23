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
	"strings"
	"sync"
	"unicode"
	"unsafe"
)

// mvt 临时地修改变量值。
type mvt struct {
	// target 当前被替换值的反射对象。
	target reflect.Value
	// restoreFuncs 还原所有替换的函数切片。
	restoreFuncs []func()
	lock         sync.Mutex
}

// NewWithTarget 创建一个 mvt 对象。
func NewWithTarget(target interface{}) *mvt {
	mvt := &mvt{}
	mvt.New(target)
	return mvt.New(target)
}

// New 创建一个 mvt 对象。
func New() *mvt {
	return &mvt{}
}

// New 重新标记一个目标。
func (mvt *mvt) New(target interface{}) *mvt {
	mvt.lock.Lock()
	defer mvt.lock.Unlock()

	val := reflect.ValueOf(target)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Map {
		nval, ok := isPtr(target)
		if !ok {
			panic("cannot set")
		}
		val = nval
	}
	mvt.target = val
	return mvt
}

// Var 替换单个变量。
func (mvt *mvt) Var(target interface{}, substitute interface{}) *mvt {
	mvt.lock.Lock()
	defer mvt.lock.Unlock()

	val, ok := isPtr(target)
	if !ok {
		panic(fmt.Sprintf("cannot set %T", target))
	}
	mvt.target = val

	mvt.restoreFuncs = append(mvt.restoreFuncs, setVal(mvt.target, reflect.ValueOf(substitute)))

	return mvt
}

// FieldByName 替换结构体字段的值。
func (mvt *mvt) FieldByName(target interface{}, name string, substitute interface{}) *mvt {
	mvt.lock.Lock()
	defer mvt.lock.Unlock()

	val, ok := isPtr(target)
	if !ok {
		panic(fmt.Sprintf("cannot set %T", target))
	}
	if mvt.target.Kind() != reflect.Struct {
		panic(fmt.Sprintf("target is not struct %T", target))
	}
	mvt.target = val

	mvt.restoreFuncs = append(mvt.restoreFuncs, setFieldByName(mvt.target, name, reflect.ValueOf(substitute)))

	return mvt
}

// Field 替换结构体字段的值。
func (mvt *mvt) Field(target interface{}, index uint, substitute interface{}) *mvt {
	mvt.lock.Lock()
	defer mvt.lock.Unlock()

	val, ok := isPtr(target)
	if !ok {
		panic(fmt.Sprintf("cannot set %T", target))
	}
	if mvt.target.Kind() != reflect.Struct {
		panic(fmt.Sprintf("target is not struct %T", target))
	}
	mvt.target = val

	mvt.restoreFuncs = append(mvt.restoreFuncs, setField(mvt.target, int(index), reflect.ValueOf(substitute)))

	return mvt
}

// Elem 替换切片的元素值。
func (mvt *mvt) Elem(target interface{}, index uint, substitute interface{}) *mvt {
	mvt.lock.Lock()
	defer mvt.lock.Unlock()

	val := reflect.ValueOf(target)
	kind := val.Kind()
	if kind != reflect.Slice {
		nval, ok := isPtr(target)
		if !ok {
			panic(fmt.Sprintf("cannot set %T", target))
		}
		if nval.Kind() != reflect.Slice {
			panic(fmt.Sprintf("target is not slice %T", target))
		}
		val = nval
	}
	mvt.target = val

	mvt.restoreFuncs = append(mvt.restoreFuncs, setElem(mvt.target, int(index), reflect.ValueOf(substitute)))

	return mvt
}

// FuncOuts 替换函数变量以固定次数返回值代替。
func (mvt *mvt) FuncOuts(target interface{}, outs OutValues) *mvt {
	mvt.lock.Lock()
	defer mvt.lock.Unlock()

	val, ok := isPtr(target)
	if !ok {
		panic(fmt.Sprintf("cannot set %T", target))
	}
	if mvt.target.Kind() != reflect.Func {
		panic(fmt.Sprintf("target is not func %T", target))
	}
	mvt.target = val

	mvt.restoreFuncs = append(mvt.restoreFuncs, setFuncOuts(mvt.target, outs))

	return mvt
}

// Map 替换映射中的某个键的值。
func (mvt *mvt) Map(target interface{}, key interface{}, substitute interface{}) *mvt {
	mvt.lock.Lock()
	defer mvt.lock.Unlock()

	val := reflect.ValueOf(target)
	if val.Kind() != reflect.Map {
		nval, ok := isPtr(target)
		if !ok {
			panic(fmt.Sprintf("cannot set %T", target))
		}
		if nval.Kind() != reflect.Map {
			panic(fmt.Sprintf("target is not map %T", target))
		}
		val = nval
	}
	mvt.target = val

	mvt.restoreFuncs = append(mvt.restoreFuncs, setMap(mvt.target, reflect.ValueOf(key), reflect.ValueOf(substitute)))

	return mvt
}

// Path 根据索引替换深层值。
// path 为以点号分割的字符串。
func (mvt *mvt) Path(target interface{}, path string, substitute interface{}) *mvt {
	mvt.lock.Lock()
	defer mvt.lock.Unlock()

	val := reflect.ValueOf(target)
	if val.Kind() == reflect.Array {
		if !isElemAddressable(target) {
			panic(fmt.Sprintf("cannot set %T", target))
		}
	} else if val.Kind() != reflect.Slice && val.Kind() != reflect.Map {
		nval, ok := isPtr(target)
		if !ok {
			panic(fmt.Sprintf("cannot set %T", target))
		}
		val = nval
	}
	mvt.target = val

	setter := &pathSet{Substitute: substitute}
	paths := strings.Split(path, ".")
	mvt.restoreFuncs = append(mvt.restoreFuncs, setter.set(reflect.Value{}, mvt.target, paths)...)

	return mvt
}

// PathByList 根据索引替换深层值。
func (mvt *mvt) PathByList(target interface{}, list []string, substitute interface{}) *mvt {
	mvt.lock.Lock()
	defer mvt.lock.Unlock()

	val := reflect.ValueOf(target)
	if val.Kind() == reflect.Array {
		if !isElemAddressable(target) {
			panic(fmt.Sprintf("cannot set %T", target))
		}
	}
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Map {
		nval, ok := isPtr(target)
		if !ok {
			panic(fmt.Sprintf("cannot set %T", target))
		}
		val = nval
	}
	mvt.target = val

	setter := &pathSet{Substitute: substitute}
	mvt.restoreFuncs = append(mvt.restoreFuncs, setter.set(reflect.Value{}, mvt.target, list)...)

	return mvt
}

// FieldNext 沿用上次替换的 target，含义同上。
func (mvt *mvt) FieldNext(index uint, substitute interface{}) *mvt {
	mvt.lock.Lock()
	defer mvt.lock.Unlock()

	if mvt.target.Kind() != reflect.Struct {
		panic("原 target 非结构体")
	}

	mvt.restoreFuncs = append(mvt.restoreFuncs, setField(mvt.target, int(index), reflect.ValueOf(substitute)))

	return mvt
}

// FieldByNameNext 沿用上次替换的 target，含义同上。
func (mvt *mvt) FieldByNameNext(name string, substitute interface{}) *mvt {
	mvt.lock.Lock()
	defer mvt.lock.Unlock()

	if mvt.target.Kind() != reflect.Struct {
		panic("原 target 非结构体")
	}

	mvt.restoreFuncs = append(mvt.restoreFuncs, setFieldByName(mvt.target, name, reflect.ValueOf(substitute)))

	return mvt
}

// ElemNext 沿用上次替换的 target，含义同上。
func (mvt *mvt) ElemNext(index int, substitute interface{}) *mvt {
	mvt.lock.Lock()
	defer mvt.lock.Unlock()

	kind := mvt.target.Kind()
	if kind != reflect.Slice {
		panic("原 target 非数组和切片")
	}

	mvt.restoreFuncs = append(mvt.restoreFuncs, setElem(mvt.target, index, reflect.ValueOf(substitute)))

	return mvt
}

// PathNext 沿用上次替换的 target，含义同上。
func (mvt *mvt) PathNext(path string, substitute interface{}) *mvt {
	mvt.lock.Lock()
	defer mvt.lock.Unlock()

	setter := &pathSet{Substitute: substitute}
	paths := strings.Split(path, ".")
	mvt.restoreFuncs = append(mvt.restoreFuncs, setter.set(reflect.Value{}, mvt.target, paths)...)

	return mvt
}

// MapNext 沿用上次替换的 target，含义同上。
func (mvt *mvt) MapNext(key interface{}, substitute interface{}) *mvt {
	mvt.lock.Lock()
	defer mvt.lock.Unlock()

	if mvt.target.Kind() != reflect.Map {
		panic("原 target 非映射")
	}

	sVal := reflect.ValueOf(substitute)
	keyVal := reflect.ValueOf(key)
	mvt.restoreFuncs = append(mvt.restoreFuncs, setMap(mvt.target, keyVal, sVal))

	return mvt
}

// FieldFuncOutsNext 沿用上次替换的 target，含义同上。
func (mvt *mvt) FieldFuncOutsNext(index uint, outs OutValues) *mvt {
	mvt.lock.Lock()
	defer mvt.lock.Unlock()

	if mvt.target.Kind() != reflect.Struct {
		panic("原 target 非结构体")
	}

	field := mvt.target.Field(int(index))
	fieldType := mvt.target.Type().Field(int(index))
	if !unicode.IsUpper(rune(fieldType.Name[0])) {
		field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
	}

	mvt.restoreFuncs = append(mvt.restoreFuncs, setFuncOuts(field, outs))

	return mvt
}

// FieldFuncOutsByNameNext 沿用上次替换的 target，含义同上。
func (mvt *mvt) FieldFuncOutsByNameNext(name string, outs OutValues) *mvt {
	mvt.lock.Lock()
	defer mvt.lock.Unlock()

	if mvt.target.Kind() != reflect.Struct {
		panic("原 target 非结构体")
	}

	field := mvt.target.FieldByName(name)
	if !unicode.IsUpper(rune(name[0])) {
		field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
	}

	mvt.restoreFuncs = append(mvt.restoreFuncs, setFuncOuts(field, outs))

	return mvt
}

// ToElem 修改 target 为切片元素。
func (mvt *mvt) ToElem(index uint) *mvt {
	mvt.lock.Lock()
	defer mvt.lock.Unlock()

	kind := mvt.target.Kind()
	if kind != reflect.Array && kind != reflect.Slice {
		panic("target 非数组和切片")
	}

	if !isElemAddressable(mvt.target) {
		panic("cannot set")
	}

	mvt.target = mvt.target.Index(int(index))

	return mvt
}

// ToField 修改 target 为字段值。
func (mvt *mvt) ToField(index uint) *mvt {
	mvt.lock.Lock()
	defer mvt.lock.Unlock()

	if mvt.target.Kind() != reflect.Struct {
		panic("target 非结构体")
	}

	field := mvt.target.Field(int(index))
	fieldType := mvt.target.Type().Field(int(index))
	if !unicode.IsUpper(rune(fieldType.Name[0])) {
		field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
	}
	if field.Kind() != reflect.Map && field.Kind() != reflect.Slice {
		val, ok := isPtr(field.Interface())
		if !ok {
			panic("cannot set")
		}
		field = val
	}
	mvt.target = field

	return mvt
}

// ToFieldByName 修改 target 为字段值。
func (mvt *mvt) ToFieldByName(name string) *mvt {
	mvt.lock.Lock()
	defer mvt.lock.Unlock()

	if mvt.target.Kind() != reflect.Struct {
		panic("target 非结构体")
	}

	field := mvt.target.FieldByName(name)
	if !unicode.IsUpper(rune(name[0])) {
		field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
	}
	if field.Kind() != reflect.Map && field.Kind() != reflect.Slice {
		val, ok := isPtr(field.Interface())
		if !ok {
			panic("cannot set")
		}
		field = val
	}
	mvt.target = field

	return mvt
}

// ToMapVal 修改 target 为 map 值。
func (mvt *mvt) ToMapVal(key interface{}) *mvt {
	mvt.lock.Lock()
	defer mvt.lock.Unlock()

	if mvt.target.Kind() != reflect.Map {
		panic("target 非映射")
	}

	val := mvt.target.MapIndex(reflect.ValueOf(key))
	if val.Kind() != reflect.Map && val.Kind() != reflect.Slice {
		nval, ok := isPtr(val.Interface())
		if !ok {
			panic("cannot set")
		}
		val = nval
	}
	mvt.target = val

	return mvt
}

// Reset 还原该 mvt 所有替换过的值。
func (mvt *mvt) Reset() {
	for i := len(mvt.restoreFuncs) - 1; i >= 0; i-- {
		mvt.restoreFuncs[i]()
	}
}

func isPtr(v interface{}) (reflect.Value, bool) {
	val := reflect.ValueOf(v)
	kind := val.Kind()
F1:
	switch kind {
	case reflect.Ptr:
		val = reflect.Indirect(val)
	case reflect.Interface:
		val = val.Elem()
		kind = val.Kind()
		goto F1
	default:
		return reflect.Value{}, false
	}
	return val, true
}

func isElemAddressable(v interface{}) bool {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Array || val.Kind() == reflect.Slice {
		valTyp := val.Type().Elem()
		kind := valTyp.Kind()
	F1:
		switch kind {
		case reflect.Ptr, reflect.Slice, reflect.Map:
			return true
		case reflect.Interface:
			val = val.Elem()
			kind = val.Kind()
			goto F1
		default:
			return false
		}
	}
	return false
}
