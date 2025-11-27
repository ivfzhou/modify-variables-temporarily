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
	"unicode"
)

// Var 替换变量的值。
// target 被替换的变量，须是指针类型，不能是 nil。
// substitute 替换成的变量。
func Var(target, substitute any) Resetter {
	if target == nil {
		panic(ErrTargetCannotBeNil)
	}
	ptrValue := reflect.ValueOf(target)
	if ptrValue.Kind() != reflect.Ptr {
		panic(ErrTargetIsNotPointer)
	}
	elemValue := ptrValue.Elem()
	if elemValue.Kind() == reflect.Invalid {
		panic(ErrTargetCannotBeNilType)
	}
	oldElem := elemValue.Interface()
	newElemValue := convertSubstituteToTypeValue(substitute, elemValue.Type())
	elemValue.Set(newElemValue)
	return newResetter(generateSetOldFunc(elemValue, oldElem))
}

// FieldByName 替换结构体字段的值。
// target 被替换字段值的结构体变量，须是结构体指针类型，不能是 nil。
// name 结构体字段的名称，可以是不导出的字段，但不能是空串。
// substitute 替换成的变量。
func FieldByName(target any, name string, substitute any) Resetter {
	if target == nil {
		panic(ErrTargetCannotBeNil)
	}
	if len(name) <= 0 {
		panic(ErrStructFieldNameCannotBeEmpty)
	}
	ptrValue := reflect.ValueOf(target)
	if ptrValue.Kind() != reflect.Ptr {
		panic(ErrTargetIsNotPointer)
	}
	structValue := ptrValue.Elem()
	if structValue.Kind() == reflect.Invalid {
		panic(ErrTargetCannotBeNilType)
	}
	if structValue.Kind() != reflect.Struct {
		panic(ErrTargetIsNotStruct)
	}
	fieldValue := getStructFieldByName(structValue, name)
	oldField := fieldValue.Interface()
	newFieldValue := convertSubstituteToTypeValue(substitute, fieldValue.Type())
	fieldValue.Set(newFieldValue)
	return newResetter(generateSetOldFunc(fieldValue, oldField))
}

// Field 替换结构体字段的值。
// target 被修改字段值的结构体变量，须是结构体指针类型，不能是 nil。
// index 字段序号。
// substitute 替换成的变量。
func Field(target any, index int, substitute any) Resetter {
	if target == nil {
		panic(ErrTargetCannotBeNil)
	}
	ptrValue := reflect.ValueOf(target)
	if ptrValue.Kind() != reflect.Ptr {
		panic(ErrTargetIsNotPointer)
	}
	structValue := ptrValue.Elem()
	if structValue.Kind() == reflect.Invalid {
		panic(ErrTargetCannotBeNilType)
	}
	if structValue.Kind() != reflect.Struct {
		panic(ErrTargetIsNotStruct)
	}
	fieldValue := getStructField(structValue, index)
	oldField := fieldValue.Interface()
	newFieldValue := convertSubstituteToTypeValue(substitute, fieldValue.Type())
	fieldValue.Set(newFieldValue)
	return newResetter(generateSetOldFunc(fieldValue, oldField))
}

// Elem 替换切片的元素值。
// target 被替换元素的切片变量，不能是 nil。
// index 被替换元素的下标。
// substitute 替换成的变量。
func Elem(target any, index int, substitute any) Resetter {
	if target == nil {
		panic(ErrTargetCannotBeNil)
	}
	sliceValue := reflect.ValueOf(target)
	if sliceValue.Kind() != reflect.Slice {
		panic(ErrTargetIsNotSlice)
	}
	if sliceValue.IsZero() {
		panic(ErrTargetCannotBeNilType)
	}
	elemValue := getSequenceIndex(sliceValue, index)
	oldElem := elemValue.Interface()
	newElemValue := convertSubstituteToTypeValue(substitute, elemValue.Type())
	elemValue.Set(newElemValue)
	return newResetter(generateSetOldFunc(elemValue, oldElem))
}

// Map 替换映射中的某个键的值。
// target 被替换元素的映射变量，不能是 nil。
// key 将被修改映射值的键。
// substitute 要替换成的变量。
func Map(target, key any, substitute any) Resetter {
	if target == nil {
		panic(ErrTargetCannotBeNil)
	}
	mapValue := reflect.ValueOf(target)
	if mapValue.Kind() != reflect.Map {
		panic(ErrTargetIsNotMap)
	}
	if mapValue.IsZero() {
		panic(ErrTargetCannotBeNilType)
	}
	keyValue, valValue := getMapValueByKey(mapValue, key)
	newValValue := convertSubstituteToTypeValue(substitute, mapValue.Type().Elem())
	mapValue.SetMapIndex(keyValue, newValValue)
	return newResetter(func() { mapValue.SetMapIndex(keyValue, valValue) })
}

