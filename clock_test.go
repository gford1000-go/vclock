package vclock

import (
	"errors"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {

	v, err := New(nil)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v.Close()

	m, err := v.GetMap()
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if len(m) != 0 {
		t.Fatal("unexpected non-zero map")
	}
}

func TestNewWithInit(t *testing.T) {

	init := map[string]uint64{"a": 1, "b": 14}
	v, err := New(init)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v.Close()

	m, err := v.GetMap()
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if len(m) != 2 {
		t.Fatal("unexpected map")
	}

	if !reflect.DeepEqual(init, m) {
		t.Fatal("maps not equal")
	}
}

func TestUseAfterClose(t *testing.T) {

	init := map[string]uint64{"a": 1, "b": 14}
	v, err := New(init)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	v.Close()

	_, err = v.GetMap()
	if err == nil {
		t.Fatal("unexpected pass when expected error")
	} else {
		if !errors.Is(err, errClosedVClock) {
			t.Fatalf("unexpected error %q\n", err.Error())
		}
	}
}

func TestDoubleClose(t *testing.T) {

	init := map[string]uint64{"a": 1, "b": 14}
	v, err := New(init)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	v.Close()

	err = v.Close()
	if err == nil {
		t.Fatal("unexpected pass when expected error")
	} else {
		if !errors.Is(err, errClosedVClock) {
			t.Fatalf("unexpected error %q\n", err.Error())
		}
	}
}

func TestSetExisting(t *testing.T) {

	init := map[string]uint64{"a": 1, "b": 14}
	v, err := New(init)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v.Close()

	err = v.Set("a", 4)
	if err == nil {
		t.Fatal("unexpected pass when expected error")
	} else {
		if !errors.Is(err, errAttemptToSetExistingId) {
			t.Fatalf("unexpected error %q\n", err.Error())
		}
	}
}

func TestSetBlank(t *testing.T) {

	init := map[string]uint64{"a": 1, "b": 14}
	v, err := New(init)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v.Close()

	err = v.Set("", 4)
	if err == nil {
		t.Fatal("unexpected pass when expected error")
	} else {
		if !errors.Is(err, errClockIdMustNotBeEmptyString) {
			t.Fatalf("unexpected error %q\n", err.Error())
		}
	}
}

func TestSetNewID(t *testing.T) {

	init := map[string]uint64{"a": 1, "b": 14}
	v, err := New(init)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v.Close()

	err = v.Set("c", 4)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	init["c"] = 4

	m, err := v.GetMap()
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if len(m) != len(init) {
		t.Fatal("unexpected map")
	}

	if !reflect.DeepEqual(init, m) {
		t.Fatal("maps not equal")
	}
}

func TestSetNewIDAfterClose(t *testing.T) {

	init := map[string]uint64{"a": 1, "b": 14}
	v, err := New(init)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	v.Close()

	err = v.Set("c", 4)
	if err == nil {
		t.Fatal("expected error but didn't get one")
	} else {
		if err != errClosedVClock {
			t.Fatalf("unexpected error %q\n", err.Error())
		}
	}
}

func TestSetTickID(t *testing.T) {

	init := map[string]uint64{"a": 1, "b": 14}
	v, err := New(init)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v.Close()

	err = v.Tick("a")
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	init["a"] = 2

	m, err := v.GetMap()
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if len(m) != len(init) {
		t.Fatal("unexpected map")
	}

	if !reflect.DeepEqual(init, m) {
		t.Fatal("maps not equal")
	}
}

func TestTickBlank(t *testing.T) {

	init := map[string]uint64{"a": 1, "b": 14}
	v, err := New(init)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v.Close()

	err = v.Tick("")
	if err == nil {
		t.Fatal("unexpected pass when expected error")
	} else {
		if !errors.Is(err, errClockIdMustNotBeEmptyString) {
			t.Fatalf("unexpected error %q\n", err.Error())
		}
	}
}

func TestTickUnknown(t *testing.T) {

	init := map[string]uint64{"a": 1, "b": 14}
	v, err := New(init)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v.Close()

	err = v.Tick("c")
	if err == nil {
		t.Fatal("unexpected pass when expected error")
	} else {
		if !errors.Is(err, errAttemptToTickUnknownId) {
			t.Fatalf("unexpected error %q\n", err.Error())
		}
	}
}

func TestTickAfterClose(t *testing.T) {

	init := map[string]uint64{"a": 1, "b": 14}
	v, err := New(init)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	v.Close()

	err = v.Tick("a")
	if err == nil {
		t.Fatal("expected error but didn't get one")
	} else {
		if err != errClosedVClock {
			t.Fatalf("unexpected error %q\n", err.Error())
		}
	}
}

func TestSerialise(t *testing.T) {

	init := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	b, err := v1.Bytes()
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	v2, err := FromBytes(b)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	m, err := v2.GetMap()
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if len(m) != len(init) {
		t.Fatal("unexpected map")
	}

	if !reflect.DeepEqual(init, m) {
		t.Fatal("maps not equal")
	}
}

func TestSerialiseClosed(t *testing.T) {

	init := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	v1.Close()

	_, err = v1.Bytes()
	if err != nil {
		if err != errClosedVClock {
			t.Fatalf("unexpected error %q\n", err.Error())
		}
	} else {
		t.Fatal("expected error but got pass")
	}
}

func TestCopy(t *testing.T) {

	init := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	v2, err := v1.Copy()
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	m, err := v2.GetMap()
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if len(m) != len(init) {
		t.Fatal("unexpected map")
	}

	if !reflect.DeepEqual(init, m) {
		t.Fatal("maps not equal")
	}
}

func TestCopyClosed(t *testing.T) {

	init := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	v1.Close()

	_, err = v1.Copy()
	if err != nil {
		if err != errClosedVClock {
			t.Fatalf("unexpected error %q\n", err.Error())
		}
	} else {
		t.Fatal("expected error but did not get one")
	}
}

func TestGet(t *testing.T) {

	init := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	_, ok := v1.Get("")
	if ok {
		t.Fatal("unexpected error - returned true when should have returned false")
	}

	_, ok = v1.Get("c")
	if ok {
		t.Fatal("unexpected error - returned true when should have returned false")
	}

	val, ok := v1.Get("a")
	if !ok {
		t.Fatal("unexpected error - returned false when should have returned true")
	}
	if val != 1 {
		t.Fatalf("unexpected error - expected %q, got %q\n", 1, val)
	}
}

func TestGetAfterClose(t *testing.T) {

	init := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	v1.Close()

	_, ok := v1.Get("a")
	if ok {
		t.Fatal("unexpected error - returned true when should have returned false")
	}
}

func TestMergeWithNil(t *testing.T) {

	init := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	err = v1.Merge(nil)
	if err == nil {
		t.Fatal("unexpected error - expected an error to be raised")
	} else {
		if err != errClockMustNotBeNil {
			t.Fatalf("unexpected error - expected %q, got %q\n", errClockMustNotBeNil, err)
		}
	}
}

func TestMergeWithSelf(t *testing.T) {

	init := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	err = v1.Merge(v1)
	if err != nil {
		if err != errClockMustNotBeNil {
			t.Fatalf("unexpected error - expected %q, got %q\n", errClockMustNotBeNil, err)
		}
	}

	m, err := v1.GetMap()
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if len(m) != len(init) {
		t.Fatal("unexpected map")
	}

	if !reflect.DeepEqual(init, m) {
		t.Fatal("maps not equal")
	}
}

func TestMergeWithAnotherClock(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"a": 3, "c": 17}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result12 := map[string]uint64{"a": 3, "b": 14, "c": 17}

	v3, err := New(nil)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v3.Close()

	comp := func(vc1, vc2 *VClock, result map[string]uint64) {

		err = vc1.Merge(vc2)
		if err != nil {
			if err != errClockMustNotBeNil {
				t.Fatalf("unexpected error - expected %q, got %q\n", errClockMustNotBeNil, err)
			}
		}

		m, err := vc1.GetMap()
		if err != nil {
			t.Fatalf("unexpected error %q\n", err.Error())
		}

		if len(m) != len(result) {
			t.Fatalf("unexpected map: len(result)=%d, len(m)=%d", len(result), len(m))
		}

		if !reflect.DeepEqual(result, m) {
			t.Fatal("maps not equal")
		}
	}

	comp(v1, v2, result12)
	comp(v2, v1, result12)

	comp(v1, v3, result12)
	comp(v3, v1, result12)
}

