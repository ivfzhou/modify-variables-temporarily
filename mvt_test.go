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

package modify_variables_temporarily_test

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"sync"
	"testing"

	mvt "gitee.com/ivfzhou/modify-variables-temporarily/v3"
)

type testInterface interface {
	m() int
}

type testImpl int

type testImpl2 int

type testStruct struct {
	unexportedField  testInterface
	unexportedField2 testImpl
	unexportedField3 *testImpl
	unexportedField4 map[any]any
}

func TestVar(t *testing.T) {
	t.Run("Target 是 nil", func(t *testing.T) {
		for range 100 {
			func() {
				defer func() {
					recovered, _ := recover().(error)
					if !errors.Is(recovered, mvt.ErrTargetCannotBeNil) {
						t.Error("no ErrTargetCannotBeNil panic occurred", recovered)
					}
				}()
				var target any
				mvt.Var(target, 0)
			}()
		}
	})

	t.Run("Target 仅有类型但无值", func(t *testing.T) {
		for range 100 {
			func() {
				defer func() {
					recovered, _ := recover().(error)
					if !errors.Is(recovered, mvt.ErrTargetCannotBeNilType) {
						t.Error("no ErrTargetCannotBeNilType panic occurred", recovered)
					}
				}()
				var target *testStruct
				mvt.Var(target, testStruct{})
			}()
		}
	})

	t.Run("Target 不是指针", func(t *testing.T) {
		for range 100 {
			func() {
				defer func() {
					recovered, _ := recover().(error)
					if !errors.Is(recovered, mvt.ErrTargetIsNotPointer) {
						t.Error("no ErrTargetIsNotPointer panic occurred", recovered)
					}
				}()
				var target = testStruct{}
				mvt.Var(target, testStruct{})
			}()
		}
	})

	t.Run("Substitute 不兼容 Target", func(t *testing.T) {
		for range 100 {
			func() {
				defer func() {
					recovered, _ := recover().(error)
					if !errors.Is(recovered, mvt.ErrIncompatibleTypeAssignment) {
						t.Error("no ErrIncompatibleTypeAssignment panic occurred", recovered)
					}
				}()
				var target int
				mvt.Var(&target, "")
			}()
		}
	})

	t.Run("Substitute 被强转赋值给 Target", func(t *testing.T) {
		for range 100 {
			originalValue := rand.Intn(1000)
			var target = testImpl(originalValue)
			newValue := rand.Intn(1000)
			reset := mvt.Var(&target, newValue)
			if int(target) != newValue {
				t.Error("target does not meet expectation", target, newValue)
			}
			reset.Reset()
			if int(target) != originalValue {
				t.Error("target does not meet expectation", target, originalValue)
			}
		}
	})

	t.Run("正常运行", func(t *testing.T) {
		for range 100 {
			originalValue := rand.Intn(1000)
			var target testInterface = testImpl(originalValue)
			newValue := rand.Intn(1000)
			reset := mvt.Var(&target, testImpl2(newValue))
			if target.m() != newValue {
				t.Error("target does not meet expectation", target, newValue)
			}
			reset.Reset()
			if target.m() != originalValue {
				t.Error("target does not meet expectation", target, originalValue)
			}
		}
	})

	t.Run("Target 替换成零值", func(t *testing.T) {
		for range 100 {
			originalValue := rand.Intn(1000)
			var target testInterface = testImpl(originalValue)
			newValue := testInterface(nil)
			reset := mvt.Var(&target, newValue)
			if target != nil {
				t.Error("target does not meet expectation", target)
			}
			reset.Reset()
			if target.m() != originalValue {
				t.Error("target does not meet expectation", target, originalValue)
			}
		}
	})

	t.Run("零值 Target 替换成其它值", func(t *testing.T) {
		for range 100 {
			var target testInterface
			newValue := rand.Intn(1000)
			reset := mvt.Var(&target, testImpl(newValue))
			if target.m() != newValue {
				t.Error("target does not meet expectation", target, newValue)
			}
			reset.Reset()
			if target != nil {
				t.Error("target does not meet expectation", target)
			}
		}
	})
}

