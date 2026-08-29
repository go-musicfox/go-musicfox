package keybindings

import (
	"reflect"
	"strings"
	"testing"
)

// TestRegisterOperateAllocatesUniqueIDs verifies RegisterOperate returns
// unique, monotonically increasing ids in the plugin range, and that
// Name()/Desc() resolve through the authoritative registry.
func TestRegisterOperateAllocatesUniqueIDs(t *testing.T) {
	opA := RegisterOperate("test_op_alloc_a", "测试操作A", nil)
	opB := RegisterOperate("test_op_alloc_b", "测试操作B", []string{"ctrl+1"})

	if opA < pluginOperateStart || opB < pluginOperateStart {
		t.Fatalf("plugin ops = %d, %d; both must be >= pluginOperateStart(%d)", opA, opB, pluginOperateStart)
	}
	if opA == opB {
		t.Fatalf("plugin ops must be unique: both = %d", opA)
	}
	if opA >= opB {
		t.Fatalf("plugin op ids must increase: opA=%d, opB=%d", opA, opB)
	}
	if got := opA.Name(); got != "test_op_alloc_a" {
		t.Errorf("opA.Name() = %q, want %q", got, "test_op_alloc_a")
	}
	if got := opA.Desc(); got != "测试操作A" {
		t.Errorf("opA.Desc() = %q, want %q", got, "测试操作A")
	}
	if got := opB.Name(); got != "test_op_alloc_b" {
		t.Errorf("opB.Name() = %q, want %q", got, "test_op_alloc_b")
	}
	if got := opB.Desc(); got != "测试操作B" {
		t.Errorf("opB.Desc() = %q, want %q", got, "测试操作B")
	}
}

// TestRegisterOperateDuplicateNamePanics verifies a name conflict panics
// (programmer error), regardless of the op id.
func TestRegisterOperateDuplicateNamePanics(t *testing.T) {
	RegisterOperate("test_op_dup", "重复操作", nil)
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("RegisterOperate(duplicate name) did not panic")
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, "test_op_dup") {
			t.Fatalf("panic message = %v, want it to mention the duplicate name", r)
		}
	}()
	RegisterOperate("test_op_dup", "重复操作", nil)
}

// TestRegisterOperateDefaultKeysInInitDefaults verifies plugin default keys
// surface through InitDefaults(true) (merged from defaultOtherOperateToKeys).
func TestRegisterOperateDefaultKeysInInitDefaults(t *testing.T) {
	op := RegisterOperate("test_op_defkeys", "默认键操作", []string{"ctrl+9", "alt+x"})

	bindings := InitDefaults(true)
	keys, ok := bindings[op]
	if !ok {
		t.Fatalf("InitDefaults(true) missing plugin op %d", op)
	}
	want := []string{"ctrl+9", "alt+x"}
	if !reflect.DeepEqual(keys, want) {
		t.Fatalf("InitDefaults(true)[op] = %v, want %v", keys, want)
	}

	// useDefault=false 只有 base，不含插件默认键（非默认模式下用户需自行配置）。
	bindings = InitDefaults(false)
	if _, ok := bindings[op]; ok {
		t.Fatalf("InitDefaults(false) must not contain plugin op %d", op)
	}
}

// TestGetOperationFromNameResolvesPluginOp verifies user-config parsing
// (ProcessUserBindings) can resolve a plugin operation name.
func TestGetOperationFromNameResolvesPluginOp(t *testing.T) {
	op := RegisterOperate("test_op_resolve", "可解析操作", nil)

	got, ok := GetOperationFromName("test_op_resolve")
	if !ok {
		t.Fatal("GetOperationFromName(plugin name) = not found")
	}
	if got != op {
		t.Fatalf("GetOperationFromName = %d, want %d", got, op)
	}

	// ProcessUserBindings 也能解析并绑定插件操作名。
	processed := ProcessUserBindings(map[string][]string{"test_op_resolve": {"ctrl+r"}})
	keys, ok := processed[op]
	if !ok {
		t.Fatal("ProcessUserBindings did not resolve plugin operation name")
	}
	if !reflect.DeepEqual(keys, []string{"ctrl+r"}) {
		t.Fatalf("ProcessUserBindings[op] = %v, want [ctrl+r]", keys)
	}
}

// TestRegisteredOperatesSnapshot verifies RegisteredOperates returns plugin ops
// (id >= pluginOperateStart) including newly registered ones.
func TestRegisteredOperatesSnapshot(t *testing.T) {
	before := RegisteredOperates()
	op := RegisterOperate("test_op_snapshot", "快照操作", nil)
	after := RegisteredOperates()

	if len(after) <= len(before) {
		t.Fatalf("RegisteredOperates() length = %d, want > %d", len(after), len(before))
	}
	found := false
	for _, registered := range after {
		if registered < pluginOperateStart {
			t.Fatalf("RegisteredOperates() returned built-in op %d", registered)
		}
		if registered == op {
			found = true
		}
	}
	if !found {
		t.Fatalf("RegisteredOperates() missing newly registered op %d", op)
	}
}