func TestMergeWithSelfClosed(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	v1.Close()

	init2 := map[string]uint64{"a": 3, "c": 17}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	err = v1.Merge(v2)
	if err != nil {
		if err != errClosedVClock {
			t.Fatalf("unexpected error - expected %q, got %q\n", errClosedVClock, err)
		}
	}
}

func TestMergeWithOtherClosed(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"a": 3, "c": 17}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	v2.Close()

	err = v1.Merge(v2)
	if err != nil {
		if err != errClosedVClock {
			t.Fatalf("unexpected error - expected %q, got %q\n", errClosedVClock, err)
		}
	}
}

func TestLastUpdate(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	v2, err := New(nil)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	last := func(vc1 *VClock, expectedId string, expectedVal uint64) {

		id, val := vc1.LastUpdate()
		if id != expectedId {
			t.Fatalf("unexpected error - expected %q, got %q\n", expectedId, id)
		}
		if val != expectedVal {
			t.Fatalf("unexpected error - expected %d, got %d\n", expectedVal, val)
		}
	}

	last(v1, "b", 14)
	last(v2, "", 0)
}

func TestLastUpdateClosed(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	v1.Close()

	last := func(vc1 *VClock, expectedId string, expectedVal uint64) {

		id, val := vc1.LastUpdate()
		if id != expectedId {
			t.Fatalf("unexpected error - expected %q, got %q\n", expectedId, id)
		}
		if val != expectedVal {
			t.Fatalf("unexpected error - expected %d, got %d\n", expectedVal, val)
		}
	}

	last(v1, "", 0)
}

