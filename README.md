## MVT(Modify variables temporarily) 意为临时地修改变量。用于单元测试。

## 说明

由于 Gomock 桩代码变量需要开发者自行设置到业务代码对应接口变量，此项目是为了方便快捷的修改变量而开发。

目的让任何可以寻址的变量都可以在测试代码里方便地替换成其它值。

可寻址的变量有结构体指针、映射（map）、切片（slice）、接口（其包含的值是可寻址的）、数组（元素为可寻址的）。

大都数情况下我们要 mock 的值是某个结构体指针变量的一个字段，该项目能链式代码风格替换这些值，之后自动还原。

## 用法

### mvt.Path(target interface{}, path string, substitute interface{})

```golang
var arr [1]*Struct{field}
defer mvt.New().Path(arr, "0.field", value).Reset()

var m map[string]Struct{filed}
defer mvt.New().Path(m, "key", value).Reset()
defer mvt.New().Path(m, "key.field", value).Reset()

var awriter *strcut{ otherWriter io.Writer}
defer mvt.New().Path(awriter, "otherWriter", value).Reset()
```

### FuncOuts(target interface{}, outs OutValues)

```golang
var fn func ()string
mvt.New().FuncOuts(fn, mvt.OutValues{{Times: []OutValue{"hello"}}}).Reset()
```

## 联系

wxid: zivfzhou

email: ivfzhou@aliyun.com