func TestFieldByName(t *testing.T) {
	t.Run("Target 是 nil", func(t *testing.T) {
		for range 100 {
			func() {
				defer func() {
					recovered, _ := recover().(error)
					if !errors.Is(recovered, mvt.ErrTargetCannotBeNil) {
						t.Error("no ErrTargetCannotBeNil panic occurred", recovered)
					}
				}()
				var target any
				mvt.FieldByName(target, "unexportedField", testImpl(0))
			}()
		}
	})

	t.Run("Target 仅有类型但无值", func(t *testing.T) {
		for range 100 {
			func() {
				defer func() {
					recovered, _ := recover().(error)
					if !errors.Is(recovered, mvt.ErrTargetCannotBeNilType) {
						t.Error("no ErrTargetCannotBeNilType panic occurred", recovered)
					}
				}()
				var target *testStruct
				mvt.FieldByName(target, "unexportedField", testImpl(0))
			}()
		}
	})

	t.Run("name 是空串", func(t *testing.T) {
		for range 100 {
			func() {
				defer func() {
					recovered, _ := recover().(error)
					if !errors.Is(recovered, mvt.ErrStructFieldNameCannotBeEmpty) {
						t.Error("no ErrStructFieldNameCannotBeEmpty panic occurred", recovered)
					}
				}()
				var target testStruct
				mvt.FieldByName(&target, "", testImpl(0))
			}()
		}
	})

	t.Run("结构体不存在该 name 字段", func(t *testing.T) {
		for range 100 {
			func() {
				defer func() {
					recovered, _ := recover().(error)
					if !errors.Is(recovered, mvt.ErrStructFieldNotFound) {
						t.Error("no ErrStructFieldNotFound panic occurred", recovered)
					}
				}()
				var target testStruct
				mvt.FieldByName(&target, " ", testImpl(0))
			}()
		}
	})

	t.Run("Target 不是结构体", func(t *testing.T) {
		for range 100 {
			func() {
				defer func() {
					recovered, _ := recover().(error)
					if !errors.Is(recovered, mvt.ErrTargetIsNotStruct) {
						t.Error("no ErrTargetIsNotStruct panic occurred", recovered)
					}
				}()
				var target = ""
				mvt.FieldByName(&target, "unexportedField", testImpl(0))
			}()
		}
	})

	t.Run("Target 不是指针", func(t *testing.T) {
		for range 100 {
			func() {
				defer func() {
					recovered, _ := recover().(error)
					if !errors.Is(recovered, mvt.ErrTargetIsNotPointer) {
						t.Error("no ErrTargetIsNotPointer panic occurred", recovered)
					}
				}()
				var target = testStruct{}
				mvt.FieldByName(target, "unexportedField", testImpl(0))
			}()
		}
	})

	t.Run("Substitute 类型不兼容 Target 字段", func(t *testing.T) {
		for range 100 {
			func() {
				defer func() {
					recovered, _ := recover().(error)
					if !errors.Is(recovered, mvt.ErrIncompatibleTypeAssignment) {
						t.Error("no ErrIncompatibleTypeAssignment panic occurred", recovered)
					}
				}()
				var target testStruct
				mvt.FieldByName(&target, "unexportedField", 0)
			}()
		}
	})

	t.Run("Substitute 强转成字段类型", func(t *testing.T) {
		for range 100 {
			originalValue := rand.Intn(1000)
			target := &testStruct{
				unexportedField2: testImpl(originalValue),
			}
			newValue := rand.Intn(1000)
			reset := mvt.FieldByName(target, "unexportedField2", newValue)
			if int(target.unexportedField2) != newValue {
				t.Error("field of target does not meet expectation", target, newValue)
			}
			reset.Reset()
			if int(target.unexportedField2) != originalValue {
				t.Error("field of target does not meet expectation", target, originalValue)
			}
		}
	})

	t.Run("Target 字段替换成零值", func(t *testing.T) {
		for range 100 {
			originalValue := rand.Intn(1000)
			target := &testStruct{
				unexportedField: testImpl(originalValue),
			}
			newValue := testInterface(nil)
			reset := mvt.FieldByName(target, "unexportedField", newValue)
			if target.unexportedField != nil {
				t.Error("field of target does not meet expectation", target)
			}
			reset.Reset()
			if target.unexportedField.m() != originalValue {
				t.Error("field of target does not meet expectation", target, originalValue)
			}
		}
	})

	t.Run("零值 Target 字段替换成其它值", func(t *testing.T) {
		for range 100 {
			target := &testStruct{
				unexportedField: nil,
			}
			newValue := rand.Intn(1000)
			reset := mvt.FieldByName(target, "unexportedField", testImpl(newValue))
			if target.unexportedField.m() != newValue {
				t.Error("field of target does not meet expectation", target, newValue)
			}
			reset.Reset()
			if target.unexportedField != nil {
				t.Error("field of target does not meet expectation", target)
			}
		}
	})
}

func TestField(t *testing.T) {
	t.Run("Target 是 nil", func(t *testing.T) {
		for range 100 {
			func() {
				defer func() {
					recovered, _ := recover().(error)
					if !errors.Is(recovered, mvt.ErrTargetCannotBeNil) {
						t.Error("no ErrTargetCannotBeNil panic occurred", recovered)
					}
				}()
				var target any
				mvt.Field(target, 0, testImpl(0))
			}()
		}
	})

	t.Run("Target 有类型但无值", func(t *testing.T) {
		for range 100 {
			func() {
				defer func() {
					recovered, _ := recover().(error)
					if !errors.Is(recovered, mvt.ErrTargetCannotBeNilType) {
						t.Error("no ErrTargetCannotBeNilType panic occurred", recovered)
					}
				}()
				var target *testStruct
				mvt.Field(target, 0, testImpl(0))
			}()
		}
	})

	t.Run("index 越界", func(t *testing.T) {
		numField := reflect.TypeOf(testStruct{}).NumField()
		for range 100 {
			func() {
				defer func() {
					recovered, _ := recover().(error)
					if !errors.Is(recovered, mvt.ErrStructFieldNotFound) {
						t.Error("no ErrStructFieldNotFound panic occurred", recovered)
					}
				}()
				index := numField
				if rand.Intn(2) > 0 {
					index = -index - 1
				}
				var target testStruct
				mvt.Field(&target, index, testImpl(0))
			}()
		}
	})

	t.Run("Target 不是结构体", func(t *testing.T) {
		for range 100 {
			func() {
				defer func() {
					recovered, _ := recover().(error)
					if !errors.Is(recovered, mvt.ErrTargetIsNotStruct) {
						t.Error("no ErrTargetIsNotStruct panic occurred", recovered)
					}
				}()
				var target any = ""
				mvt.Field(&target, 0, testImpl(0))
			}()
		}
	})

	t.Run("Target 不是指针", func(t *testing.T) {
		for range 100 {
			func() {
				defer func() {
					recovered, _ := recover().(error)
					if !errors.Is(recovered, mvt.ErrTargetIsNotPointer) {
						t.Error("no ErrTargetIsNotPointer panic occurred", recovered)
					}
				}()
				var target = testStruct{}
				mvt.Field(target, 0, testImpl(0))
			}()
		}
	})

	t.Run("Substitute 不兼容 Target 的字段类型", func(t *testing.T) {
		for range 100 {
			func() {
				defer func() {
					recovered, _ := recover().(error)
					if !errors.Is(recovered, mvt.ErrIncompatibleTypeAssignment) {
						t.Error("no ErrIncompatibleTypeAssignment panic occurred", recovered)
					}
				}()
				var target testStruct
				mvt.Field(&target, 0, "")
			}()
		}
	})

	t.Run("Substitute 被强转成 Target 字段类型", func(t *testing.T) {
		numField := reflect.TypeOf(testStruct{}).NumField()
		for range 100 {
			originalValue := rand.Intn(1000)
			target := &testStruct{
				unexportedField2: testImpl(originalValue),
			}
			newValue := rand.Intn(1000)
			index := 1
			if rand.Intn(2) > 0 {
				index = -numField + index
			}
			reset := mvt.Field(target, index, newValue)
			if target.unexportedField2.m() != newValue {
				t.Error("field of target does not meet expectation", target, newValue)
			}
			reset.Reset()
			if target.unexportedField2.m() != originalValue {
				t.Error("field of target does not meet expectation", target, originalValue)
			}
		}
	})

	t.Run("正常运行", func(t *testing.T) {
		numField := reflect.TypeOf(testStruct{}).NumField()
		for range 100 {
			originalValue := rand.Intn(1000)
			target := &testStruct{
				unexportedField: testImpl(originalValue),
			}
			index := 0
			if rand.Intn(2) > 0 {
				index = -numField + index
			}
			newValue := rand.Intn(1000)
			reset := mvt.Field(target, index, testImpl2(newValue))
			if target.unexportedField.m() != newValue {
				t.Error("field of target does not meet expectation", target, newValue)
			}
			reset.Reset()
			if target.unexportedField.m() != originalValue {
				t.Error("field of target does not meet expectation", target, originalValue)
			}
		}
	})

	t.Run("Target 字段替换成零值", func(t *testing.T) {
		numField := reflect.TypeOf(testStruct{}).NumField()
		for range 100 {
			originalValue := rand.Intn(1000)
			target := &testStruct{
				unexportedField: testImpl(originalValue),
			}
			index := 0
			if rand.Intn(2) > 0 {
				index = -numField + index
			}
			newValue := testInterface(nil)
			reset := mvt.Field(target, index, newValue)
			if target.unexportedField != nil {
				t.Error("field of target does not meet expectation", target)
			}
			reset.Reset()
			if target.unexportedField.m() != originalValue {
				t.Error("field of target does not meet expectation", target, originalValue)
			}
		}
	})

	t.Run("零值 Target 字段替换成其它值", func(t *testing.T) {
		numField := reflect.TypeOf(testStruct{}).NumField()
		for range 100 {
			target := &testStruct{
				unexportedField: nil,
			}
			newValue := rand.Intn(1000)
			index := 0
			if rand.Intn(2) > 0 {
				index = -numField + index
			}
			reset := mvt.Field(target, index, testImpl(newValue))
			if target.unexportedField.m() != newValue {
				t.Error("field of target does not meet expectation", target, newValue)
			}
			reset.Reset()
			if target.unexportedField != nil {
				t.Error("field of target does not meet expectation", target)
			}
		}
	})
}

