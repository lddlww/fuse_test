// Copyright (c) 2014 The go-patricia AUTHORS
//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package patricia

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"testing"
)

// Tests -----------------------------------------------------------------------

// HeapOverhead is allowed tolerance for Go's runtime/GC to increase the allocated memory
// (to avoid failing tests on insignificant growth amounts)
//
// Can be overwritten by setting PATRICIA_TESTS_HEAP_OVERHEAD env variable.
var HeapOverhead uint64 = 20000

func init() {
	if v := os.Getenv("PATRICIA_TESTS_HEAP_OVERHEAD"); v != "" {
		i, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			panic(fmt.Errorf("failed to parse PATRICIA_TESTS_HEAP_OVERHEAD: %v", err))
		}
		HeapOverhead = i
	}
}

func TestTrie_InsertDense(t *testing.T) {
	trie := NewTrie()

	data := []testData{
		{"aba", 0, success},
		{"abb", 1, success},
		{"abc", 2, success},
		{"abd", 3, success},
		{"abe", 4, success},
		{"abf", 5, success},
		{"abg", 6, success},
		{"abh", 7, success},
		{"abi", 8, success},
		{"abj", 9, success},
		{"abk", 0, success},
		{"abl", 1, success},
		{"abm", 2, success},
		{"abn", 3, success},
		{"abo", 4, success},
		{"abp", 5, success},
		{"abq", 6, success},
		{"abr", 7, success},
		{"abs", 8, success},
		{"abt", 9, success},
		{"abu", 0, success},
		{"abv", 1, success},
		{"abw", 2, success},
		{"abx", 3, success},
		{"aby", 4, success},
		{"abz", 5, success},
	}

	for _, v := range data {
		t.Logf("INSERT prefix=%v, item=%v, success=%v", v.key, v.value, v.retVal)
		if ok := trie.Insert(Prefix(v.key), v.value); ok != v.retVal {
			t.Errorf("Unexpected return value, expected=%v, got=%v", v.retVal, ok)
		}
	}
}

func TestTrie_InsertDensePreceeding(t *testing.T) {
	trie := NewTrie()
	start := byte(70)
	// create a dense node
	for i := byte(0); i <= DefaultMaxChildrenPerSparseNode; i++ {
		if !trie.Insert(Prefix([]byte{start + i}), true) {
			t.Errorf("insert failed, prefix=%v", start+i)
		}
	}
	// insert some preceding keys
	for i := byte(1); i < start; i *= i + 1 {
		if !trie.Insert(Prefix([]byte{start - i}), true) {
			t.Errorf("insert failed, prefix=%v", start-i)
		}
	}
}

func TestTrie_InsertDenseDuplicatePrefixes(t *testing.T) {
	trie := NewTrie()

	data := []testData{
		{"aba", 0, success},
		{"abb", 1, success},
		{"abc", 2, success},
		{"abd", 3, success},
		{"abe", 4, success},
		{"abf", 5, success},
		{"abg", 6, success},
		{"abh", 7, success},
		{"abi", 8, success},
		{"abj", 9, success},
		{"abk", 0, success},
		{"abl", 1, success},
		{"abm", 2, success},
		{"abn", 3, success},
		{"abo", 4, success},
		{"abp", 5, success},
		{"abq", 6, success},
		{"abr", 7, success},
		{"abs", 8, success},
		{"abt", 9, success},
		{"abu", 0, success},
		{"abv", 1, success},
		{"abw", 2, success},
		{"abx", 3, success},
		{"aby", 4, success},
		{"abz", 5, success},
		{"aba", 0, failure},
		{"abb", 1, failure},
		{"abc", 2, failure},
		{"abd", 3, failure},
		{"abe", 4, failure},
	}

	for _, v := range data {
		t.Logf("INSERT prefix=%v, item=%v, success=%v", v.key, v.value, v.retVal)
		if ok := trie.Insert(Prefix(v.key), v.value); ok != v.retVal {
			t.Errorf("Unexpected return value, expected=%v, got=%v", v.retVal, ok)
		}
	}
}