func TestEqualSame(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	v2, err := v1.Copy()
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.Equal(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if !result {
		t.Fatal("expected equality (true) but false returned")
	}
}

func TestEqualClockClosed(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	v2, err := v1.Copy()
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	v1.Close()

	_, err = v1.Equal(v2)
	if err == nil {
		t.Fatal("unexpected success when error expected")
	} else {
		if err != errClosedVClock {
			t.Fatalf("unexpected error %q\n", err.Error())
		}
	}
}

func TestEqualOtherClockClosed(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	v2, err := v1.Copy()
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	v2.Close()

	_, err = v1.Equal(v2)
	if err == nil {
		t.Fatal("unexpected success when error expected")
	} else {
		if err != errClosedVClock {
			t.Fatalf("unexpected error %q\n", err.Error())
		}
	}
}

func TestEqualOtherClockNil(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	_, err = v1.Equal(nil)
	if err == nil {
		t.Fatal("unexpected success when error expected")
	} else {
		if err != errClockMustNotBeNil {
			t.Fatalf("unexpected error %q\n", err.Error())
		}
	}
}

func TestEqualConcurrentClocks(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"c": 1, "d": 14}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.Equal(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if result {
		t.Fatal("unreleated vector clocks are compared as equal")
	}
}

func TestEqualClocksOverlapping(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"a": 1, "c": 14}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.Equal(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if result {
		t.Fatal("unreleated vector clocks are compared as equal")
	}
}

func TestEqualDescendent(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"a": 1, "b": 13}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.Equal(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if result {
		t.Fatal("unreleated vector clocks are compared as equal")
	}
}

func TestEqualAncestor1(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"a": 0}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.Equal(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if result {
		t.Fatal("unreleated vector clocks are compared as equal")
	}
}

func TestEqualAncestor2(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"a": 1, "b": 13}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.Equal(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if result {
		t.Fatal("unreleated vector clocks are compared as equal")
	}
}

func TestConcurrentSame(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	v2, err := v1.Copy()
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.Concurrent(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if result {
		t.Fatal("expected inequality (false) but true returned")
	}
}

func TestConcurrent1(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"c": 2, "d": 12}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.Concurrent(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if !result {
		t.Fatal("expected equality (true) but false returned")
	}
}

func TestConcurrent2(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"a": 1, "d": 12}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.Concurrent(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if !result {
		t.Fatal("expected equality (true) but false returned")
	}
}

func TestConcurrent3(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"a": 2, "d": 12}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.Concurrent(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if result {
		t.Fatal("expected inequality (false) but true returned")
	}
}

func TestConcurrent4(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"a": 1, "b": 13}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.Concurrent(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if result {
		t.Fatal("expected inequality (false) but true returned")
	}
}

func TestConcurrent5(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"a": 1, "b": 14, "c": 2}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.Concurrent(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if result {
		t.Fatal("expected inequality (false) but true returned")
	}
}

func TestConcurrent6(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"a": 1, "b": 14, "c": 2, "d": 1, "e": 54}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.Concurrent(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if result {
		t.Fatal("expected inequality (false) but true returned")
	}
}

func TestConcurrentClockClosed(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	v2, err := v1.Copy()
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	v1.Close()

	_, err = v1.Concurrent(v2)
	if err == nil {
		t.Fatal("unexpected success when error expected")
	} else {
		if err != errClosedVClock {
			t.Fatalf("unexpected error %q\n", err.Error())
		}
	}
}

func TestConcurrentOtherClockClosed(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	v2, err := v1.Copy()
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	v2.Close()

	_, err = v1.Concurrent(v2)
	if err == nil {
		t.Fatal("unexpected success when error expected")
	} else {
		if err != errClosedVClock {
			t.Fatalf("unexpected error %q\n", err.Error())
		}
	}
}

func TestConcurrentOtherClockNil(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	_, err = v1.Concurrent(nil)
	if err == nil {
		t.Fatal("unexpected success when error expected")
	} else {
		if err != errClockMustNotBeNil {
			t.Fatalf("unexpected error %q\n", err.Error())
		}
	}
}

func TestDescendsFromSame(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	v2, err := v1.Copy()
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.DescendsFrom(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if result {
		t.Fatal("expected inequality (false) but true returned")
	}
}

func TestDescendsFrom1(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"c": 8, "d": 11}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.DescendsFrom(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if result {
		t.Fatal("expected inequality (false) but true returned")
	}
}

func TestDescendsFrom2(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"a": 1, "d": 11}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.DescendsFrom(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if result {
		t.Fatal("expected inequality (false) but true returned")
	}
}

func TestDescendsFrom3(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"a": 2, "d": 11}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.DescendsFrom(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if result {
		t.Fatal("expected inequality (false) but true returned")
	}
}

func TestDescendsFrom4(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"a": 2}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.DescendsFrom(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if result {
		t.Fatal("expected inequality (false) but true returned")
	}
}

func TestDescendsFrom5(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"a": 2, "b": 13}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.DescendsFrom(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if result {
		t.Fatal("expected inequality (false) but true returned")
	}
}

func TestDescendsFrom6(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"a": 2, "b": 14}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.DescendsFrom(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if !result {
		t.Fatal("expected equality (true) but false returned")
	}
}

func TestDescendsFrom7(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"a": 2, "b": 14, "c": 2}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.DescendsFrom(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if !result {
		t.Fatal("expected equality (true) but false returned")
	}
}

func TestDescendsFrom8(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"a": 2, "b": 15}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.DescendsFrom(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if !result {
		t.Fatal("expected equality (true) but false returned")
	}
}

func TestDescendsFrom9(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"a": 1, "b": 15}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.DescendsFrom(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if !result {
		t.Fatal("expected equality (true) but false returned")
	}
}

func TestDescendsFrom10(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"a": 1, "b": 15, "c": 3, "d": 7}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.DescendsFrom(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if !result {
		t.Fatal("expected equality (true) but false returned")
	}
}

func TestDescendsFrom11(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"a": 0, "b": 14}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.DescendsFrom(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if result {
		t.Fatal("expected inequality (false) but true returned")
	}
}

func TestDescendsFrom12(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"a": 0, "b": 13}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.DescendsFrom(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if result {
		t.Fatal("expected inequality (false) but true returned")
	}
}

func TestDescendsFrom13(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"a": 0, "c": 13}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.DescendsFrom(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if result {
		t.Fatal("expected inequality (false) but true returned")
	}
}

func TestDescendsFrom14(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"a": 0, "c": 13, "d": 17}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.DescendsFrom(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if result {
		t.Fatal("expected inequality (false) but true returned")
	}
}

func TestDescendsFrom15(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"a": 2, "c": 13, "d": 17}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.DescendsFrom(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if result {
		t.Fatal("expected inequality (false) but true returned")
	}
}

func TestDescendsFromClockClosed(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	v2, err := v1.Copy()
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	v1.Close()

	_, err = v1.DescendsFrom(v2)
	if err == nil {
		t.Fatal("unexpected success when error expected")
	} else {
		if err != errClosedVClock {
			t.Fatalf("unexpected error %q\n", err.Error())
		}
	}
}

func TestDescendsFromOtherClockClosed(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	v2, err := v1.Copy()
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	v2.Close()

	_, err = v1.DescendsFrom(v2)
	if err == nil {
		t.Fatal("unexpected success when error expected")
	} else {
		if err != errClosedVClock {
			t.Fatalf("unexpected error %q\n", err.Error())
		}
	}
}

func TestDescendsFromOtherClockNil(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	_, err = v1.DescendsFrom(nil)
	if err == nil {
		t.Fatal("unexpected success when error expected")
	} else {
		if err != errClockMustNotBeNil {
			t.Fatalf("unexpected error %q\n", err.Error())
		}
	}
}

func TestAncestorOfSame(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	v2, err := v1.Copy()
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.AncestorOf(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if result {
		t.Fatal("expected inequality (false) but true returned")
	}
}

func TestAncestorOfClockClosed(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	v2, err := v1.Copy()
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	v1.Close()

	_, err = v1.AncestorOf(v2)
	if err == nil {
		t.Fatal("unexpected success when error expected")
	} else {
		if err != errClosedVClock {
			t.Fatalf("unexpected error %q\n", err.Error())
		}
	}
}

func TestAncestorOfOtherClockClosed(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	v2, err := v1.Copy()
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	v2.Close()

	_, err = v1.AncestorOf(v2)
	if err == nil {
		t.Fatal("unexpected success when error expected")
	} else {
		if err != errClosedVClock {
			t.Fatalf("unexpected error %q\n", err.Error())
		}
	}
}

func TestAncestorOfOtherClockNil(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	_, err = v1.AncestorOf(nil)
	if err == nil {
		t.Fatal("unexpected success when error expected")
	} else {
		if err != errClockMustNotBeNil {
			t.Fatalf("unexpected error %q\n", err.Error())
		}
	}
}

func TestAncestorOf1(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 14}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"a": 0}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.AncestorOf(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if !result {
		t.Fatal("expected equality (true) but false returned")
	}
}

func TestAncestorOf2(t *testing.T) {

	init1 := map[string]uint64{"a": 1}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"a": 0}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.AncestorOf(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if !result {
		t.Fatal("expected equality (true) but false returned")
	}
}

func TestAncestorOf3(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 23, "c": 8}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"a": 0}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.AncestorOf(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if !result {
		t.Fatal("expected equality (true) but false returned")
	}
}

func TestAncestorOf4(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 23, "c": 8}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"a": 0, "b": 23}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.AncestorOf(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if !result {
		t.Fatal("expected equality (true) but false returned")
	}
}

func TestAncestorOf5(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 23, "c": 8}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"a": 0, "b": 23, "c": 8}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.AncestorOf(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if !result {
		t.Fatal("expected equality (true) but false returned")
	}
}