func TestElem(t *testing.T) {
	t.Run("Target 是 nil", func(t *testing.T) {
		for range 100 {
			func() {
				defer func() {
					recovered, _ := recover().(error)
					if !errors.Is(recovered, mvt.ErrTargetCannotBeNil) {
						t.Error("no ErrTargetCannotBeNil panic occurred", recovered)
					}
				}()
				var target any
				mvt.Elem(target, 0, testImpl(0))
			}()
		}
	})

	t.Run("Target 有类型但无值", func(t *testing.T) {
		for range 100 {
			func() {
				defer func() {
					recovered, _ := recover().(error)
					if !errors.Is(recovered, mvt.ErrTargetCannotBeNilType) {
						t.Error("no ErrTargetCannotBeNilType panic occurred", recovered)
					}
				}()
				var target []testInterface
				mvt.Elem(target, 0, testImpl(0))
			}()
		}
	})

	t.Run("index 越界", func(t *testing.T) {
		for range 100 {
			func() {
				defer func() {
					recovered, _ := recover().(error)
					if !errors.Is(recovered, mvt.ErrIndexOutOfBound) {
						t.Error("no ErrIndexOutOfBound panic occurred", recovered)
					}
				}()
				var target = []testInterface{nil}
				index := len(target)
				if rand.Intn(2) > 0 {
					index = -index - 1
				}
				mvt.Elem(target, index, testImpl(0))
			}()
		}
	})

	t.Run("Target 不是切片类型", func(t *testing.T) {
		for range 100 {
			func() {
				defer func() {
					recovered, _ := recover().(error)
					if !errors.Is(recovered, mvt.ErrTargetIsNotSlice) {
						t.Error("no ErrTargetIsNotSlice panic occurred", recovered)
					}
				}()
				var target any = ""
				mvt.Elem(target, 0, "")
			}()
		}
	})

	t.Run("Substitute 不兼容 Target 元素类型", func(t *testing.T) {
		for range 100 {
			func() {
				defer func() {
					recovered, _ := recover().(error)
					if !errors.Is(recovered, mvt.ErrIncompatibleTypeAssignment) {
						t.Error("no ErrIncompatibleTypeAssignment panic occurred", recovered)
					}
				}()
				var target = []testInterface{nil}
				mvt.Elem(target, 0, "")
			}()
		}
	})

	t.Run("Substitute 被强转成 Target 元素的类型", func(t *testing.T) {
		for range 100 {
			originalValue := rand.Intn(1000)
			target := []testImpl{testImpl(originalValue)}
			newValue := rand.Intn(1000)
			index := len(target) - 1
			index2 := index
			if rand.Intn(2) > 0 {
				index = -len(target) + index
			}
			reset := mvt.Elem(target, index, newValue)
			if target[index2].m() != newValue {
				t.Error("elem of target does not meet expectation", target, newValue)
			}
			reset.Reset()
			if target[index2].m() != originalValue {
				t.Error("elem of target does not meet expectation", target, originalValue)
			}
		}
	})

	t.Run("正常运行", func(t *testing.T) {
		for range 100 {
			originalValue := rand.Intn(1000)
			target := []testInterface{testImpl(originalValue)}
			index := len(target) - 1
			index2 := index
			if rand.Intn(2) > 0 {
				index = -len(target) + index
			}
			newValue := rand.Intn(1000)
			reset := mvt.Elem(target, index, testImpl2(newValue))
			if target[index2].m() != newValue {
				t.Error("elem of target does not meet expectation", target, newValue)
			}
			reset.Reset()
			if target[index2].m() != originalValue {
				t.Error("elem of target does not meet expectation", target, originalValue)
			}
		}
	})

	t.Run("Target 元素替换成零值", func(t *testing.T) {
		for range 100 {
			originalValue := rand.Intn(1000)
			target := []testInterface{testImpl(originalValue)}
			newValue := testInterface(nil)
			index := len(target) - 1
			index2 := index
			if rand.Intn(2) > 0 {
				index = -len(target) + index
			}
			reset := mvt.Elem(target, index, newValue)
			if target[index2] != nil {
				t.Error("elem of target does not meet expectation", target)
			}
			reset.Reset()
			if target[index2].m() != originalValue {
				t.Error("elem of target does not meet expectation", target, originalValue)
			}
		}
	})

	t.Run("零值 Target 元素替换成其它值", func(t *testing.T) {
		for range 100 {
			target := []testInterface{nil}
			newValue := rand.Intn(1000)
			index := len(target) - 1
			index2 := index
			if rand.Intn(2) > 0 {
				index = -len(target) + index
			}
			reset := mvt.Elem(target, index, testImpl(newValue))
			if target[index2].m() != newValue {
				t.Error("elem of target does not meet expectation", target, newValue)
			}
			reset.Reset()
			if target[index2] != nil {
				t.Error("elem of target does not meet expectation", target)
			}
		}
	})
}

