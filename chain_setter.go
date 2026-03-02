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
	"strings"
)

const (
	toElem actionType = iota + 1
	toStructField
	toStructFieldByName
	toMapValue
	toSeqElem
)

type ChainSetter interface {
	Chainer
	Setter
}

type actionType int

type action struct {
	typ  actionType
	args []any
}

type chainSetter struct {
	value   *reflect.Value
	actions []*action
}

func (c *chainSetter) Elem() ChainSetter {
	act := &action{typ: toElem}
	actions := make([]*action, len(c.actions)+1)
	copy(actions, c.actions)
	actions[len(c.actions)] = act
	c.actions = append(c.actions, act)
	return &chainSetter{c.value, actions}
}

func (c *chainSetter) FieldByName(name string) ChainSetter {
	act := &action{toStructFieldByName, []any{name}}
	actions := make([]*action, len(c.actions)+1)
	copy(actions, c.actions)
	actions[len(c.actions)] = act
	c.actions = append(c.actions, act)
	return &chainSetter{c.value, actions}
}

func (c *chainSetter) Field(index int) ChainSetter {
	act := &action{toStructField, []any{index}}
	actions := make([]*action, len(c.actions)+1)
	copy(actions, c.actions)
	actions[len(c.actions)] = act
	c.actions = append(c.actions, act)
	return &chainSetter{c.value, actions}
}

func (c *chainSetter) MapValue(key any) ChainSetter {
	act := &action{toMapValue, []any{key}}
	actions := make([]*action, len(c.actions)+1)
	copy(actions, c.actions)
	actions[len(c.actions)] = act
	c.actions = append(c.actions, act)
	return &chainSetter{c.value, actions}
}

func (c *chainSetter) Index(index int) ChainSetter {
	act := &action{toSeqElem, []any{index}}
	actions := make([]*action, len(c.actions)+1)
	copy(actions, c.actions)
	actions[len(c.actions)] = act
	c.actions = append(c.actions, act)
	return &chainSetter{c.value, actions}
}

func (c *chainSetter) Set(substitute any) Resetter {
	if len(c.actions) <= 0 {
		panic(ErrNoActions)
	}

	callbackFuncs, restoreFuncs, value, _ := c.seekValue(*c.value)

	old := value.Interface()
	substituteValue := convertSubstituteToTypeValue(substitute, value.Type())
	value.Set(substituteValue)
	reset := generateSetOldFunc(value, old)
	restoreFuncs = append(restoreFuncs, reset)

	for i := len(callbackFuncs) - 1; i >= 0; i-- {
		callbackFuncs[i]()
	}

	return newResetter(func() {
		for i := len(restoreFuncs) - 1; i >= 0; i-- {
			restoreFuncs[i]()
		}
	})
}

func (c *chainSetter) SetFuncOuts(outs []OutValue) Resetter {
	if len(c.actions) <= 0 {
		panic(ErrNoActions)
	}

	callbackFuncs, restoreFuncs, value, typeChain := c.seekValue(*c.value)
	if value.Kind() != reflect.Func {
		panic(newTypeInvalid(ErrTargetIsNotFunc, typeChain))
	}
	old := value.Interface()
	newFuncValue := makeFunc(value, outs)
	value.Set(newFuncValue)
	reset := generateSetOldFunc(value, old)
	restoreFuncs = append(restoreFuncs, reset)

	for i := len(callbackFuncs) - 1; i >= 0; i-- {
		callbackFuncs[i]()
	}

	return newResetter(func() {
		for i := len(restoreFuncs) - 1; i >= 0; i-- {
			restoreFuncs[i]()
		}
	})
}