func TestAncestorOf6(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 23, "c": 8}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"a": 0, "b": 23, "c": 7}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.AncestorOf(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if !result {
		t.Fatal("expected equality (true) but false returned")
	}
}

func TestAncestorOf7(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 23, "c": 8}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"a": 0, "b": 24, "c": 7}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.AncestorOf(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if result {
		t.Fatal("expected inequality (false) but true returned")
	}
}

func TestAncestorOf8(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 23, "c": 8}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"a": 1, "b": 24, "c": 8}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.AncestorOf(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if result {
		t.Fatal("expected inequality (false) but true returned")
	}
}

func TestAncestorOf9(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 23, "c": 8}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"a": 1}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.AncestorOf(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if !result {
		t.Fatal("expected equality (true) but false returned")
	}
}

func TestAncestorOf10(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 23, "c": 8}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"a": 0, "d": 3}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.AncestorOf(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if result {
		t.Fatal("expected inequality (false) but true returned")
	}
}

func TestAncestorOf11(t *testing.T) {

	init1 := map[string]uint64{"a": 1, "b": 23, "c": 8}
	v1, err := New(init1)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v1.Close()

	init2 := map[string]uint64{"d": 3}
	v2, err := New(init2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}
	defer v2.Close()

	result, err := v1.AncestorOf(v2)
	if err != nil {
		t.Fatalf("unexpected error %q\n", err.Error())
	}

	if result {
		t.Fatal("expected inequality (false) but true returned")
	}
}