func TestMap(t *testing.T) {
	t.Run("Target 是 nil", func(t *testing.T) {
		for range 100 {
			func() {
				defer func() {
					recovered, _ := recover().(error)
					if !errors.Is(recovered, mvt.ErrTargetCannotBeNil) {
						t.Error("no ErrTargetCannotBeNil panic occurred", recovered)
					}
				}()
				var target any
				mvt.Map(target, 0, nil)
			}()
		}
	})

	t.Run("Target 有类型但无值", func(t *testing.T) {
		for range 100 {
			func() {
				defer func() {
					recovered, _ := recover().(error)
					if !errors.Is(recovered, mvt.ErrTargetCannotBeNilType) {
						t.Error("no ErrTargetCannotBeNilType panic occurred", recovered)
					}
				}()
				var target map[any]any
				mvt.Map(target, 0, nil)
			}()
		}
	})

	t.Run("Target 不是 map", func(t *testing.T) {
		for range 100 {
			func() {
				defer func() {
					recovered, _ := recover().(error)
					if !errors.Is(recovered, mvt.ErrTargetIsNotMap) {
						t.Error("no ErrTargetIsNotMap panic occurred", recovered)
					}
				}()
				var fn int
				mvt.Map(fn, 0, nil)
			}()
		}
	})

	t.Run("key 是 nil", func(t *testing.T) {
		for range 100 {
			target := make(map[any]any)
			reset := mvt.Map(target, nil, 0)
			v, ok := target[nil]
			if !ok || v != 0 {
				t.Error("target does not meet expectation", ok, v)
			}
			reset.Reset()
			_, ok = target[nil]
			if ok {
				t.Error("target does not meet expectation", ok)
			}
		}
	})

	t.Run("key 类型不兼容", func(t *testing.T) {
		for range 100 {
			func() {
				defer func() {
					recovered, _ := recover().(error)
					if !errors.Is(recovered, mvt.ErrInvalidMapKeyType) {
						t.Error("no ErrInvalidMapKeyType panic occurred", recovered)
					}
				}()
				target := make(map[testInterface]any)
				mvt.Map(target, "", 0)
			}()
		}
	})

	t.Run("key 被强转", func(t *testing.T) {
		for range 100 {
			target := make(map[testImpl]any)
			reset := mvt.Map(target, 1, 0)
			v, ok := target[testImpl(1)]
			if !ok || v != 0 {
				t.Error("target does not meet expectation", ok, v)
			}
			reset.Reset()
			_, ok = target[testImpl(1)]
			if ok {
				t.Error("target does not meet expectation", ok)
			}
		}
	})

	t.Run("value 是 nil", func(t *testing.T) {
		for range 100 {
			target := make(map[any]any)
			reset := mvt.Map(target, 0, nil)
			v, ok := target[0]
			if !ok || v != nil {
				t.Error("target does not meet expectation", ok, v)
			}
			reset.Reset()
			_, ok = target[nil]
			if ok {
				t.Error("target does not meet expectation", ok)
			}
		}
	})

	t.Run("value 类型不兼容", func(t *testing.T) {
		for range 100 {
			func() {
				defer func() {
					recovered, _ := recover().(error)
					if !errors.Is(recovered, mvt.ErrIncompatibleTypeAssignment) {
						t.Error("no ErrIncompatibleTypeAssignment panic occurred", recovered)
					}
				}()
				target := make(map[any]testInterface)
				mvt.Map(target, 0, 0)
			}()
		}
	})

	t.Run("value 被强转", func(t *testing.T) {
		for range 100 {
			target := make(map[any]testImpl)
			originalValue := rand.Intn(1000)
			target[0] = testImpl(originalValue)
			newValue := rand.Intn(1000)
			reset := mvt.Map(target, 0, newValue)
			v, ok := target[0]
			if !ok || v.m() != newValue {
				t.Error("target does not meet expectation", target, newValue)
			}
			reset.Reset()
			v, ok = target[0]
			if !ok || v.m() != originalValue {
				t.Error("target does not meet expectation", target, originalValue)
			}
		}
	})

	t.Run("正常运行", func(t *testing.T) {
		for range 100 {
			target := make(map[any]testInterface)
			originalValue := rand.Intn(1000)
			target[0] = testImpl(originalValue)
			newValue := rand.Intn(1000)
			reset := mvt.Map(target, 0, testImpl2(newValue))
			v, ok := target[0]
			if !ok || v.m() != newValue {
				t.Error("target does not meet expectation", target, newValue)
			}
			reset.Reset()
			v, ok = target[0]
			if !ok || v.m() != originalValue {
				t.Error("target does not meet expectation", target, originalValue)
			}
		}
	})
}

