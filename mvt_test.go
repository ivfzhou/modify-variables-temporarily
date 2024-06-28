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
	"fmt"
	"testing"

	mvt "gitee.com/ivfzhou/modify-variables-temporarily/v2"
)

func TestMvt_Var(t *testing.T) {
	x := 1
	target := &x
	m := mvt.New()
	m.Var(target, 2)
	if *target != 2 {
		t.Errorf("Mvt.Var test failure, *p is %v", *target)
	}
	if x != 2 {
		t.Errorf("Mvt.Var test failure, x is %v", x)
	}
	m.Reset()
	if *target != 1 {
		t.Errorf("Mvt.Var test failure, *p is %v", *target)
	}
	if x != 1 {
		t.Errorf("Mvt.Var test failure, x is %v", x)
	}
}

func ExampleMvt_Var() {
	target := 1
	m := mvt.New()
	m.Var(&target, 2)

	// now x is 2

	m.Reset()

	// now x is 1
}

func TestMvt_FieldByName(t *testing.T) {
	type S struct {
		name    string
		Petname string
	}
	target := &S{
		name:    "zs",
		Petname: "ls",
	}
	m := mvt.New()
	m.FieldByName(target, "Petname", "ww")
	if target.Petname != "ww" {
		t.Errorf("Mvt.FieldByName test failure, Petname is %v", target.Petname)
	}
	m.FieldByName(target, "name", "zl")
	if target.name != "zl" {
		t.Errorf("Mvt.FieldByName test failure, name is %v", target.name)
	}
	m.Reset()
	if target.Petname != "ls" {
		t.Errorf("Mvt.FieldByName test failure, Petname is %v", target.Petname)
	}
	if target.name != "zs" {
		t.Errorf("Mvt.FieldByName test failure, name is %v", target.name)
	}
}

func ExampleMvt_FieldByName() {
	type S struct {
		name    string
		Petname string
	}
	target := &S{
		name:    "zs",
		Petname: "ls",
	}
	m := mvt.New()
	m.FieldByName(target, "Petname", "ww").FieldByName(target, "name", "zl")

	// now target.Petname is ww, and target.name is zl

	m.Reset()

	// now target.Petname is ls, and target.name is zs
}

func TestMvt_Field(t *testing.T) {
	type S struct {
		name    string
		Petname string
	}
	target := &S{
		name:    "zs",
		Petname: "ls",
	}
	m := mvt.New()
	m.Field(target, 1, "ww")
	if target.Petname != "ww" {
		t.Errorf("Mvt.FieldByName test failure, Petname is %v", target.Petname)
	}
	m.Field(target, 0, "zl")
	if target.name != "zl" {
		t.Errorf("Mvt.FieldByName test failure, name is %v", target.name)
	}
	m.Reset()
	if target.Petname != "ls" {
		t.Errorf("Mvt.FieldByName test failure, Petname is %v", target.Petname)
	}
	if target.name != "zs" {
		t.Errorf("Mvt.FieldByName test failure, name is %v", target.name)
	}
}

func ExampleMvt_Field() {
	type S struct {
		name    string
		Petname string
	}
	target := &S{
		name:    "zs",
		Petname: "ls",
	}
	m := mvt.New()
	m.Field(target, 1, "ww").Field(target, 0, "zl")

	// now target.Petname is ww, and target.name is zl

	m.Reset()

	// now target.Petname is ls, and target.name is zs
}

func TestMvt_Elem(t *testing.T) {
	target := []int{1, 2, 3}
	m := mvt.New()
	m.Elem(target, 0, -1).Elem(target, 1, -2).Elem(target, 2, -3)
	if target[0] != -1 {
		t.Errorf("Mvt.Elem test failure, %v", target[0])
	}
	if target[1] != -2 {
		t.Errorf("Mvt.Elem test failure, %v", target[1])
	}
	if target[2] != -3 {
		t.Errorf("Mvt.Elem test failure, %v", target[2])
	}
	m.Reset()
	if target[0] != 1 {
		t.Errorf("Mvt.Elem test failure, %v", target[0])
	}
	if target[1] != 2 {
		t.Errorf("Mvt.Elem test failure, %v", target[1])
	}
	if target[2] != 3 {
		t.Errorf("Mvt.Elem test failure, %v", target[2])
	}
}

func ExampleMvt_Elem() {
	target := []int{1, 2, 3}
	m := mvt.New()
	m.Elem(target, 0, -1).Elem(target, 1, -2).Elem(target, 2, -3)

	// now target is [-1, -2, -3]

	m.Reset()

	// now target is [1, 2, 3]
}

func TestMvt_FuncOuts(t *testing.T) {
	target := func(x int) (int, int) {
		return x, x
	}
	m := mvt.New()
	m.FuncOuts(&target, mvt.OutValues{
		{
			Values: []interface{}{1, 2},
			Times:  1,
		},
		{
			Values: []interface{}{3, 4},
			Times:  2,
		},
	})
	if x, y := target(1); x != 1 || y != 2 {
		t.Errorf("Mvt.FuncOuts test failure, %v %v", x, y)
	}
	if x, y := target(1); x != 3 || y != 4 {
		t.Errorf("Mvt.FuncOuts test failure, %v %v", x, y)
	}
	if x, y := target(1); x != 3 || y != 4 {
		t.Errorf("Mvt.FuncOuts test failure, %v %v", x, y)
	}
	if x, y := target(1); x != 1 || y != 1 {
		t.Errorf("Mvt.FuncOuts test failure, %v %v", x, y)
	}
	m.Reset()
	if x, y := target(1); x != 1 || y != 1 {
		t.Errorf("Mvt.FuncOuts test failure, %v %v", x, y)
	}
}

