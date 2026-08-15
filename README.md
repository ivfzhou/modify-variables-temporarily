# 一、说明

临时修改变量值，用于单元测试中临时替换变量值以完成测试闭环。

[![codecov](https://codecov.io/gh/ivfzhou/modify-variables-temporarily/graph/badge.svg?token=QYBRAOTH5K)](https://codecov.io/gh/ivfzhou/modify-variables-temporarily)
[![Go Reference](https://pkg.go.dev/badge/gitee.com/ivfzhou/modify-variables-temporarily/v3.svg)](https://pkg.go.dev/gitee.com/ivfzhou/modify-variables-temporarily/v3)

# 二、特性

- 支持修改**任何类型**的变量：基本类型、指针、结构体（含未导出字段）、切片、数组、映射、接口、函数等
- 支持 **nil 指针 / nil 映射 / nil 接口**的自动初始化与回退
- 支持**链式导航**到深层嵌套字段进行修改
- 支持**函数返回值的 mock**（按调用次数依次返回预设值）
- 所有修改均可通过 `Reset()` **安全回退**到原始值
- 类型不兼容时自动尝试**类型转换**赋值

# 三、安装

```shell
go get gitee.com/ivfzhou/modify-variables-temporarily/v3@latest
```

```go
import mvt "gitee.com/ivfzhou/modify-variables-temporarily/v3"
```

# 四、快速开始

### 场景：修改其他包中未导出结构体的字段

当一个结构体是**未导出的**，且位于另一个包中无法直接修改时：

```go
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

---

# 五、API 一览

| API | 说明 |
|-----|------|
| [`Var`](#var-替换变量值) | 替换变量的值 |
| [`FieldByName`](#fieldbyname-按名称替换结构体字段) | 按名称替换结构体字段的值 |
| [`Field`](#field-按索引替换结构体字段) | 按索引替换结构体字段的值 |
| [`Elem`](#elem-替换切片元素值) | 替换切片元素的值 |
| [`Map`](#map-替换映射键值) | 替换映射中某个键的值 |
| [`FuncOuts`](#funcouts-mock-函数返回值) | Mock 函数返回值 |
| [`Chain`](#chain-链式导航) | 链式导航到深层嵌套变量并修改 |

所有 API 返回 `Resetter` 接口，调用 `Reset()` 即可回退修改。

---

# 六、简单 API

### Var — 替换变量值

```go
var target any
reset := mvt.Var(&target, "new value")
defer reset.Reset()
fmt.Println(target) // new value
// reset 后:
fmt.Println(target) // <nil>
```

### FieldByName — 按名称替换结构体字段

支持修改**未导出字段**：

```go
type MyStruct struct {
	name string // 未导出字段
}
var target = MyStruct{name: "old"}

reset := mvt.FieldByName(&target, "name", "new")
defer reset.Reset()
fmt.Println(target.name) // new
// reset 后:
fmt.Println(target.name) // old
```

### Field — 按索引替换结构体字段

支持负数索引（从末尾倒数）：

```go
var target struct {
	x int
	y string
}
reset := mvt.Field(&target, 0, 100)
// reset := mvt.Field(&target, -2, 100) // 同样指向第一个字段
defer reset.Reset()
fmt.Println(target.x) // 100
```

### Elem — 替换切片元素值

支持负数索引：

```go
var target = []any{1, "ok"}
reset := mvt.Elem(target, 1, "changed")
defer reset.Reset()
fmt.Println(target[1]) // changed
// reset 后:
fmt.Println(target[1]) // ok
```

### Map — 替换映射键值

```go
m := map[string]any{}
reset := mvt.Map(m, "key", "value")
defer reset.Reset()
fmt.Println(m["key"]) // value
// reset 后:
fmt.Println(m["key"]) // <nil>
```

### FuncOuts — Mock 函数返回值

将函数替换为按调用次数返回预设值的函数：

```go
fn := func() (any, error) { return nil, nil }

reset := mvt.FuncOuts(&fn, []mvt.OutValue{
	{Values: []any{"ok", errors.New("error occurred")}},
})
defer reset.Reset()

v, err := fn() // ("ok", "error occurred")
v, err = fn()  // 原始函数返回: (<nil>, <nil>)
```

#### OutValue 结构体

```go
type OutValue struct {
	Values []any // 函数调用的返回值列表
	Times  int   // 该组返回值的重复次数（0 或 1 表示仅一次）
```

示例 — 指定多次返回：

```go
outs := []mvt.OutValue{
	{Values: []any{1}},          // 第 1 次调用返回 1
	{Values: []any{"abc"}, Times: 2}, // 第 2~3 次调用返回 "abc"
	{Values: []any{nil}},         // 第 4 次调用返回 nil
}
```

> 当预设的返回值次数用尽后，后续调用将执行原始函数。

---

# 七、Chain — 链式导航

`Chain` 是最强大的 API，通过链式调用导航到任意深层的嵌套变量并进行修改。

### 核心接口

```go
Chainer {                    // 导航接口
    Elem() ChainSetter       // 指针/接口 → 内部值
    FieldByName(name)        // 结构体 → 字段（按名称）
    Field(index)             // 结构体 → 字段（按索引）
    MapValue(key)            // 映射 → 键对应的值
    Index(index)             // 切片/数组 → 元素
}

ChainSetter extends Chainer, Setter  // 同时具备导航和设置能力

Setter {                      // 设置接口
    Set(value) Resetter       // 替换值
    SetFuncOuts(outs) Resetter // Mock 函数返回值
}

Resetter {
    Reset()                   // 回退所有修改
}
```

### 使用示例

#### 1. 修改 nil 指针指向的未导出结构体字段

即使目标变量是 `nil` 指针，也可以自动初始化并修改：

```go
type aUnexportedStruct struct {
	name string
}
var Data *aUnexportedStruct // nil

reset := mvt.Chain(&Data).Elem().Elem().FieldByName("name").Set("any your want")
defer func() {
	reset.Reset()
	fmt.Println(Data) // <nil>
}()

fmt.Println(Data.name) // any your want
```

#### 2. 修改接口类型的值及内部实际类型

```go
type aInterface interface{}

type aImpl struct {
	name any
}

type aImpl2 struct {
	other any
}

var Data aInterface = aImpl{name: "impl1"}

// 替换接口内部的实际类型
impl2 := aImpl2{other: "impl2"}
reset := mvt.Chain(&Data).Elem().Set(impl2)
fmt.Println(Data.(aImpl2).other) // impl2
reset.Reset()
fmt.Println(Data.(aImpl).name)  // impl1

// 或深入修改接口内部类型的字段
reset2 := mvt.Chain(&Data).Elem().Elem().Field(0).Set("changed")
fmt.Println(Data.(aImpl).name)  // changed
reset2.Reset()
fmt.Println(Data.(aImpl).name)  // impl1
```

#### 3. 修改切片、数组、映射、函数等任意类型

即使 `map` 是 `nil`，仍然可以修改：

```go
type aUnexportedStruct struct {
	m   map[any]any
	arr [1]any
	s   []any
	fn  func()
	i   int
	str string
}
var Data *aUnexportedStruct

key := 1
reset := mvt.Chain(&Data).Elem().Elem().FieldByName("m").MapValue(key).Set("any your want")
defer func() {
	reset.Reset()
	fmt.Println(Data) // <nil>
}()

fmt.Println(Data.m[key]) // any your want
```

#### 4. 使用匿名结构体替代无法引用的类型

如果目标变量的类型无法在你的代码中直接引用：

```go
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

#### 5. 修改深层嵌套结构的字段

链路深度没有限制：

```go
type aUnexportedStruct struct {
	m map[any][1]struct {
		field string
	}
}
var Data *aUnexportedStruct

key := 1
reset := mvt.Chain(&Data).
	Elem().Elem().
	FieldByName("m").
	MapValue(key).
	Index(0).
	Field(0).
	Set("any your want")
defer func() {
	reset.Reset()
	fmt.Println(Data) // <nil>
}()

fmt.Println(Data.m[key][0].field) // any your want
```

#### 6. 多次独立修改同一数据源

可以创建多个独立的 `Resetter` 分别管理不同位置的修改：

```go
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

// 修改 #1: 深层嵌套的 field
reset := mvt.Chain(&Data).Elem().Elem().
	FieldByName("m").MapValue(key).Index(0).Field(0).Set(s)
defer func() {
	reset.Reset()
	fmt.Println(Data) // <nil>
}()
fmt.Println(Data.m[key][0].field[0]) // any your want

// 修改 #2: 顶层的 s 字段
reset2 := mvt.Chain(&Data).Elem().Elem().FieldByName("s").Set(s)
defer func() {
	reset2.Reset()
	fmt.Println(Data.s == nil) // true
}()
fmt.Println(Data.s[0]) // any your want

// 修改 #3: 另一个深层嵌套字段
reset3 := mvt.Chain(&Data).Elem().Elem().
	FieldByName("m").MapValue(key).Index(0).Field(1).Field(0).Set(s)
defer func() {
	reset3.Reset()
	fmt.Println(Data.m[key][0].field2.s == nil) // true
}()
fmt.Println(Data.m[key][0].field2.s[0]) // any your want
```

> 注意：修改的是**引用类型**（如 slice），多处共享同一底层数据。

#### 7. Mock 深层嵌套函数的返回值

无需声明新函数，直接指定返回值序列即可：

```go
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
	{Values: []any{1}},           // 第 1 次
	{Values: []any{"abc"}, Times: 2}, // 第 2~3 次
	{Values: []any{nil}},         // 第 4 次
}
reset := mvt.Chain(&Data).Elem().Elem().
	FieldByName("m").MapValue(key).Index(0).
	Field(0).FieldByName("fn").
	SetFuncOuts(outs)
defer func() {
	reset.Reset()
	fmt.Println(Data) // <nil>
}()

for range 4 {
	fmt.Println(Data.m[key][0].field.fn(nil))
	// 输出: 1 abc abc <nil>
}
```

#### 8. 预设返回值耗尽后自动恢复原函数

```go
type aUnexportedStruct struct {
	fn func(any) any
}
var Data = &aUnexportedStruct{fn: func(any) any {
	return "nothing"
}}

outs := []mvt.OutValue{
	{Values: []any{1}},
	{Values: []any{"abc"}, Times: 2},
	{Values: []any{nil}},
}
defer mvt.Chain(&Data).Elem().Elem().FieldByName("fn").SetFuncOuts(outs).Reset()

for range 4 {
	fmt.Println(Data.fn(nil))
	// 输出: 1 abc abc <nil>
}

// 预设值耗尽后运行原函数（注意：若原函数为 nil 会触发空指针异常）
fmt.Println(Data.fn(nil)) // nothing
```

---

## 错误处理

当传入非法参数时，API 会 panic 并附带明确的错误信息。常见错误如下：

| 错误 | 触发条件 |
|------|----------|
| `ErrTargetCannotBeNil` | target 为 nil |
| `ErrTargetCannotBeNilType` | target 有类型但无值（如 nil 指针） |
| `ErrTargetIsNotPointer` | target 不是指针类型 |
| `ErrTargetIsNotStruct` | target 不是结构体 |
| `ErrTargetIsNotSlice` | target 不是切片 |
| `ErrTargetIsNotMap` | target 不是映射 |
| `ErrTargetIsNotFunc` | target 不是函数 |
| `ErrIncompatibleTypeAssignment` | 赋值类型不兼容且无法转换 |
| `ErrStructFieldNameCannotBeEmpty` | 字段名为空字符串 |
| `ErrStructFieldNotFound` | 结构体不存在该字段 |
| `ErrIndexOutOfBound` | 数组/切片索引越界 |
| `ErrInvalidMapKeyType` | 映射键类型不兼容 |
