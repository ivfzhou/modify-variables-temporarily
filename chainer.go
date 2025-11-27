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

// Chainer 路由到要修改的变量。
type Chainer interface {
	// Elem 当前变量是指针或接口类型，获取它的内部类型变量。
	Elem() ChainSetter

	// FieldByName 当前变量是结构体类型，获取它的一个字段变量。
	FieldByName(name string) ChainSetter

	// Field 当前变量是结构体类型，获取它的一个字段变量。
	Field(index int) ChainSetter

	// MapValue 当前变量是映射类型，获取它的一个键的值变量。
	MapValue(key any) ChainSetter

	// Index 当前变量是数组或切片类型，获取它的一个元素的变量。
	Index(index int) ChainSetter
}