func TestTrie_CloneDense(t *testing.T) {
	trie := NewTrie()

	data := []testData{
		{"aba", 0, success},
		{"abb", 1, success},
		{"abc", 2, success},
		{"abd", 3, success},
		{"abe", 4, success},
		{"abf", 5, success},
		{"abg", 6, success},
		{"abh", 7, success},
		{"abi", 8, success},
		{"abj", 9, success},
		{"abk", 0, success},
		{"abl", 1, success},
		{"abm", 2, success},
		{"abn", 3, success},
		{"abo", 4, success},
		{"abp", 5, success},
		{"abq", 6, success},
		{"abr", 7, success},
		{"abs", 8, success},
		{"abt", 9, success},
		{"abu", 0, success},
		{"abv", 1, success},
		{"abw", 2, success},
		{"abx", 3, success},
		{"aby", 4, success},
		{"abz", 5, success},
	}

	for _, v := range data {
		t.Logf("INSERT prefix=%v, item=%v, success=%v", v.key, v.value, v.retVal)
		if ok := trie.Insert(Prefix(v.key), v.value); ok != v.retVal {
			t.Errorf("Unexpected return value, expected=%v, got=%v", v.retVal, ok)
		}
	}

	t.Log("CLONE")
	clone := trie.Clone()

	for _, v := range data {
		t.Logf("GET prefix=%v, item=%v", v.key, v.value)
		if item := clone.Get(Prefix(v.key)); item != v.value {
			t.Errorf("Unexpected return value, expected=%v, got=%v", v.value, item)
		}
	}

	prefix := "xxx"
	item := 666
	t.Logf("INSERT prefix=%v, item=%v", prefix, item)
	if ok := trie.Insert(Prefix(prefix), item); !ok {
		t.Errorf("Unexpected return value, expected=true, got=%v", ok)
	}
	t.Logf("GET cloned prefix=%v", prefix)
	if item := clone.Get(Prefix(prefix)); item != nil {
		t.Errorf("Unexpected return value, expected=nil, got=%v", item)
	}
}

func TestTrie_DeleteDense(t *testing.T) {
	trie := NewTrie()

	data := []testData{
		{"aba", 0, success},
		{"abb", 1, success},
		{"abc", 2, success},
		{"abd", 3, success},
		{"abe", 4, success},
		{"abf", 5, success},
		{"abg", 6, success},
		{"abh", 7, success},
		{"abi", 8, success},
		{"abj", 9, success},
		{"abk", 0, success},
		{"abl", 1, success},
		{"abm", 2, success},
		{"abn", 3, success},
		{"abo", 4, success},
		{"abp", 5, success},
		{"abq", 6, success},
		{"abr", 7, success},
		{"abs", 8, success},
		{"abt", 9, success},
		{"abu", 0, success},
		{"abv", 1, success},
		{"abw", 2, success},
		{"abx", 3, success},
		{"aby", 4, success},
		{"abz", 5, success},
	}

	for _, v := range data {
		t.Logf("INSERT prefix=%v, item=%v, success=%v", v.key, v.value, v.retVal)
		if ok := trie.Insert(Prefix(v.key), v.value); ok != v.retVal {
			t.Errorf("Unexpected return value, expected=%v, got=%v", v.retVal, ok)
		}
	}

	for _, v := range data {
		t.Logf("DELETE word=%v, success=%v", v.key, v.retVal)
		if ok := trie.Delete([]byte(v.key)); ok != v.retVal {
			t.Errorf("Unexpected return value, expected=%v, got=%v", v.retVal, ok)
		}
	}
}

func TestTrie_DeleteLeakageDense(t *testing.T) {
	trie := NewTrie()

	genTestData := func() *testData {
		// Generate a random integer as a key.
		key := strconv.FormatUint(rand.Uint64(), 10)
		return &testData{key: key, value: "v", retVal: success}
	}

	testSize := 100
	data := make([]*testData, 0, testSize)
	for i := 0; i < testSize; i++ {
		data = append(data, genTestData())
	}

	oldBytes := heapAllocatedBytes()

	// repeat insertion/deletion for 10K times to catch possible memory issues
	for i := 0; i < 10000; i++ {
		for _, v := range data {
			if ok := trie.Insert(Prefix(v.key), v.value); ok != v.retVal {
				t.Errorf("Unexpected return value, expected=%v, got=%v", v.retVal, ok)
			}
		}

		for _, v := range data {
			if ok := trie.Delete([]byte(v.key)); ok != v.retVal {
				t.Errorf("Unexpected return value, expected=%v, got=%v", v.retVal, ok)
			}
		}
	}

	if newBytes := heapAllocatedBytes(); newBytes > oldBytes+HeapOverhead {
		t.Logf("Size=%d, Total=%d, Trie state:\n%s\n", trie.size(), trie.total(), trie.dump())
		t.Errorf("Heap space leak, grew %d bytes (%d to %d)\n", newBytes-oldBytes, oldBytes, newBytes)
	}

	if numChildren := trie.children.length(); numChildren != 0 {
		t.Errorf("Trie is not empty: %v children found", numChildren)
	}
}

func heapAllocatedBytes() uint64 {
	runtime.GC()

	ms := runtime.MemStats{}
	runtime.ReadMemStats(&ms)
	return ms.Alloc
}
