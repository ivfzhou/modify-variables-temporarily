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
	"reflect"
	"testing"

	modify_variables_temporarily "gitee.com/ivfzhou/modify-variables-temporarily"
)

func TestMvt_Path(t *testing.T) {
	type Somebody struct {
		Name string
	}
	target := &Somebody{}
	m := modify_variables_temporarily.NewWithTarget(target).PathNext("Name", "zhangsan")
	assert(t, target.Name, "zhangsan")
	m.Reset()
	assert(t, target.Name, "")

	arr := [1]*Somebody{
		{"lisi"},
	}
	m.Path(arr, "0.Name", "zhangsan")
	assert(t, arr[0].Name, "zhangsan")
	m.Reset()
	assert(t, arr[0].Name, "lisi")
}

func assert(t *testing.T, a interface{}, b interface{}) {
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("%T dose not equal %T", a, b)
	}
}