func ExampleMvt_FuncOuts() {
	target := func(x int) (int, int) {
		return x, x
	}
	m := mvt.New()
	m.FuncOuts(&target, mvt.OutValues{
		{
			Values: []interface{}{1, 2},
			Times:  1,
		},
		{
			Values: []interface{}{3, 4},
			Times:  2,
		},
	})

	target(1) // will return (1, 2)
	target(1) // will return (3, 4)
	target(1) // will return (3, 4)
	target(1) // will return (1, 1), call actual function
}

func TestMvt_Map(t *testing.T) {
	target := map[int]int{1: 1, 2: 2}
	m := mvt.New()
	m.Map(target, 1, 0).Map(target, 2, 0)
	if target[1] != 0 {
		t.Errorf("Mvt.Map test failure, %d", target[1])
	}
	if target[2] != 0 {
		t.Errorf("Mvt.Map test failure, %d", target[2])
	}
	if len(target) != 2 {
		t.Errorf("Mvt.Map test failure, %d", len(target))
	}
	m.Reset()
	if target[1] != 1 {
		t.Errorf("Mvt.Map test failure, %d", target[1])
	}
	if target[2] != 2 {
		t.Errorf("Mvt.Map test failure, %d", target[2])
	}
	if len(target) != 2 {
		t.Errorf("Mvt.Map test failure, %d", len(target))
	}
}

func ExampleMvt_Map() {
	target := map[int]int{1: 1, 2: 2}
	m := mvt.New()
	m.Map(target, 1, 0).Map(target, 2, 0)

	// target[1] = 0 and target[2] = 0

	m.Reset()

	// target[1] = 1 and target[2] = 2
}

func TestMvt_Path(t *testing.T) {
	type S struct {
		AString any
		ASlice  any
		AMap    any
		AArray  any
		aString any
	}
	x := 1
	target := [3]any{
		&S{
			AString: "zs",
			aString: "zs",
			ASlice:  []int{1},
			AMap:    map[int]string{1: "a"},
			AArray:  [1]*int{&x},
		},
	}
	m := mvt.New()

	m.Path(target, "0.AString", "ls")
	if target[0].(*S).AString != "ls" {
		t.Errorf("mvt.Path test failure, %v", target[0])
	}
	m.Reset()
	if target[0].(*S).AString != "zs" {
		t.Errorf("mvt.Path test failure, %v", target[0])
	}

	m.Path(target, "0.ASlice.0", -1)
	if target[0].(*S).ASlice.([]int)[0] != -1 {
		t.Errorf("mvt.Path test failure, %v", target[0])
	}
	m.Reset()
	if target[0].(*S).ASlice.([]int)[0] != 1 {
		t.Errorf("mvt.Path test failure, %v", target[0])
	}

	m.Path(target, "0.AMap.1", "b")
	if target[0].(*S).AMap.(map[int]string)[1] != "b" {
		t.Errorf("mvt.Path test failure, %v", target[0])
	}
	m.Reset()
	if target[0].(*S).AMap.(map[int]string)[1] != "a" {
		t.Errorf("mvt.Path test failure, %v", target[0])
	}

	m.Path(target, "0.AArray.0", 1)
	if *(target[0].(*S).AArray.([1]*int)[0]) != 1 {
		t.Errorf("mvt.Path test failure, %v", target[0])
	}
	m.Reset()
	if *(target[0].(*S).AArray.([1]*int)[0]) != 1 {
		t.Errorf("mvt.Path test failure, %v", target[0])
	}

	m.Path(target, "0.aString", "ls")
	if target[0].(*S).aString != "ls" {
		t.Errorf("mvt.Path test failure, %v", target[0])
	}
	m.Reset()
	if target[0].(*S).aString != "zs" {
		t.Errorf("mvt.Path test failure, %v", target[0])
	}
}

func ExampleMvt_Path() {
	type S struct {
		AString any
		ASlice  any
		AMap    any
		AArray  any
	}
	x := 1
	target := [3]any{
		&S{
			AString: "zs",
			ASlice:  []int{1},
			AMap:    map[int]string{1: "a"},
			AArray:  [1]*int{&x},
		},
	}
	m := mvt.New()

	m.Path(target, "0.AString", "ls")
	fmt.Println(target[0].(*S).AString) // print ls
	m.Reset()

	m.Path(target, "0.ASlice.0", -1)
	fmt.Println(target[0].(*S).ASlice.([]int)[0]) // print -1
	m.Reset()

	m.Path(target, "0.AMap.1", "b")
	fmt.Println(target[0].(*S).AMap.(map[int]string)[1]) // print "b"
	m.Reset()

	m.Path(target, "0.AArray.0", 1)
	fmt.Println(*(target[0].(*S).AArray.([1]*int)[0])) // print 1
	m.Reset()
}
