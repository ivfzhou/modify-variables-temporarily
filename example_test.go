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

	mvt "gitee.com/ivfzhou/modify-variables-temporarily/v3"
)

func ExampleVar() {
	var target any
	func() {
		defer mvt.Var(&target, "any your want").Reset()
		fmt.Println(target.(string))
	}()
	fmt.Println(target)
	// Output:
	// any your want
	// <nil>
}

func ExampleFieldByName() {
	var target struct {
		name string
	}
	func() {
		defer mvt.FieldByName(&target, "name", "any your want").Reset()
		fmt.Println(target.name)
	}()
	fmt.Println(target.name)
	// Output:
	// any your want
	//
}

func ExampleField() {
	var target struct {
		name string
	}
	func() {
		defer mvt.Field(&target, 0, "any your want").Reset()
		fmt.Println(target.name)
	}()
	fmt.Println(target.name)
	// Output:
	// any your want
	//
}

func ExampleElem() {
	var target = []any{1, "ok"}
	func() {
		defer mvt.Elem(target, 1, "any your want").Reset()
		fmt.Println(target[1])
	}()
	fmt.Println(target[1])
	// Output:
	// any your want
	// ok
}

func ExampleMap() {
	m := map[string]any{}
	func() {
		defer mvt.Map(m, "key", "any your want").Reset()
		fmt.Println(m["key"])
	}()
	fmt.Println(m["key"])
	// Output:
	// any your want
	// <nil>
}

func ExampleFuncOuts() {
	fn := func() (any, error) { return nil, nil }
	func() {
		defer mvt.FuncOuts(&fn, []mvt.OutValue{
			{Values: []any{"ok", errors.New("error occurred")}},
		})
		a, err := fn()
		fmt.Println(a, err)
		a, err = fn()
		fmt.Println(a, err)
	}()
	a, err := fn()
	fmt.Println(a, err)
	// Output:
	// <nil> <nil>
	// <nil> <nil>
	// ok error occurred
}

func ExampleChain_run1() {
	type aUnexportedStruct struct {
		m map[any][1]struct {
			field string
		}
	}
	var Data *aUnexportedStruct

	key := 1
	reset := mvt.Chain(&Data).Elem().Elem().FieldByName("m").MapValue(key).Index(0).Field(0).Set("any your want")
	defer func() {
		reset.Reset()
		fmt.Println(Data)
	}()

	fmt.Println(Data.m[key][0].field)
	// Output:
	// any your want
	// <nil>
}

func ExampleChain_run2() {
	type aUnexportedStruct struct {
		name string
	}
	var Data *aUnexportedStruct

	reset := mvt.Chain(&Data).Elem().Elem().FieldByName("name").Set("any your want")
	defer func() {
		reset.Reset()
		fmt.Println(Data)
	}()

	fmt.Println(Data.name)
	// Output:
	// any your want
	// <nil>
}

func ExampleChain_run3() {
	type aInterface interface{}

	type aImpl struct {
		name any
	}

	type aImpl2 struct {
		other any
	}

	var Data aInterface = aImpl{name: "impl1"}

	impl2 := aImpl2{other: "impl2"}
	reset := mvt.Chain(&Data).Elem().Set(impl2)
	fmt.Println(Data.(aImpl2).other)
	reset.Reset()
	fmt.Println(Data.(aImpl).name)

	reset2 := mvt.Chain(&Data).Elem().Elem().Field(0).Set("changed")
	fmt.Println(Data.(aImpl).name)
	reset2.Reset()
	fmt.Println(Data.(aImpl).name)
	// Output:
	// impl2
	// impl1
	// changed
	// impl1
}

func ExampleChain_run4() {
	type aUnexportedStruct struct {
		name string
		m    map[any]any
		arr  [1]any
		s    []any
		fn   func()
		i    int
		str  string
	}
	var Data *aUnexportedStruct

	key := 1
	reset := mvt.Chain(&Data).Elem().Elem().FieldByName("m").MapValue(key).Set("any your want") // 尽管 map 是 nil，仍然可以设置。
	defer func() {
		reset.Reset()
		fmt.Println(Data)
	}()

	fmt.Println(Data.m[key])
	// Output:
	// any your want
	// <nil>
}

