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
	"testing"

	mvt "gitee.com/ivfzhou/modify-variables-temporarily/v3"
)

// TestFuncOutsEmptyValues 回归：OutValue.Values 为空切片时不应 panic，
// 缺失的返回值应补零值。
func TestFuncOutsEmptyValues(t *testing.T) {
	fn := func() (int, string) { return 1, "origin" }
	reset := mvt.FuncOuts(&fn, []mvt.OutValue{{Values: []any{}}})
	a, b := fn()
	if a != 0 || b != "" {
		t.Errorf("empty Values should be padded with zero values, got (%v, %q)", a, b)
	}
	a, b = fn()
	if a != 1 || b != "origin" {
		t.Errorf("after preset exhausted should call original, got (%v, %q)", a, b)
	}
	reset.Reset()
}

// TestFuncOutsZeroOutputs 无返回值的函数也应能正常 mock。
func TestFuncOutsZeroOutputs(t *testing.T) {
	fn := func() {}
	reset := mvt.FuncOuts(&fn, []mvt.OutValue{{}, {Times: 2}})
	// 不应 panic。
	fn()
	fn()
	fn()
	reset.Reset()
}

// TestFuncOutsNegativeTimes 负数 Times 应被当作 1 次处理。
func TestFuncOutsNegativeTimes(t *testing.T) {
	fn := func() int { return 99 }
	reset := mvt.FuncOuts(&fn, []mvt.OutValue{{Values: []any{7}, Times: -5}})
	if v := fn(); v != 7 {
		t.Errorf("first call should return preset 7, got %d", v)
	}
	if v := fn(); v != 99 {
		t.Errorf("preset exhausted should call original 99, got %d", v)
	}
	reset.Reset()
}

// TestChainBranching 回归：复用同一个 Chainer 分支导航时不应相互污染。
func TestChainBranching(t *testing.T) {
	data := map[any]any{}
	c := mvt.Chain(&data).Elem()
	c.MapValue("a").Set(1)
	c.MapValue("b").Set(2)
	if data["a"] != 1 {
		t.Errorf("branching: data[a] = %v, want 1", data["a"])
	}
	if data["b"] != 2 {
		t.Errorf("branching: data[b] = %v, want 2", data["b"])
	}
	if len(data) != 2 {
		t.Errorf("branching: len(data) = %d, want 2", len(data))
	}
}

// TestChainMapNilValueRestore 回归：映射中「存在键且值为 nil」与「键不存在」
// 应被区分对待，回退后键仍需保留。
func TestChainMapNilValueRestore(t *testing.T) {
	data := map[any]any{"key": nil}
	reset := mvt.Chain(&data).Elem().MapValue("key").Set(123)
	if data["key"] != 123 {
		t.Errorf("after set: data[key] = %v, want 123", data["key"])
	}
	reset.Reset()
	v, ok := data["key"]
	if !ok {
		t.Errorf("after reset: key should still be present with nil value")
	}
	if v != nil {
		t.Errorf("after reset: data[key] = %v, want nil", v)
	}
}

// TestMapNilValueRestore 顶层 Map 对「存在键且值为 nil」的回退也应保留键。
func TestMapNilValueRestore(t *testing.T) {
	data := map[any]any{"key": nil}
	reset := mvt.Map(data, "key", 123)
	if data["key"] != 123 {
		t.Errorf("after set: data[key] = %v, want 123", data["key"])
	}
	reset.Reset()
	v, ok := data["key"]
	if !ok {
		t.Errorf("after reset: key should still be present with nil value")
	}
	if v != nil {
		t.Errorf("after reset: data[key] = %v, want nil", v)
	}
}

// TestVarNilSubstitute 顶层 Var 使用 nil 替换时，非指针类型应被置为零值。
func TestVarNilSubstitute(t *testing.T) {
	i := 42
	reset := mvt.Var(&i, nil)
	if i != 0 {
		t.Errorf("nil substitute should set zero value, got %d", i)
	}
	reset.Reset()
	if i != 42 {
		t.Errorf("after reset should restore original, got %d", i)
	}
}

// TestResetterIdempotent 多次调用 Reset 应安全且只回退一次。
func TestResetterIdempotent(t *testing.T) {
	i := 42
	reset := mvt.Var(&i, 100)
	if i != 100 {
		t.Fatalf("set failed, got %d", i)
	}
	reset.Reset()
	reset.Reset()
	if i != 42 {
		t.Errorf("after reset should restore original, got %d", i)
	}
}

