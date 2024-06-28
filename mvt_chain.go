package modify_variables_temporarily

import (
	"fmt"
	"reflect"
	"sync"
	"unicode"
	"unsafe"
)

type Reseter interface {
	Reset()
}

type resetFn func()

type MvtChain struct {
	lock    sync.Mutex
	used    bool
	pTarget reflect.Value
	mapKey  reflect.Value
	target  reflect.Value
}

func NewMvtChain(target interface{}) *MvtChain {
	val, _ := indirect(target)
	return &MvtChain{target: val}
}

func (mc *MvtChain) Field(index uint) *MvtChain {
	mc.lock.Lock()
	defer mc.lock.Unlock()
	if mc.used {
		panic("mvt is used")
	}

	for mc.target.Kind() == reflect.Ptr || mc.target.Kind() == reflect.Interface {
		mc.target = mc.target.Elem()
	}
	if mc.target.Kind() != reflect.Struct {
		panic(fmt.Sprintf("the target is not a struct: [%s]", mc.target.Type()))
	}
	field := mc.target.Field(int(index))
	fieldType := mc.target.Type().Field(int(index))
	if !unicode.IsUpper(rune(fieldType.Name[0])) {
		field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
	}
	mc.target = field
	return mc
}

func (mc *MvtChain) FieldByName(name string) *MvtChain {
	mc.lock.Lock()
	defer mc.lock.Unlock()
	if mc.used {
		panic("mvt is used")
	}

	for mc.target.Kind() == reflect.Ptr || mc.target.Kind() == reflect.Interface {
		mc.target = mc.target.Elem()
	}
	if mc.target.Kind() != reflect.Struct {
		panic(fmt.Sprintf("the target is not a struct: [%s]", mc.target.Type()))
	}
	field := mc.target.FieldByName(name)
	if !unicode.IsUpper(rune(name[0])) {
		field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
	}
	mc.target = field
	return mc
}

func (mc *MvtChain) Elem(index uint) *MvtChain {
	mc.lock.Lock()
	defer mc.lock.Unlock()
	if mc.used {
		panic("mvt is used")
	}

	for mc.target.Kind() == reflect.Ptr || mc.target.Kind() == reflect.Interface {
		mc.target = mc.target.Elem()
	}
	if mc.target.Kind() != reflect.Array && mc.target.Kind() != reflect.Slice {
		panic(fmt.Sprintf("the target is not an array or slice: [%s]", mc.target.Type()))
	}
	val := mc.target.Index(int(index))
	mc.target = val
	return mc
}

func (mc *MvtChain) Map(key any) *MvtChain {
	mc.lock.Lock()
	defer mc.lock.Unlock()
	if mc.used {
		panic("mvt is used")
	}

	for mc.target.Kind() == reflect.Ptr || mc.target.Kind() == reflect.Interface {
		mc.target = mc.target.Elem()
	}
	if mc.target.Kind() != reflect.Map {
		panic(fmt.Sprintf("the target is not a map: [%s]", mc.target.Type()))
	}
	kv := reflect.ValueOf(key)
	val := mc.target.MapIndex(kv)
	mc.pTarget = mc.target
	mc.mapKey = kv
	mc.target = val
	return mc
}

func (mc *MvtChain) Set(substitute interface{}) Reseter {
	mc.lock.Lock()
	defer mc.lock.Unlock()
	if mc.used {
		panic("mvt is used")
	}
	mc.used = true

	value := reflect.ValueOf(substitute)
	if mc.target.CanSet() {
		old := reflect.ValueOf(mc.target.Interface())
		mc.target.Set(value.Convert(mc.target.Type()))
		return resetFn(func() {
			mc.target.Set(old)
		})
	}

	if mc.pTarget.Kind() == reflect.Map {
		mc.pTarget.SetMapIndex(mc.mapKey, value)
		return resetFn(func() {
			mc.pTarget.SetMapIndex(mc.mapKey, mc.target)
		})
	}

	if mc.target.Kind() == reflect.Ptr || mc.target.Kind() == reflect.Interface {
		mc.target = mc.target.Elem()
		for mc.target.Kind() == reflect.Ptr || mc.target.Kind() == reflect.Interface {
			mc.target = mc.target.Elem()
		}
		if mc.target.CanSet() {
			old := reflect.ValueOf(mc.target.Interface())
			mc.target.Set(value.Convert(mc.target.Type()))
			return resetFn(func() {
				mc.target.Set(old)
			})
		}
	}

	panic(fmt.Sprintf("the target cannot set: [%s]", mc.target.Type()))
}

func (r resetFn) Reset() {
	r()
}
