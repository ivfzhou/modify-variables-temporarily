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
)

// Mvt 临时地修改变量值。
type Mvt struct {
	// restoreFuncs 还原所有替换的函数切片。
	restoreFuncs []func()
	lock         sync.Mutex
}

// New 创建一个 Mvt 对象。
func New() *Mvt {
	return &Mvt{}
}

// Var 替换单个变量。可替换的变量类型有：指针、接口。
func (mvt *Mvt) Var(target interface{}, substitute interface{}) *Mvt {
	mvt.lock.Lock()
	defer mvt.lock.Unlock()

	val, ok := indirect(target)
	if !ok {
		panic(fmt.Sprintf("the target cannot be set: [%T]", target))
	}

	mvt.restoreFuncs = append(mvt.restoreFuncs, setVal(val, reflect.ValueOf(substitute)))

	return mvt
}

// FieldByName 替换结构体字段的值。
func (mvt *Mvt) FieldByName(target interface{}, name string, substitute interface{}) *Mvt {
	mvt.lock.Lock()
	defer mvt.lock.Unlock()

	val, ok := indirect(target)
	if !ok {
		panic(fmt.Sprintf("the target cannot be set: [%T]", target))
	}
	if val.Kind() != reflect.Struct {
		panic(fmt.Sprintf("the target is not a struct: [%T]", target))
	}

	mvt.restoreFuncs = append(mvt.restoreFuncs, setFieldByName(val, name, reflect.ValueOf(substitute)))

	return mvt
}

// Field 替换结构体字段的值。
func (mvt *Mvt) Field(target interface{}, index uint, substitute interface{}) *Mvt {
	mvt.lock.Lock()
	defer mvt.lock.Unlock()

	val, ok := indirect(target)
	if !ok {
		panic(fmt.Sprintf("the target cannot be set: [%T]", target))
	}
	if val.Kind() != reflect.Struct {
		panic(fmt.Sprintf("the target is not a struct: [%T]", target))
	}

	mvt.restoreFuncs = append(mvt.restoreFuncs, setField(val, int(index), reflect.ValueOf(substitute)))

	return mvt
}

// Elem 替换切片的元素值。
func (mvt *Mvt) Elem(target interface{}, index uint, substitute interface{}) *Mvt {
	mvt.lock.Lock()
	defer mvt.lock.Unlock()

	val, _ := indirect(target)
	if val.Kind() != reflect.Slice {
		panic(fmt.Sprintf("the target cannot be set: [%T]", target))
	}

	mvt.restoreFuncs = append(mvt.restoreFuncs, setElem(val, int(index), reflect.ValueOf(substitute)))

	return mvt
}

// FuncOuts 替换函数变量以固定次数返回值代替。
func (mvt *Mvt) FuncOuts(target interface{}, outs OutValues) *Mvt {
	mvt.lock.Lock()
	defer mvt.lock.Unlock()

	val, ok := indirect(target)
	if !ok {
		panic(fmt.Sprintf("the target cannot be set: [%T]", target))
	}
	if val.Kind() != reflect.Func {
		panic(fmt.Sprintf("the target is not a function: [%T]", target))
	}

	mvt.restoreFuncs = append(mvt.restoreFuncs, setFuncOuts(val, outs))

	return mvt
}

// Map 替换映射中的某个键的值。
func (mvt *Mvt) Map(target interface{}, key interface{}, substitute interface{}) *Mvt {
	mvt.lock.Lock()
	defer mvt.lock.Unlock()

	val, _ := indirect(target)
	if val.Kind() != reflect.Map {
		panic(fmt.Sprintf("the target is not a map: [%T]", target))
	}

	mvt.restoreFuncs = append(mvt.restoreFuncs, setMap(val, reflect.ValueOf(key), reflect.ValueOf(substitute)))

	return mvt
}

// Path 根据索引替换深层值。
// path 为以点号分割的字符串。
func (mvt *Mvt) Path(target interface{}, path string, substitute interface{}) *Mvt {
	mvt.lock.Lock()
	defer mvt.lock.Unlock()

	setter := &pathSetter{Substitute: substitute}
	paths := strings.Split(path, ".")
	if len(paths) <= 0 {
		panic("then path cannot be empty")
	}
	for _, v := range paths {
		if len(v) <= 0 {
			panic("then path element cannot be empty")
		}
	}
	mvt.restoreFuncs = append(mvt.restoreFuncs, setter.set(reflect.Value{}, reflect.ValueOf(target), paths))

	return mvt
}

// Reset 还原该 Mvt 所有替换过的值。
func (mvt *Mvt) Reset() {
	for i := len(mvt.restoreFuncs) - 1; i >= 0; i-- {
		mvt.restoreFuncs[i]()
	}
}

func indirect(v interface{}) (reflect.Value, bool) {
	val := reflect.ValueOf(v)
	flag := false

F1:
	switch kind := val.Kind(); kind {
	case reflect.Ptr:
		val = reflect.Indirect(val)
		flag = true
		goto F1
	case reflect.Interface:
		val = val.Elem()
		flag = true
		goto F1
	default:
		return val, flag
	}
}