func ExampleChain_run5() {
	type aUnexportedStruct struct {
		name string
	}
	var Data *aUnexportedStruct

	reset := mvt.Chain(&Data).Elem().Elem().Set(struct{ name string }{"any your want"})
	defer func() {
		reset.Reset()
		fmt.Println(Data)
	}()

	fmt.Println(Data.name)
	// Output:
	// any your want
	// <nil>
}

func ExampleChain_run6() {
	type aUnexportedStruct struct {
		name string
	}
	var Data = &aUnexportedStruct{name: "something"}

	reset := mvt.Chain(&Data).Elem().Elem().Field(0).Set("any your want")
	defer func() {
		reset.Reset()
		fmt.Println(Data.name)
	}()

	fmt.Println(Data.name)
	// Output:
	// any your want
	// something
}

func ExampleChain_run7() {
	type aUnexportedStruct struct {
		m map[any][1]struct {
			field  []string
			field2 struct {
				s []string
			}
		}
		s []string
	}
	var Data *aUnexportedStruct

	key := 1
	s := []string{"any your want"}
	reset := mvt.Chain(&Data).Elem().Elem().FieldByName("m").MapValue(key).Index(0).Field(0).Set(s)
	defer func() {
		reset.Reset()
		fmt.Println(Data)
	}()
	fmt.Println(Data.m[key][0].field[0])

	reset2 := mvt.Chain(&Data).Elem().Elem().FieldByName("s").Set(s)
	defer func() {
		reset2.Reset()
		fmt.Println(Data.s == nil)
	}()
	fmt.Println(Data.s[0])

	reset3 := mvt.Chain(&Data).Elem().Elem().FieldByName("m").MapValue(key).Index(0).Field(1).Field(0).Set(s)
	defer func() {
		reset3.Reset()
		fmt.Println(Data.m[key][0].field2.s == nil)
	}()
	fmt.Println(Data.m[key][0].field2.s[0])

	s[0] = "I changed"
	fmt.Println(Data.m[key][0].field[0])
	fmt.Println(Data.s[0])
	fmt.Println(Data.m[key][0].field2.s[0])
	// Output:
	// any your want
	// any your want
	// any your want
	// I changed
	// I changed
	// I changed
	// true
	// true
	// <nil>
}

func ExampleChain_run8() {
	type aUnexportedStruct struct {
		m map[any][1]struct {
			field struct {
				fn func(any) any
			}
		}
	}
	var Data *aUnexportedStruct

	key := 1
	outs := []mvt.OutValue{
		{
			Values: []any{1},
		},
		{
			Values: []any{"abc"},
			Times:  2,
		},
		{
			Values: []any{nil},
		},
	}
	reset := mvt.Chain(&Data).Elem().Elem().FieldByName("m").MapValue(key).Index(0).Field(0).FieldByName("fn").SetFuncOuts(outs)
	defer func() {
		reset.Reset()
		fmt.Println(Data)
	}()

	for range 4 {
		fmt.Println(Data.m[key][0].field.fn(nil))
	}
	// Output:
	// 1
	// abc
	// abc
	// <nil>
	// <nil>
}

func ExampleChain_run9() {
	type aUnexportedStruct struct {
		fn func(any) any
	}
	var Data = &aUnexportedStruct{func(any) any {
		return "nothing"
	}}

	outs := []mvt.OutValue{
		{
			Values: []any{1},
		},
		{
			Values: []any{"abc"},
			Times:  2,
		},
		{
			Values: []any{nil},
		},
	}
	defer mvt.Chain(&Data).Elem().Elem().FieldByName("fn").SetFuncOuts(outs).Reset()

	for range 4 {
		fmt.Println(Data.fn(nil))
	}

	fmt.Println(Data.fn(nil))
	// Output:
	// 1
	// abc
	// abc
	// <nil>
	// nothing
}
