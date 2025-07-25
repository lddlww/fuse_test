/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package typeurl

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type test struct {
	Name string
	Age  int
}

func clear() {
	registry = make(map[reflect.Type]string)
}

var _ Any = &anypb.Any{}

func TestRegisterPointerGetPointer(t *testing.T) {
	clear()
	expected := "test"
	Register(&test{}, "test")

	url, err := TypeURL(&test{})
	if err != nil {
		t.Fatal(err)
	}
	if url != expected {
		t.Fatalf("expected %q but received %q", expected, url)
	}
}

func TestMarshal(t *testing.T) {
	clear()
	expected := "test"
	Register(&test{}, "test")

	v := &test{
		Name: "koye",
		Age:  6,
	}
	any, err := MarshalAny(v)
	if err != nil {
		t.Fatal(err)
	}
	if any.GetTypeUrl() != expected {
		t.Fatalf("expected %q but received %q", expected, any.GetTypeUrl())
	}

	// marshal it again and make sure we get the same thing back.
	newany, err := MarshalAny(any)
	if err != nil {
		t.Fatal(err)
	}

	val := any.GetValue()
	newval := newany.GetValue()

	// Ensure pointer to same exact slice
	newval[0] = val[0] ^ 0xff

	if !bytes.Equal(newval, val) {
		t.Fatalf("expected to get back same object: %v != %v", newany, any)
	}

}

func TestMarshalUnmarshal(t *testing.T) {
	clear()
	Register(&test{}, "test")

	v := &test{
		Name: "koye",
		Age:  6,
	}
	any, err := MarshalAny(v)
	if err != nil {
		t.Fatal(err)
	}
	nv, err := UnmarshalAny(any)
	if err != nil {
		t.Fatal(err)
	}
	td, ok := nv.(*test)
	if !ok {
		t.Fatal("expected value to cast to *test")
	}
	if td.Name != "koye" {
		t.Fatal("invalid name")
	}
	if td.Age != 6 {
		t.Fatal("invalid age")
	}
}

func TestMarshalUnmarshalTo(t *testing.T) {
	clear()
	Register(&test{}, "test")

	in := &test{
		Name: "koye",
		Age:  6,
	}
	any, err := MarshalAny(in)
	if err != nil {
		t.Fatal(err)
	}
	out := &test{}
	err = UnmarshalTo(any, out)
	if err != nil {
		t.Fatal(err)
	}
	if out.Name != "koye" {
		t.Fatal("invalid name")
	}
	if out.Age != 6 {
		t.Fatal("invalid age")
	}
}

type test2 struct {
	Name string
}

func TestUnmarshalToInvalid(t *testing.T) {
	clear()
	Register(&test{}, "test1")
	Register(&test2{}, "test2")

	in := &test{
		Name: "koye",
		Age:  6,
	}
	any, err := MarshalAny(in)
	if err != nil {
		t.Fatal(err)
	}

	out := &test2{}
	err = UnmarshalTo(any, out)
	if err == nil || err.Error() != `can't unmarshal type "test1" to output "test2"` {
		t.Fatalf("unexpected result: %+v", err)
	}
}

func TestFromAny(t *testing.T) {
	actual := MarshalProto(nil)
	if actual != nil {
		t.Fatalf("expected nil, got %v", actual)
	}
}

func TestIs(t *testing.T) {
	clear()
	Register(&test{}, "test")

	v := &test{
		Name: "koye",
		Age:  6,
	}
	any, err := MarshalAny(v)
	if err != nil {
		t.Fatal(err)
	}
	if !Is(any, &test{}) {
		t.Fatal("Is(any, test{}) should be true")
	}

	// shouldn't crash
	Is(nil, &test{})
}

func TestRegisterDiffUrls(t *testing.T) {
	clear()
	defer func() {
		if err := recover(); err == nil {
			t.Error("registering the same type with different urls should panic")
		}
	}()
	Register(&test{}, "test")
	Register(&test{}, "test", "two")
}

func TestCheckNil(t *testing.T) {
	var a *anyType

	actual := a.GetValue()
	if actual != nil {
		t.Fatalf("expected nil, got %v", actual)
	}
}

func TestProtoFallback(t *testing.T) {
	expected := time.Now()
	b, err := proto.Marshal(timestamppb.New(expected))
	if err != nil {
		t.Fatal(err)
	}
	x, err := UnmarshalByTypeURL("type.googleapis.com/google.protobuf.Timestamp", b)
	if err != nil {
		t.Fatal(err)
	}
	ts, ok := x.(*timestamppb.Timestamp)
	if !ok {
		t.Fatalf("failed to convert %+v to Timestamp", x)
	}
	if expected.Sub(ts.AsTime()) != 0 {
		t.Fatalf("expected %+v but got %+v", expected, ts.AsTime())
	}
}

func TestUnmarshalNotFound(t *testing.T) {
	_, err := UnmarshalByTypeURL("doesntexist", []byte("{}"))
	if err == nil {
		t.Fatalf("expected error unmarshalling type which does not exist")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("unexpected error unmarshalling type which does not exist: %v", err)
	}
}

func TestUnmarshalJSON(t *testing.T) {
	url := t.Name()
	Register(&timestamppb.Timestamp{}, url)

	expected := timestamppb.Now()

	dt, err := json.Marshal(expected)
	if err != nil {
		t.Fatal(err)
	}

	var actual timestamppb.Timestamp
	if err := UnmarshalToByTypeURL(url, dt, &actual); err != nil {
		t.Fatal(err)
	}

	if !expected.AsTime().Equal(actual.AsTime()) {
		t.Fatalf("expected value to be %q, got: %q", expected.AsTime(), actual.AsTime())
	}
}