// FuncOuts 替换函数变量以固定次数返回值代替。
// target 要被替换返回值的函数指针变量，不能是 nil。
// outs 替换成的输出值，函数返回值将会复制 outs 中的值返回，
func FuncOuts(target any, outs []OutValue) Resetter {
	if target == nil {
		panic(ErrTargetCannotBeNil)
	}
	ptrValue := reflect.ValueOf(target)
	if ptrValue.Kind() != reflect.Ptr {
		panic(ErrTargetIsNotPointer)
	}
	funcValue := ptrValue.Elem()
	if funcValue.Kind() != reflect.Func {
		panic(ErrTargetIsNotFunc)
	}
	fn := funcValue.Interface()
	funcValue.Set(makeFunc(funcValue, outs))
	return newResetter(generateSetOldFunc(funcValue, fn))
}

// Chain 根据索引替换深层值。
// target 必须是可寻址的变量，不能是 nil。
func Chain(target any) Chainer {
	if target == nil {
		panic(ErrTargetCannotBeNil)
	}
	value := reflect.ValueOf(target)
	if value.IsZero() {
		panic(ErrTargetCannotBeNilType)
	}
	switch value.Kind() {
	case reflect.Slice, reflect.Map, reflect.Ptr:
	default:
		if !value.CanSet() {
			panic(ErrTargetCannotBeSet)
		}
	}
	return &chainSetter{value: &value}
}

func convertSubstituteToTypeValue(substitute any, typ reflect.Type) reflect.Value {
	if substitute == nil {
		return reflect.Zero(typ)
	}
	substituteValue := reflect.ValueOf(substitute)
	substituteType := substituteValue.Type()
	if !substituteType.AssignableTo(typ) {
		if substituteType.ConvertibleTo(typ) {
			substituteValue = substituteValue.Convert(typ)
		} else {
			panic(newIncompatibleTypeAssignmentError(substituteType.String(), typ.String()))
		}
	}
	return substituteValue
}

func generateSetOldFunc(value reflect.Value, old any) func() {
	oldValue := reflect.ValueOf(old)
	valType := value.Type()
	if old == nil {
		oldValue = reflect.Zero(valType)
	}
	oldValueType := oldValue.Type()
	if !oldValueType.AssignableTo(valType) {
		if oldValueType.ConvertibleTo(valType) {
			oldValue = oldValue.Convert(valType)
		} else {
			panic(newIncompatibleTypeAssignmentError(oldValueType.String(), valType.String()))
		}
	}
	return func() { value.Set(oldValue) }
}

func getStructFieldByName(structValue reflect.Value, name string) reflect.Value {
	fieldValue := structValue.FieldByName(name)
	if !fieldValue.IsValid() {
		panic(newStructFieldNotFoundByNameError(structValue.Type().String(), name))
	}
	fieldType := fieldValue.Type()
	if !unicode.IsUpper(rune(name[0])) {
		fieldValue = reflect.NewAt(fieldType, fieldValue.Addr().UnsafePointer()).Elem()
	}
	return fieldValue
}

func getStructField(structValue reflect.Value, index int) reflect.Value {
	numField := structValue.NumField()
	if index >= numField || index < -numField {
		panic(newStructFieldNotFoundError(structValue.Type().String(), index))
	}
	revisedIndex := index
	if index < 0 {
		revisedIndex = -index
		revisedIndex = numField - revisedIndex
	}
	fieldValue := structValue.Field(revisedIndex)
	fieldType := fieldValue.Type()
	if !unicode.IsUpper(rune(structValue.Type().Field(revisedIndex).Name[0])) {
		fieldValue = reflect.NewAt(fieldType, fieldValue.Addr().UnsafePointer()).Elem()
	}
	return fieldValue
}

func getSequenceIndex(seqValue reflect.Value, index int) reflect.Value {
	length := seqValue.Len()
	if index >= length || index < -length {
		panic(newIndexOutOfBoundError(index, seqValue.Type().String(), length))
	}
	revisedIndex := index
	if index < 0 {
		revisedIndex = -revisedIndex
		revisedIndex = length - revisedIndex
	}
	return seqValue.Index(revisedIndex)
}

func getMapValueByKey(mapValue reflect.Value, key any) (keyValue reflect.Value, valValue reflect.Value) {
	if key == nil {
		keyValue = reflect.Zero(mapValue.Type().Key())
		valValue = mapValue.MapIndex(keyValue)
	} else {
		keyValue = reflect.ValueOf(key)
		keyType := reflect.ValueOf(key).Type()
		mapKeyType := mapValue.Type().Key()
		if !keyType.AssignableTo(mapKeyType) {
			if keyType.ConvertibleTo(mapKeyType) {
				keyValue = keyValue.Convert(mapKeyType)
			} else {
				panic(newInvalidMapKeyError(keyType.String(), mapValue.Type().String()))
			}
		}
		valValue = mapValue.MapIndex(keyValue)
	}
	return
}
