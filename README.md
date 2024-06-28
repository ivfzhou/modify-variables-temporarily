### 说明

MVT 临时地修改变量值。用于单元测试中，临时替换某一变量值，以方便测试闭环。

让任何可以寻址的变量都可以在测试代码里方便地替换成其它值。

可寻址的变量有结构体指针、映射（map）、切片（slice）、接口（其包含的值是可寻址的）、数组（元素为可寻址的）。

大都数情况下我们要替换的变量是某个结构体指针变量的一个字段，mvt 能链式代码风格替换这些值，之后自动还原。

### 使用

```shell
go get gitee.com/ivfzhou/modify-variables-temporarily/v2@latest
```

#### 替换指针变量

```golang
target := 1
m := mvt.New()
m.Var(&target, 2)

// now x is 2

m.Reset()

// now x is 1
```

#### 替换结构体字段值

```golang
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

// now target.Petname is ls, and target.name is zs

m.Reset()

// now target.Petname is ww, and target.name is zl

// also mvt by index
m.Field(target, 1, "ww").Field(target, 0, "zl")
```

#### 替换切换中的元素

```golang
target := []int{1, 2, 3}
m := mvt.New()
m.Elem(target, 0, -1).Elem(target, 1, -2).Elem(target, 2, -3)

// now target is [-1, -2, -3]

m.Reset()

// now target is [1, 2, 3]
```

#### 替换函数返回值

```golang
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
```

#### 替换映射中的值

```golang
target := map[int]int{1: 1, 2: 2}
m := mvt.New()
m.Map(target, 1, 0).Map(target, 2, 0)

// target[1] = 0 and target[2] = 0

m.Reset()

// target[1] = 1 and target[2] = 2
```

#### 链式替换

```golang
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
```

#### 另一种风格的链式替换

```golang
reset := mvt.NewMvtChain(target).Elem(0).FieldByName("AString").Set("ls")
fmt.Println(target[0].(*S).AString) // print ls
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
```

联系电邮：ivfzhou@126.com
