package modify_variables_temporarily_test

import (
	"reflect"
	"testing"

	"gitee.com/ivfzhou/modify-variables-temporarily"
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