func TestFuncOuts(t *testing.T) {
	t.Run("Target 是 nil", func(t *testing.T) {
		for range 100 {
			func() {
				defer func() {
					recovered, _ := recover().(error)
					if !errors.Is(recovered, mvt.ErrTargetCannotBeNil) {
						t.Error("no ErrTargetCannotBeNil panic occurred", recovered)
					}
				}()
				var target any
				mvt.FuncOuts(target, []mvt.OutValue{})
			}()
		}
	})

	t.Run("Target 不是指针", func(t *testing.T) {
		for range 100 {
			func() {
				defer func() {
					recovered, _ := recover().(error)
					if !errors.Is(recovered, mvt.ErrTargetIsNotPointer) {
						t.Error("no ErrTargetIsNotPointer panic occurred", recovered)
					}
				}()
				var target = func() {}
				mvt.FuncOuts(target, []mvt.OutValue{})
			}()
		}
	})

	t.Run("Target 有类型但无值", func(t *testing.T) {
		for range 100 {
			func() {
				defer func() {
					recovered, _ := recover().(error)
					if !errors.Is(recovered, mvt.ErrTargetIsNotFunc) {
						t.Error("no ErrTargetIsNotFunc panic occurred", recovered)
					}
				}()
				var target *func()
				mvt.FuncOuts(target, []mvt.OutValue{})
			}()
		}
	})

	t.Run("Target 不是函数", func(t *testing.T) {
		for range 100 {
			func() {
				defer func() {
					recovered, _ := recover().(error)
					if !errors.Is(recovered, mvt.ErrTargetIsNotFunc) {
						t.Error("no ErrTargetIsNotFunc panic occurred", recovered)
					}
				}()
				var target any = ""
				mvt.FuncOuts(&target, []mvt.OutValue{})
			}()
		}
	})

	t.Run("正常运行", func(t *testing.T) {
		fn := func() (int, int) { return 0, 0 }
		for range 100 {
			outValues := []mvt.OutValue{
				{
					Values: []any{rand.Intn(1000), rand.Intn(1000)},
					Times:  2,
				},
				{
					Values: []any{rand.Intn(1000), rand.Intn(1000)},
				},
			}
			reset := mvt.FuncOuts(&fn, outValues)
			v, v2 := fn()
			if v != outValues[0].Values[0] || v2 != outValues[0].Values[1] {
				t.Error("values func returned does not meet expectation", v, v2, outValues[0].Values[0], outValues[0].Values[1])
			}
			v, v2 = fn()
			if v != outValues[0].Values[0] || v2 != outValues[0].Values[1] {
				t.Error("values func returned does not meet expectation", v, v2, outValues[0].Values[0], outValues[0].Values[1])
			}
			v, v2 = fn()
			if v != outValues[1].Values[0] || v2 != outValues[1].Values[1] {
				t.Error("values func returned does not meet expectation", v, v2, outValues[0].Values[0], outValues[0].Values[1])
			}
			v, v2 = fn()
			if v != 0 || v2 != 0 {
				t.Error("values func returned does not meet expectation", v, v2)
			}
			reset.Reset()
			v, v2 = fn()
			if v != 0 || v2 != 0 {
				t.Error("values func returned does not meet expectation", v, v2)
			}
		}
	})

	t.Run("outs 被强转", func(t *testing.T) {
		fn := func() (int, testImpl) { return 0, 0 }
		for range 100 {
			outValues := []mvt.OutValue{
				{
					Values: []any{rand.Intn(1000), rand.Intn(1000)},
				},
			}
			reset := mvt.FuncOuts(&fn, outValues)
			v, v2 := fn()
			if v != outValues[0].Values[0] || v2.m() != outValues[0].Values[1] {
				t.Error("values func returned does not meet expectation", v, v2, outValues[0].Values[0], outValues[0].Values[1])
			}
			reset.Reset()
			v, v2 = fn()
			if v != 0 || v2.m() != 0 {
				t.Error("values func returned does not meet expectation", v, v2)
			}
		}
	})

	t.Run("outs 有多余", func(t *testing.T) {
		fn := func() (int, int) { return 0, 0 }
		for range 100 {
			outValues := []mvt.OutValue{
				{
					Values: []any{rand.Intn(1000), rand.Intn(1000), rand.Intn(1000)},
				},
			}
			reset := mvt.FuncOuts(&fn, outValues)
			v, v2 := fn()
			if v != outValues[0].Values[0] || v2 != outValues[0].Values[1] {
				t.Error("values func returned does not meet expectation", v, v2, outValues[0].Values[0], outValues[0].Values[1])
			}
			reset.Reset()
			v, v2 = fn()
			if v != 0 || v2 != 0 {
				t.Error("values func returned does not meet expectation", v, v2)
			}
		}
	})

	t.Run("outs 少了", func(t *testing.T) {
		fn := func() (int, int) { return 0, 0 }
		for range 100 {
			outValues := []mvt.OutValue{
				{
					Values: []any{rand.Intn(1000)},
				},
			}
			reset := mvt.FuncOuts(&fn, outValues)
			v, v2 := fn()
			if v != outValues[0].Values[0] || v2 != 0 {
				t.Error("values func returned does not meet expectation", v, v2, outValues[0].Values[0], outValues[0].Values[1])
			}
			reset.Reset()
			v, v2 = fn()
			if v != 0 || v2 != 0 {
				t.Error("values func returned does not meet expectation", v, v2)
			}
		}
	})

	t.Run("outs 中有 nil", func(t *testing.T) {
		fn := func() (int, int) { return 1, 1 }
		for range 100 {
			outValues := []mvt.OutValue{
				{
					Values: []any{nil, nil},
				},
			}
			reset := mvt.FuncOuts(&fn, outValues)
			v, v2 := fn()
			if v != 0 || v2 != 0 {
				t.Error("values func returned does not meet expectation", v, v2, outValues[0].Values[0], outValues[0].Values[1])
			}
			reset.Reset()
			v, v2 = fn()
			if v != 1 || v2 != 1 {
				t.Error("values func returned does not meet expectation", v, v2)
			}
		}
	})

	t.Run("outs 类型不兼容", func(t *testing.T) {
		fn := func() testInterface { return nil }
		for range 100 {
			func() {
				defer func() {
					recovered, _ := recover().(error)
					if !errors.Is(recovered, mvt.ErrIncompatibleTypeAssignment) {
						t.Error("no ErrIncompatibleTypeAssignment panic occurred", recovered)
					}
				}()
				outValues := []mvt.OutValue{
					{
						Values: []any{rand.Intn(1000)},
					},
				}
				mvt.FuncOuts(&fn, outValues)
			}()
		}
	})

	t.Run("并发运行", func(t *testing.T) {
		fn := func() (int, int) { return 0, 0 }
		for range 100 {
			resultValues := make(map[string]int, 100)
			outValues := make([]mvt.OutValue, 100)
			runTimes := 0
			for j := range outValues {
				v := rand.Intn(1000)
				v2 := rand.Intn(1000)
				key := fmt.Sprintf("%d_%d", v, v2)
				times := rand.Intn(5)
				tmp := times
				if tmp <= 0 {
					tmp = 1
				}
				runTimes += tmp
				_, ok := resultValues[key]
				if ok {
					resultValues[key] += tmp
				} else {
					resultValues[key] = tmp
				}
				outValues[j] = mvt.OutValue{
					Values: []any{v, v2},
					Times:  times,
				}
			}
			reset := mvt.FuncOuts(&fn, outValues)
			wg := sync.WaitGroup{}
			originalReturned := 10
			wg.Add(runTimes)
			resultChan := make(chan []int, 100)
			for range runTimes {
				go func() {
					v, v2 := fn()
					go func() {
						defer wg.Done()
						resultChan <- []int{v, v2}
					}()
				}()
			}
			go func() {
				wg.Wait()
				close(resultChan)
			}()
			for v := range resultChan {
				key := fmt.Sprintf("%d_%d", v[0], v[1])
				if times, ok := resultValues[key]; !ok {
					if v[0] == 0 && v[1] == 0 {
						if originalReturned <= 0 {
							t.Error("values func returned does not meet expectation", v[0], v[1])
						} else {
							originalReturned--
						}
					}
				} else {
					if times <= 0 {
						t.Error("values func returned does not meet expectation", v[0], v[1])
					} else {
						resultValues[key]--
					}
				}
			}
			for _, v := range resultValues {
				if v != 0 {
					t.Error("values func returned does not meet expectation", v)
				}
			}
			reset.Reset()
			v, v2 := fn()
			if v != 0 || v2 != 0 {
				t.Error("values func returned does not meet expectation", v, v2)
			}
		}
	})
}

