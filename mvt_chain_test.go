package modify_variables_temporarily_test

import (
	"fmt"
	"testing"

	mvt "gitee.com/ivfzhou/modify-variables-temporarily/v2"
)

func TestMvtChain_Set(t *testing.T) {
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

	reset := mvt.NewMvtChain(target).Elem(0).FieldByName("AString").Set("ls")
	if target[0].(*S).AString != "ls" {
		t.Errorf("mvt.Path test failure, %v", target[0])
	}
	reset.Reset()
	if target[0].(*S).AString != "zs" {
		t.Errorf("mvt.Path test failure, %v", target[0])
	}

	reset = mvt.NewMvtChain(target).Elem(0).FieldByName("ASlice").Elem(0).Set(-1)
	if target[0].(*S).ASlice.([]int)[0] != -1 {
		t.Errorf("mvt.Path test failure, %v", target[0])
	}
	reset.Reset()
	if target[0].(*S).ASlice.([]int)[0] != 1 {
		t.Errorf("mvt.Path test failure, %v", target[0])
	}

	reset = mvt.NewMvtChain(target).Elem(0).FieldByName("AMap").Map(1).Set("b")
	if target[0].(*S).AMap.(map[int]string)[1] != "b" {
		t.Errorf("mvt.Path test failure, %v", target[0])
	}
	reset.Reset()
	if target[0].(*S).AMap.(map[int]string)[1] != "a" {
		t.Errorf("mvt.Path test failure, %v", target[0])
	}

	reset = mvt.NewMvtChain(target).Elem(0).FieldByName("AArray").Elem(0).Set(1)
	if *(target[0].(*S).AArray.([1]*int)[0]) != 1 {
		t.Errorf("mvt.Path test failure, %v", target[0])
	}
	reset.Reset()
	if *(target[0].(*S).AArray.([1]*int)[0]) != 1 {
		t.Errorf("mvt.Path test failure, %v", target[0])
	}

	reset = mvt.NewMvtChain(target).Elem(0).FieldByName("aString").Set("ls")
	if target[0].(*S).aString != "ls" {
		t.Errorf("mvt.Path test failure, %v", target[0])
	}
	reset.Reset()
	if target[0].(*S).aString != "zs" {
		t.Errorf("mvt.Path test failure, %v", target[0])
	}
}

func ExampleMvtChain_Set() {
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

	reset := mvt.NewMvtChain(target).Elem(0).FieldByName("AString").Set("ls")
	fmt.Println(target[0].(*S).AString) // print "ls"
	reset.Reset()

	reset = mvt.NewMvtChain(target).Elem(0).FieldByName("ASlice").Elem(0).Set(-1)
	fmt.Println(target[0].(*S).ASlice.([]int)[0]) // print -1
	reset.Reset()

	reset = mvt.NewMvtChain(target).Elem(0).FieldByName("AMap").Map(1).Set("b")
	fmt.Println(target[0].(*S).AMap.(map[int]string)[1]) // print b
	reset.Reset()

	reset = mvt.NewMvtChain(target).Elem(0).FieldByName("AArray").Elem(0).Set(1)
	fmt.Println(*(target[0].(*S).AArray.([1]*int)[0])) // print 1
	reset.Reset()

	reset = mvt.NewMvtChain(target).Elem(0).FieldByName("aString").Set("ls")
	fmt.Println(target[0].(*S).aString) // print ls
	reset.Reset()
}
