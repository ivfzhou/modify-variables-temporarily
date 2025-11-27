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
	"sync/atomic"
)

// OutValue 函数返回值。
type OutValue struct {
	// Values 函数调用的返回值
	Values []any
	// Times 每组返回值返回次数。0 和 1 表示仅返回一次。
	Times int
}

func makeFunc(funcValue reflect.Value, outs []OutValue) reflect.Value {
	funcType := funcValue.Type()
	outValues := generateFuncOutValues(funcType, outs)
	length := int64(len(outValues))
	count := int64(-1)
	keptFuncValue := reflect.ValueOf(funcValue.Interface())
	return reflect.MakeFunc(funcType, func(ins []reflect.Value) []reflect.Value {
		index := atomic.AddInt64(&count, 1)
		if index < length {
			return outValues[index]
		} else {
			return keptFuncValue.Call(ins)
		}
	})
}

func generateFuncOutValues(funcType reflect.Type, outs []OutValue) [][]reflect.Value {
	result := make([][]reflect.Value, 0, len(outs))
	numOut := funcType.NumOut()
	for _, v := range outs {
		out := make([]reflect.Value, 0, numOut)
		for index, value := range v.Values {
			if index >= numOut {
				break
			}
			outValueType := funcType.Out(index)
			newOutValue := reflect.New(outValueType).Elem()
			newOutValue.Set(convertSubstituteToTypeValue(value, outValueType))
			out = append(out, newOutValue)
		}
		for len(out) < numOut {
			outValueType := funcType.Out(len(out) - 1)
			out = append(out, reflect.Zero(outValueType))
		}
		if v.Times <= 0 {
			v.Times = 1
		}
		for i := 0; i < v.Times; i++ {
			result = append(result, out)
		}
	}
	return result
}
