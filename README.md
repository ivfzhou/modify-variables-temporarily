# 一、说明

临时地修改变量值，用于单元测试中，临时替换某一变量值，以方便测试闭环。

[![codecov](https://codecov.io/gh/ivfzhou/modify-variables-temporarily/graph/badge.svg?token=QYBRAOTH5K)](https://codecov.io/gh/ivfzhou/modify-variables-temporarily)
[![Go Reference](https://pkg.go.dev/badge/gitee.com/ivfzhou/modify-variables-temporarily/v3.svg)](https://pkg.go.dev/gitee.com/ivfzhou/modify-variables-temporarily/v3)

# 二、使用

```shell
go get gitee.com/ivfzhou/modify-variables-temporarily/v3@latest
```

```golang
import mvt "gitee.com/ivfzhou/modify-variables-temporarily/v3"
```

当一个结构体是不导出的，而它又在另一个包，不方便修改它的值时：
```golang
type aUnexportedStruct struct {
	name string
}
var Data = &aUnexportedStruct{name: "something"}

reset := mvt.Chain(&Data).Elem().Elem().Field(0).Set("any your want")
defer func() {
	reset.Reset()
	fmt.Println(Data.name) // something
}()

fmt.Println(Data.name) // any your want
```

也许这个 Data 是一个 nil 指针，也可以修改：
```golang
type aUnexportedStruct struct {
	name string
}
var Data *aUnexportedStruct

reset := mvt.Chain(&Data).Elem().Elem().FieldByName("name").Set("any your want")
defer func() {
	reset.Reset()
	fmt.Println(Data) // <nil>
}()

fmt.Println(Data.name) // any your want
```

也可以修改接口类型的值，以及它内部实际类型的值：
```golang
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
fmt.Println(Data.(aImpl2).other) // impl2
reset.Reset()
fmt.Println(Data.(aImpl).name) // impl1

reset2 := mvt.Chain(&Data).Elem().Elem().Field(0).Set("changed")
fmt.Println(Data.(aImpl).name) // changed
reset2.Reset()
fmt.Println(Data.(aImpl).name) // impl1
```

当然也可以修改切片、数组、映射、函数等任何类型的值：
```golang
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
reset := mvt.Chain(&Data).Elem().Elem().FieldByName("m").MapValue(key).Set("any your want") // 尽管 map 是 nil，仍然可以修改。
defer func() {
	reset.Reset()
	fmt.Println(Data) // <nil>
}()

fmt.Println(Data.m[key]) // any your want
```

如果要修改的变量，它的类型不能在你的代码中直接引用，我们也可以使用它的实际类型变量去修改它：
```golang
type aUnexportedStruct struct {
	name string
}
var Data *aUnexportedStruct

reset := mvt.Chain(&Data).Elem().Elem().Set(struct{ name string }{"any your want"})
defer func() {
	reset.Reset()
	fmt.Println(Data) // <nil>
}()

fmt.Println(Data.name) // any your want
```

有时候，要修改的值在更深层次的结构中，这也没问题：
```golang
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
	fmt.Println(Data) // <nil>
}()

fmt.Println(Data.m[key][0].field) // any your want
```

有时候，仅一次修改不满足要求。那么，可以进行多次修改：
```golang
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
	fmt.Println(Data) // <nil>
}()
fmt.Println(Data.m[key][0].field[0]) // any your want

reset2 := mvt.Chain(&Data).Elem().Elem().FieldByName("s").Set(s)
defer func() {
	reset2.Reset()
	fmt.Println(Data.s == nil) // true
}()
fmt.Println(Data.s[0]) // any your want

reset3 := mvt.Chain(&Data).Elem().Elem().FieldByName("m").MapValue(key).Index(0).Field(1).Field(0).Set(s)
defer func() {
	reset3.Reset()
	fmt.Println(Data.m[key][0].field2.s == nil) // true
}()
fmt.Println(Data.m[key][0].field2.s[0]) // any your want

s[0] = "I changed"
fmt.Println(Data.m[key][0].field[0])    // I changed
fmt.Println(Data.s[0])                  // I changed
fmt.Println(Data.m[key][0].field2.s[0]) // I changed
```

如果要修改的变量是一个函数，你不想声明一个函数去修改它，而仅是想修改它的返回值，可以这样做：
```golang
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
	fmt.Println(Data) // <nil>
}()

for range 4 {
	fmt.Println(Data.m[key][0].field.fn(nil)) // 1 abc abc <nil>
}
```

如果函数返回值次数用尽了，那么将运行原本的函数变量：
```golang
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
	fmt.Println(Data.fn(nil)) // 1 abc abc <nil>
}

// 注意，如果 Data.fn 是 nil，将触发空指针异常。
fmt.Println(Data.fn(nil)) // nothing
```