func TestChainErrors(t *testing.T) {
	t.Run("Target 是 nil", func(t *testing.T) {
		assertPanic(t, mvt.ErrTargetCannotBeNil, func() {
			mvt.Chain(nil)
		})
	})

	t.Run("Target 有类型但无值", func(t *testing.T) {
		var p *int
		assertPanic(t, mvt.ErrTargetCannotBeNilType, func() {
			mvt.Chain(p)
		})
	})

	t.Run("Target 不可设置", func(t *testing.T) {
		assertPanic(t, mvt.ErrTargetCannotBeSet, func() {
			mvt.Chain(123)
		})
	})

	t.Run("无导航动作直接 Set", func(t *testing.T) {
		var x int
		assertPanic(t, mvt.ErrNoActions, func() {
			mvt.Chain(&x).(mvt.ChainSetter).Set(1)
		})
	})

	t.Run("无导航动作直接 SetFuncOuts", func(t *testing.T) {
		var x int
		assertPanic(t, mvt.ErrNoActions, func() {
			mvt.Chain(&x).(mvt.ChainSetter).SetFuncOuts([]mvt.OutValue{})
		})
	})

	t.Run("SetFuncOuts 目标不是函数", func(t *testing.T) {
		var x any = 123
		assertPanic(t, mvt.ErrTargetIsNotFunc, func() {
			mvt.Chain(&x).Elem().SetFuncOuts([]mvt.OutValue{{Values: []any{1}}})
		})
	})

	t.Run("结构体字段索引越界", func(t *testing.T) {
		var target testStruct
		assertPanic(t, mvt.ErrStructFieldNotFound, func() {
			mvt.Field(&target, -999, testImpl(0))
		})
	})

	t.Run("切片元素索引越界", func(t *testing.T) {
		target := []any{1, 2, 3}
		assertPanic(t, mvt.ErrIndexOutOfBound, func() {
			mvt.Elem(target, 3, 0)
		})
	})
}

func TestChainNavigateErrors(t *testing.T) {
	t.Run("Elem 目标是标量", func(t *testing.T) {
		i := 1
		assertPanic(t, mvt.ErrTargetIsNotPointerOrInterface, func() {
			mvt.Chain(&i).Elem().Elem().Set(0)
		})
	})

	t.Run("Elem 目标是 nil 接口", func(t *testing.T) {
		var x any
		assertPanic(t, mvt.ErrCannotToNext, func() {
			mvt.Chain(&x).Elem().Elem().Set(0)
		})
	})

	t.Run("FieldByName 名称是空串", func(t *testing.T) {
		var target testStruct
		assertPanic(t, mvt.ErrStructFieldNameCannotBeEmpty, func() {
			mvt.Chain(&target).Elem().FieldByName("").Set(0)
		})
	})

	t.Run("MapValue 目标不是 map", func(t *testing.T) {
		var target testStruct
		assertPanic(t, mvt.ErrTargetIsNotMap, func() {
			mvt.Chain(&target).Elem().MapValue(1).Set(0)
		})
	})

	t.Run("Field 目标不是结构体", func(t *testing.T) {
		var x any = 123
		assertPanic(t, mvt.ErrTargetIsNotStruct, func() {
			mvt.Chain(&x).Elem().Field(0).Set(0)
		})
	})

	t.Run("Index 目标不是序列", func(t *testing.T) {
		var x any = 123
		assertPanic(t, mvt.ErrTargetIsNotSliceOrArray, func() {
			mvt.Chain(&x).Elem().Index(0).Set(0)
		})
	})
}

// TestChainInterfaceNilPointer 接口内部持有 nil 指针时，Elem 应能自动分配内存并正确回退。
func TestChainInterfaceNilPointer(t *testing.T) {
	var p *int
	var x any = p
	reset := mvt.Chain(&x).Elem().Elem().Elem().Set(42)
	got, ok := x.(*int)
	if !ok || got == nil || *got != 42 {
		t.Errorf("expected x to be *int(42), got %#v", x)
	}
	reset.Reset()
	got, ok = x.(*int)
	if !ok || got != nil {
		t.Errorf("after reset x should be (*int)(nil), got %#v", x)
	}
}

// TestChainInterfaceNilMap 接口内部持有 nil map 时，应能初始化并修改键值且正确回退。
func TestChainInterfaceNilMap(t *testing.T) {
	var m map[any]any
	var x any = m
	reset := mvt.Chain(&x).Elem().Elem().MapValue("k").Set(7)
	got, ok := x.(map[any]any)
	if !ok || got["k"] != 7 {
		t.Errorf("expected x to be map{k:7}, got %#v", x)
	}
	reset.Reset()
	got, ok = x.(map[any]any)
	if !ok || got != nil {
		t.Errorf("after reset x should be nil map, got %#v", x)
	}
}

// TestChainMapNilPointer 映射值为 nil 指针时，应能自动分配内存并正确回退。
func TestChainMapNilPointer(t *testing.T) {
	data := map[any]*int{"k": nil}
	reset := mvt.Chain(&data).Elem().MapValue("k").Elem().Set(5)
	if data["k"] == nil || *data["k"] != 5 {
		t.Errorf("expected data[k] = *int(5), got %#v", data["k"])
	}
	reset.Reset()
	if data["k"] != nil {
		t.Errorf("after reset data[k] should be nil, got %#v", data["k"])
	}
}

// assertPanic 断言 fn 会 panic 且 panic 值通过 errors.Is 匹配 wantErr。
func assertPanic(t *testing.T, wantErr error, fn func()) {
	t.Helper()
	defer func() {
		recovered, _ := recover().(error)
		if !errors.Is(recovered, wantErr) {
			t.Errorf("expected panic wrapping %v, got %v", wantErr, recovered)
		}
	}()
	fn()
}