func (c *chainSetter) seekValue(value reflect.Value) (
	callbackFuncs, restoreFuncs []func(), lastValue reflect.Value, typeChain string) {

	types := make([]string, 0, len(c.actions))
	restoreFuncs = make([]func(), 0, len(c.actions)+1)
	callbackFuncs = make([]func(), 0, len(c.actions)+1)
	for _, v := range c.actions {
		if !value.IsValid() {
			panic(newTypeInvalid(ErrCannotToNext, strings.Join(types, " -> ")))
		}
		types = append(types, value.Type().String())

		switch v.typ {
		case toElem:
			switch value.Kind() {
			case reflect.Pointer:
				elemValue := value.Elem()
				if !elemValue.IsValid() {
					panic(newTypeInvalid(ErrCannotToNext, strings.Join(types, " -> ")))
				}

				// 指向的类型是零值变量。
				if elemValue.IsZero() {
					switch elemValue.Kind() {
					case reflect.Pointer: // 指针指向一个指针，指向的指针是 nil 值，那么初始化下它。
						elemValue.Set(reflect.New(elemValue.Type().Elem())) // 给指向的指针变量分配空间。

						// 回退修改时，把指向的指针变量重新设置成 nil。
						tmpElemValue := elemValue
						restoreFuncs = append(restoreFuncs, func() { tmpElemValue.SetZero() })
					case reflect.Map: // 指针指向一个映射，该映射是 nil 值，那么初始化下它。
						elemValue.Set(reflect.MakeMap(elemValue.Type())) // 给该映射变量分配空间。

						// 回退修改时，把指向的指针变量重新设置成 nil。
						tmpMapValue := elemValue
						restoreFuncs = append(restoreFuncs, func() { tmpMapValue.SetZero() })
					case reflect.Interface: // 初始化一个，避免获取的 Elem() 是 Invalid。
						elemValue.Set(reflect.New(elemValue.Type()).Elem())

						// 回退修改时，把它设置成 nil。
						tmpInterfaceValue := elemValue
						restoreFuncs = append(restoreFuncs, func() { tmpInterfaceValue.SetZero() })
					}
				}

				value = elemValue
			case reflect.Interface:
				implValue := value.Elem()
				if !implValue.IsValid() {
					panic(newTypeInvalid(ErrCannotToNext, strings.Join(types, " -> ")))
				}

				// 接口内部类型不可寻址，所以我们分配一个新的变量使用。同时复制下值。
				newImplValue := reflect.New(implValue.Type()).Elem()
				newImplValue.Set(implValue)

				// 如果是 nil 指针或映射，需要分配下内存。
				if newImplValue.IsZero() {
					switch newImplValue.Kind() {
					case reflect.Pointer:
						newImplValue.Set(reflect.New(newImplValue.Type().Elem()))
					case reflect.Map:
						newImplValue.Set(reflect.MakeMap(newImplValue.Type()))
					case reflect.Interface:
						newImplValue.Set(reflect.New(newImplValue.Type()).Elem())
					}
				}

				// 在回退修改时，把接口设置成原来的值。
				tmpInterfaceValue := value
				restoreFuncs = append(restoreFuncs, func() { tmpInterfaceValue.Set(implValue) })

				// 其它变量修改完时，把新建变量赋值给接口。
				tmpInterfaceValue2 := value
				callbackFuncs = append(callbackFuncs, func() { tmpInterfaceValue2.Set(newImplValue) })

				value = newImplValue
			default:
				panic(newTypeInvalid(ErrTargetIsNotPointerOrInterface, strings.Join(types, " -> ")))
			}
		case toStructField:
			switch value.Kind() {
			case reflect.Struct:
				index := v.args[0].(int)
				fieldValue := getStructField(value, index)
				if !fieldValue.IsValid() {
					panic(newTypeInvalid(ErrCannotToNext, strings.Join(types, " -> ")))
				}

				if fieldValue.IsZero() {
					switch fieldValue.Kind() {
					case reflect.Map: // 分配一个新映射使用。
						fieldValue.Set(reflect.MakeMap(fieldValue.Type()))

						// 回退修改时，重置成 nil 映射。
						tmpMapValue := fieldValue
						restoreFuncs = append(restoreFuncs, func() { tmpMapValue.SetZero() })
					case reflect.Pointer: // 也需要分配一个新变量使用。
						fieldValue.Set(reflect.New(fieldValue.Type().Elem()))

						// 回退修改时，重置成 nil 指针。
						tmpPtrValue := fieldValue
						restoreFuncs = append(restoreFuncs, func() { tmpPtrValue.SetZero() })
					case reflect.Interface:
						fieldValue.Set(reflect.New(fieldValue.Type()).Elem())

						tmpInterfaceValue := fieldValue
						restoreFuncs = append(restoreFuncs, func() { tmpInterfaceValue.SetZero() })
					}
				}

				value = fieldValue
			default:
				panic(newTypeInvalid(ErrTargetIsNotStruct, strings.Join(types, " -> ")))
			}
		case toStructFieldByName:
			switch value.Kind() {
			case reflect.Struct:
				name := v.args[0].(string)
				if len(name) <= 0 {
					panic(ErrStructFieldNameCannotBeEmpty)
				}
				fieldValue := getStructFieldByName(value, name)
				if !fieldValue.IsValid() {
					panic(newTypeInvalid(ErrCannotToNext, strings.Join(types, " -> ")))
				}

				if fieldValue.IsZero() {
					switch fieldValue.Kind() {
					case reflect.Pointer:
						fieldValue.Set(reflect.New(fieldValue.Type().Elem()))

						// 回退修改时，重置成 nil 指针。
						tmpPtrValue := fieldValue
						restoreFuncs = append(restoreFuncs, func() { tmpPtrValue.SetZero() })
					case reflect.Map:
						fieldValue.Set(reflect.MakeMap(fieldValue.Type()))

						// 回退修改时，重置成 nil 映射。
						tmpMapValue := fieldValue
						restoreFuncs = append(restoreFuncs, func() { tmpMapValue.SetZero() })
					case reflect.Interface:
						fieldValue.Set(reflect.New(fieldValue.Type()).Elem())

						tmpInterfaceValue := fieldValue
						restoreFuncs = append(restoreFuncs, func() { tmpInterfaceValue.SetZero() })
					}
				}

				value = fieldValue
			default:
				panic(newTypeInvalid(ErrTargetIsNotStruct, strings.Join(types, " -> ")))
			}
		case toMapValue:
			switch value.Kind() {
			case reflect.Map:
				key := v.args[0]
				keyValue, mapValValue := getMapValueByKey(value, key)
				mapOldValValue := reflect.Zero(value.Type().Elem())
				if mapValValue.IsValid() {
					mapOldValValue = reflect.ValueOf(mapValValue.Interface())
				}

				// 映射键值存在，新建一个键值变量，然后复制下原键值。
				newMapValValue := reflect.New(value.Type().Elem()).Elem()
				// 键值存在时复制。
				if mapValValue.IsValid() {
					newMapValValue.Set(mapValValue)
				}

				// nil 指针和映射需要分配下内存。
				if newMapValValue.IsZero() {
					switch newMapValValue.Kind() {
					case reflect.Map:
						newMapValValue.Set(reflect.MakeMap(newMapValValue.Type()))
					case reflect.Pointer:
						newMapValValue.Set(reflect.New(newMapValValue.Type().Elem()))
					case reflect.Interface:
						newMapValValue.Set(reflect.New(newMapValValue.Type()).Elem())
					}
				}

				// 映射的键值是不可寻址的，所以修改完其它变量时，回来设置下映射键值。
				tmpMapValue := value
				callbackFuncs = append(callbackFuncs, func() { tmpMapValue.SetMapIndex(keyValue, newMapValValue) })

				// 回退修改时，把映射的键修改回来。
				restoreFuncs = append(restoreFuncs, func() { tmpMapValue.SetMapIndex(keyValue, mapOldValValue) })

				value = newMapValValue
			default:
				panic(newTypeInvalid(ErrTargetIsNotMap, strings.Join(types, " -> ")))
			}
		case toSeqElem:
			switch value.Kind() {
			case reflect.Slice, reflect.Array:
				index := v.args[0].(int)
				elemValue := getSequenceIndex(value, index)
				if !elemValue.IsValid() {
					panic(newTypeInvalid(ErrCannotToNext, strings.Join(types, " -> ")))
				}

				// nil 指针和映射需要初始化下。
				if elemValue.IsZero() {
					switch elemValue.Kind() {
					case reflect.Pointer:
						elemValue.Set(reflect.New(elemValue.Type().Elem()))

						tmpPtrValue := elemValue
						restoreFuncs = append(restoreFuncs, func() { tmpPtrValue.SetZero() })
					case reflect.Map:
						elemValue.Set(reflect.MakeMap(elemValue.Type()))

						tmpMapValue := elemValue
						restoreFuncs = append(restoreFuncs, func() { tmpMapValue.SetZero() })
					case reflect.Interface:
						elemValue.Set(reflect.New(elemValue.Type()).Elem())

						tmpInterfaceValue := elemValue
						restoreFuncs = append(restoreFuncs, func() { tmpInterfaceValue.SetZero() })
					}
				}

				value = elemValue
			default:
				panic(newTypeInvalid(ErrTargetIsNotSliceOrArray, strings.Join(types, " -> ")))
			}
		}
	}
	lastValue = value
	typeChain = strings.Join(types, " -> ")

	return
}
