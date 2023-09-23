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

// Package matcher 扩展 gomock.Matcher 接口实现。
package matcher

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/golang/mock/gomock"
)

type setMatcher struct {
	x interface{}
}

// Set 设置函数入参值为 v，v 可以为指针值。
// 如果 v 是 nil，则入参值设为零值。
// 如果 v 的类型和要设置的参数类型不一致，方法 Matches 返回 false，其它情形返回 true。
// 如果参数类型不是指针或者可寻址的接口则会在方法 matches 时 panic。
func Set(v interface{}) gomock.Matcher {
	return &setMatcher{v}
}

func (s *setMatcher) Matches(x interface{}) bool {
	val := reflect.ValueOf(x)
	switch val.Type().Kind() {
	case reflect.Ptr, reflect.Interface:
		mval := reflect.Indirect(reflect.ValueOf(s.x))
		if !mval.IsValid() {
			val.Elem().Set(reflect.Zero(val.Type().Elem()))
		} else if mval.Type().AssignableTo(val.Type().Elem()) {
			val.Elem().Set(mval)
		} else {
			return false
		}
		return true
	default:
		panic(fmt.Sprintf("入参数据类型非指针或接口：%T", x))
	}
}

func (s *setMatcher) String() string {
	bytes, _ := json.Marshal(s.x)
	return fmt.Sprintf("\n设置入参值Matcher: %T\n%s\n", s.x, bytes)
}