func TestChain(t *testing.T) {
	t.Run("测试指针 1", func(t *testing.T) {
		for range 100 {
			var target any
			value := rand.Intn(1000)
			reset := mvt.Chain(&target).Elem().Set(value)
			if target != value {
				t.Log("target value does not meet expectation", target, value)
			}
			reset.Reset()
			if target != nil {
				t.Error("target value does not meet expectation", target)
			}
		}
	})

	t.Run("测试指针 2", func(t *testing.T) {
		for range 100 {
			var target *any
			value := rand.Intn(1000)
			reset := mvt.Chain(&target).Elem().Elem().Set(value)
			if *target != value {
				t.Log("target value does not meet expectation", target, value)
			}
			reset.Reset()
			if target != nil {
				t.Error("target value does not meet expectation", target)
			}
		}
	})

	t.Run("测试指针 3", func(t *testing.T) {
		for range 100 {
			var target *map[any]any
			key := rand.Intn(1000)
			value := map[any]any{key: rand.Intn(1000)}
			reset := mvt.Chain(&target).Elem().Elem().Set(value)
			if (*target)[key] != value[key] {
				t.Log("target value does not meet expectation", target, value)
			}
			reset.Reset()
			if target != nil {
				t.Error("target value does not meet expectation", target)
			}
		}
	})

	t.Run("测试指针 4", func(t *testing.T) {
		for range 100 {
			var target *testInterface
			value := testImpl(rand.Intn(1000))
			reset := mvt.Chain(&target).Elem().Elem().Set(value)
			if *target != value {
				t.Log("target value does not meet expectation", target, value)
			}
			reset.Reset()
			if target != nil {
				t.Error("target value does not meet expectation", target)
			}
		}
	})

	t.Run("测试指针 5", func(t *testing.T) {
		for range 100 {
			var impl testInterface = testImpl(rand.Intn(1000))
			var target *testInterface = &impl
			value := testImpl(rand.Intn(1000))
			reset := mvt.Chain(&target).Elem().Elem().Elem().Set(value)
			if *target != value {
				t.Log("target value does not meet expectation", target, value)
			}
			reset.Reset()
			if *target != impl {
				t.Error("target value does not meet expectation", target)
			}
		}
	})

	t.Run("测试结构体 1", func(t *testing.T) {
		for range 100 {
			var target testStruct
			chainer := mvt.Chain(&target).Elem()
			if rand.Intn(2) > 0 {
				chainer = chainer.Field(0)
			} else {
				chainer = chainer.FieldByName("unexportedField")
			}
			value := testImpl(rand.Intn(1000))
			reset := chainer.Set(value)
			if target.unexportedField != value {
				t.Error("target value does not meet expectation", target, value)
			}
			reset.Reset()
			if target.unexportedField != nil {
				t.Error("target value does not meet expectation", target.unexportedField)
			}
		}
	})

	t.Run("测试结构体 2", func(t *testing.T) {
		for range 100 {
			var target testStruct
			impl := testImpl(rand.Intn(1000))
			target.unexportedField = impl
			chainer := mvt.Chain(&target).Elem()
			if rand.Intn(2) > 0 {
				chainer = chainer.Field(0)
			} else {
				chainer = chainer.FieldByName("unexportedField")
			}
			value := testImpl(rand.Intn(1000))
			reset := chainer.Elem().Set(value)
			if target.unexportedField != value {
				t.Error("target value does not meet expectation", target, value)
			}
			reset.Reset()
			if target.unexportedField != impl {
				t.Error("target value does not meet expectation", target.unexportedField)
			}
		}
	})

	t.Run("测试结构体 3", func(t *testing.T) {
		for range 100 {
			var target testStruct
			chainer := mvt.Chain(&target).Elem()
			if rand.Intn(2) > 0 {
				chainer = chainer.Field(2)
			} else {
				chainer = chainer.FieldByName("unexportedField3")
			}
			value := testImpl(rand.Intn(1000))
			reset := chainer.Elem().Set(value)
			if *target.unexportedField3 != value {
				t.Error("target value does not meet expectation", target, value)
			}
			reset.Reset()
			if target.unexportedField3 != nil {
				t.Error("target value does not meet expectation", target.unexportedField)
			}
		}
	})

	t.Run("测试结构体 4", func(t *testing.T) {
		for range 100 {
			var target testStruct
			impl := testImpl(rand.Intn(1000))
			target.unexportedField3 = &impl
			chainer := mvt.Chain(&target).Elem()
			if rand.Intn(2) > 0 {
				chainer = chainer.Field(2)
			} else {
				chainer = chainer.FieldByName("unexportedField3")
			}
			value := testImpl(rand.Intn(1000))
			reset := chainer.Elem().Set(value)
			if *target.unexportedField3 != value {
				t.Error("target value does not meet expectation", target, value)
			}
			reset.Reset()
			if *target.unexportedField3 != impl {
				t.Error("target value does not meet expectation", target.unexportedField, impl)
			}
		}
	})

	t.Run("测试结构体 5", func(t *testing.T) {
		for range 100 {
			var target testStruct
			chainer := mvt.Chain(&target).Elem()
			if rand.Intn(2) > 0 {
				chainer = chainer.Field(3)
			} else {
				chainer = chainer.FieldByName("unexportedField4")
			}
			key := rand.Intn(1000)
			value := map[any]any{key: testImpl(rand.Intn(1000))}
			reset := chainer.Set(value)
			if target.unexportedField4[key] != value[key] {
				t.Error("target value does not meet expectation", target, value)
			}
			reset.Reset()
			if target.unexportedField4 != nil {
				t.Error("target value does not meet expectation", target.unexportedField)
			}
		}
	})

	t.Run("测试结构体 6", func(t *testing.T) {
		for range 100 {
			var target testStruct
			key := rand.Intn(1000)
			impl := testImpl(rand.Intn(1000))
			target.unexportedField4 = map[any]any{key: impl}
			chainer := mvt.Chain(&target).Elem()
			if rand.Intn(2) > 0 {
				chainer = chainer.Field(3)
			} else {
				chainer = chainer.FieldByName("unexportedField4")
			}
			value := map[any]any{key: testImpl(rand.Intn(1000))}
			reset := chainer.Set(value)
			if target.unexportedField4[key] != value[key] {
				t.Error("target value does not meet expectation", target, value)
			}
			reset.Reset()
			if target.unexportedField4[key] != impl {
				t.Error("target value does not meet expectation", target.unexportedField, impl)
			}
		}
	})

	t.Run("测试序列 1", func(t *testing.T) {
		for range 100 {
			var target []any = []any{nil}
			value := rand.Intn(1000)
			reset := mvt.Chain(&target).Elem().Index(0).Set(value)
			if target[0] != value {
				t.Error("target value does not meet expectation", target, value)
			}
			reset.Reset()
			if target[0] != nil {
				t.Error("target value does not meet expectation", target)
			}
		}
	})

	t.Run("测试序列 2", func(t *testing.T) {
		for range 100 {
			var target [1]any = [1]any{nil}
			value := rand.Intn(1000)
			reset := mvt.Chain(&target).Elem().Index(0).Set(value)
			if target[0] != value {
				t.Error("target value does not meet expectation", target, value)
			}
			reset.Reset()
			if target[0] != nil {
				t.Error("target value does not meet expectation", target)
			}
		}
	})

	t.Run("测试序列 3", func(t *testing.T) {
		for range 100 {
			var target []testInterface = []testInterface{testInterface(nil)}
			impl := testImpl(rand.Intn(1000))
			target[0] = impl
			value := testImpl(rand.Intn(1000))
			reset := mvt.Chain(&target).Elem().Index(0).Elem().Set(value)
			if target[0] != value {
				t.Error("target value does not meet expectation", target, value)
			}
			reset.Reset()
			if target[0] != impl {
				t.Error("target value does not meet expectation", target, impl)
			}
		}
	})

	t.Run("测试序列 4", func(t *testing.T) {
		for range 100 {
			var target [1]testInterface = [1]testInterface{testInterface(nil)}
			impl := testImpl(rand.Intn(1000))
			target[0] = impl
			value := testImpl(rand.Intn(1000))
			reset := mvt.Chain(&target).Elem().Index(0).Elem().Set(value)
			if target[0] != value {
				t.Error("target value does not meet expectation", target, value)
			}
			reset.Reset()
			if target[0] != impl {
				t.Error("target value does not meet expectation", target, impl)
			}
		}
	})

	t.Run("测试序列 5", func(t *testing.T) {
		for range 100 {
			var target []map[any]any = []map[any]any{nil}
			impl := testImpl(rand.Intn(1000))
			key := rand.Intn(1000)
			target[0] = map[any]any{key: impl}
			value := map[any]any{key: testImpl(rand.Intn(1000))}
			reset := mvt.Chain(&target).Elem().Index(0).Set(value)
			if target[0][key] != value[key] {
				t.Error("target value does not meet expectation", target, value)
			}
			reset.Reset()
			if target[0][key] != impl {
				t.Error("target value does not meet expectation", target, impl)
			}
		}
	})

	t.Run("测试序列 6", func(t *testing.T) {
		for range 100 {
			var target [1]map[any]any = [1]map[any]any{nil}
			impl := testImpl(rand.Intn(1000))
			key := rand.Intn(1000)
			target[0] = map[any]any{key: impl}
			value := map[any]any{key: testImpl(rand.Intn(1000))}
			reset := mvt.Chain(&target).Elem().Index(0).Set(value)
			if target[0][key] != value[key] {
				t.Error("target value does not meet expectation", target, value)
			}
			reset.Reset()
			if target[0][key] != impl {
				t.Error("target value does not meet expectation", target, impl)
			}
		}
	})

	t.Run("测试序列 7", func(t *testing.T) {
		for range 100 {
			var target []map[any]any = []map[any]any{nil}
			key := rand.Intn(1000)
			value := map[any]any{key: testImpl(rand.Intn(1000))}
			reset := mvt.Chain(&target).Elem().Index(0).Set(value)
			if target[0][key] != value[key] {
				t.Error("target value does not meet expectation", target, value)
			}
			reset.Reset()
			if target[0] != nil {
				t.Error("target value does not meet expectation", target)
			}
		}
	})

	t.Run("测试序列 8", func(t *testing.T) {
		for range 100 {
			var target [1]map[any]any = [1]map[any]any{nil}
			key := rand.Intn(1000)
			value := map[any]any{key: testImpl(rand.Intn(1000))}
			reset := mvt.Chain(&target).Elem().Index(0).Set(value)
			if target[0][key] != value[key] {
				t.Error("target value does not meet expectation", target, value)
			}
			reset.Reset()
			if target[0] != nil {
				t.Error("target value does not meet expectation", target)
			}
		}
	})

	t.Run("测试序列 9", func(t *testing.T) {
		for range 100 {
			var target []*any = []*any{nil}
			value := rand.Intn(1000)
			reset := mvt.Chain(&target).Elem().Index(0).Elem().Set(value)
			if *(target[0]) != value {
				t.Error("target value does not meet expectation", target, value)
			}
			reset.Reset()
			if target[0] != nil {
				t.Error("target value does not meet expectation", target)
			}
		}
	})

	t.Run("测试序列 10", func(t *testing.T) {
		for range 100 {
			var target [1]*any = [1]*any{nil}
			value := rand.Intn(1000)
			reset := mvt.Chain(&target).Elem().Index(0).Elem().Set(value)
			if *(target[0]) != value {
				t.Error("target value does not meet expectation", target, value)
			}
			reset.Reset()
			if target[0] != nil {
				t.Error("target value does not meet expectation", target)
			}
		}
	})

	t.Run("测试序列 11", func(t *testing.T) {
		for range 100 {
			var impl any = testImpl(rand.Intn(1000))
			var target []*any = []*any{&impl}
			value := testImpl(rand.Intn(1000))
			reset := mvt.Chain(&target).Elem().Index(0).Elem().Set(value)
			if *(target[0]) != value {
				t.Error("target value does not meet expectation", target, value)
			}
			reset.Reset()
			if *(target[0]) != impl {
				t.Error("target value does not meet expectation", target)
			}
		}
	})

	t.Run("测试序列 12", func(t *testing.T) {
		for range 100 {
			var impl any = testImpl(rand.Intn(1000))
			var target [1]*any = [1]*any{&impl}
			value := testImpl(rand.Intn(1000))
			reset := mvt.Chain(&target).Elem().Index(0).Elem().Set(value)
			if *(target[0]) != value {
				t.Error("target value does not meet expectation", target, value)
			}
			reset.Reset()
			if *(target[0]) != impl {
				t.Error("target value does not meet expectation", target)
			}
		}
	})

	t.Run("测试 map 1", func(t *testing.T) {
		for range 100 {
			var data map[any]any
			key := rand.Intn(1000)
			value := rand.Intn(1000)
			reset := mvt.Chain(&data).Elem().MapValue(key).Set(value)
			if data[key] != value {
				t.Error("target value does not meet expectation", data, value)
			}
			reset.Reset()
			if data != nil {
				t.Error("target value does not meet expectation", data)
			}
		}
	})

	t.Run("测试 map 2", func(t *testing.T) {
		for range 100 {
			var data = map[any]any{}
			key := rand.Intn(1000)
			value := rand.Intn(1000)
			reset := mvt.Chain(&data).Elem().MapValue(key).Set(value)
			if data[key] != value {
				t.Error("target value does not meet expectation", data, value)
			}
			reset.Reset()
			if data[key] != nil {
				t.Error("target value does not meet expectation", data)
			}
		}
	})

	t.Run("测试 map 3", func(t *testing.T) {
		for range 100 {
			key := rand.Intn(1000)
			impl := testImpl(rand.Intn(1000))
			var data = map[any]any{key: impl}
			value := testImpl(rand.Intn(1000))
			reset := mvt.Chain(&data).Elem().MapValue(key).Elem().Set(value)
			if data[key] != value {
				t.Error("target value does not meet expectation", data, value)
			}
			reset.Reset()
			if data[key] != impl {
				t.Error("target value does not meet expectation", data, impl)
			}
		}
	})

	t.Run("测试 map 4", func(t *testing.T) {
		for range 100 {
			key := rand.Intn(1000)
			var data = map[any]testInterface{key: nil}
			value := testImpl(rand.Intn(1000))
			reset := mvt.Chain(&data).Elem().MapValue(key).Set(value)
			if data[key] != value {
				t.Error("target value does not meet expectation", data, value)
			}
			reset.Reset()
			if data[key] != nil {
				t.Error("target value does not meet expectation", data)
			}
		}
	})

	t.Run("测试 map 5", func(t *testing.T) {
		for range 100 {
			key := rand.Intn(1000)
			var impl testInterface = testImpl(rand.Intn(1000))
			var data = map[any]*testInterface{key: &impl}
			value := testImpl2(rand.Intn(1000))
			reset := mvt.Chain(&data).Elem().MapValue(key).Elem().Set(value)
			if *(data[key]) != value {
				t.Error("target value does not meet expectation", data, value)
			}
			reset.Reset()
			if *(data[key]) != impl {
				t.Error("target value does not meet expectation", data, impl)
			}
		}
	})

	t.Run("测试 map 6", func(t *testing.T) {
		for range 100 {
			var data = map[any]map[any]any{}
			key := rand.Intn(1000)
			value := rand.Intn(1000)
			reset := mvt.Chain(&data).Elem().MapValue(key).MapValue(key).Set(value)
			if data[key][key] != value {
				t.Error("target value does not meet expectation", data, value)
			}
			reset.Reset()
			if data[key] != nil {
				t.Error("target value does not meet expectation", data)
			}
		}
	})

	t.Run("测试函数 1", func(t *testing.T) {
		for range 100 {
			var data = func() any { return nil }
			value := rand.Intn(1000)
			reset := mvt.Chain(&data).Elem().SetFuncOuts([]mvt.OutValue{
				{
					Values: []any{value},
				},
			})
			result := data()
			if result != value {
				t.Error("target value does not meet expectation", result, value)
			}
			reset.Reset()
			result2 := data()
			if result2 != nil {
				t.Error("target value does not meet expectation", result2)
			}
		}
	})

	t.Run("测试函数 2", func(t *testing.T) {
		for range 100 {
			var data = []any{func() any { return nil }}
			value := rand.Intn(1000)
			reset := mvt.Chain(&data).Elem().Index(0).Elem().SetFuncOuts([]mvt.OutValue{
				{
					Values: []any{value},
				},
			})
			result := data[0].(func() any)()
			if result != value {
				t.Error("target value does not meet expectation", result, value)
			}
			reset.Reset()
			result2 := data[0].(func() any)()
			if result2 != nil {
				t.Error("target value does not meet expectation", result2)
			}
		}
	})

	t.Run("测试函数 3", func(t *testing.T) {
		for range 100 {
			var fn any = func() any { return nil }
			var data *any = &fn
			value := rand.Intn(1000)
			reset := mvt.Chain(&data).Elem().Elem().Elem().SetFuncOuts([]mvt.OutValue{
				{
					Values: []any{value},
				},
			})
			result := (*data).(func() any)()
			if result != value {
				t.Error("target value does not meet expectation", result, value)
			}
			reset.Reset()
			result2 := (*data).(func() any)()
			if result2 != nil {
				t.Error("target value does not meet expectation", result2)
			}
		}
	})

	t.Run("复杂链路", func(t *testing.T) {
		type st2 struct {
			field [1]map[any]any
		}
		type st struct {
			field [1]*st2
		}
		var data map[any][]st
		key := rand.Intn(1000)
		defer mvt.Chain(&data).Elem().MapValue(key).Set(make([]st, 1)).Reset()
		value := rand.Intn(1000)
		defer mvt.Chain(&data).Elem().MapValue(key).Index(0).Field(0).Index(0).Elem().Field(0).Index(0).
			MapValue(key).Set(value).Reset()
		if data[key][0].field[0].field[0][key] != value {
			t.Error("target value does not meet expectation", data, value)
		}
	})
}

func (i testImpl) m() int { return int(i) }

func (i testImpl2) m() int { return int(i) }

func (i testStruct) m() int { return 0 }
