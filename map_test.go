package hamt

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	m := New()
	assertCount(t, m, 0)
	assertAllNodesNil(t, m)
}

func TestMapAssocAndGet(t *testing.T) {
	m := New()
	m2 := m.Store("hello", "world")

	assertCount(t, m2, 1)
	assertCount(t, m, 0)
	assertAllNodesNil(t, m)
	assertSomeNodesFull(t, m2)
	res, ok := m2.Load("hello")
	require.True(t, ok)
	require.NotNil(t, res)
	require.Equal(t, res, "world")

	m3 := m2.Store("bar", "baz")

	assertCount(t, m3, 2)
	assertCount(t, m2, 1)
	assertSomeNodesFull(t, m3)
	assertSomeNodesFull(t, m2)
	_, ok = m2.Load("bar")
	require.False(t, ok)

	res, ok = m3.Load("bar")
	require.True(t, ok)
	require.NotNil(t, res)
	require.Equal(t, res, "baz")

	m4 := m3.Store("bar", "foo")

	assertCount(t, m3, 2)
	assertCount(t, m4, 2)
	assertSomeNodesFull(t, m3)
	assertSomeNodesFull(t, m4)
	res, ok = m3.Load("bar")
	require.True(t, ok)
	require.NotNil(t, res)
	require.Equal(t, res, "baz")

	res, ok = m4.Load("bar")
	require.True(t, ok)
	require.NotNil(t, res)
	require.Equal(t, res, "foo")
}

func BenchmarkStore10(b *testing.B) {
	for n := 0; n < b.N; n++ {
		map10.Store(11, "hello world")
	}
}

func BenchmarkStore100(b *testing.B) {
	for n := 0; n < b.N; n++ {
		map100.Store(101, "hello world")
	}
}

func BenchmarkStore1000(b *testing.B) {
	for n := 0; n < b.N; n++ {
		map1000.Store(1001, "hello world")
	}
}

func storeN(n int, value interface{}) *Map {
	m := New()
	for i := 0; i < n; i++ {
		m = m.Store(i, value)
	}
	return m
}

type TestStruct struct {
	N int
	S string
}

func BenchmarkStoreStruct10(b *testing.B) {
	for n := 0; n < b.N; n++ {
		map10.Store(11, TestStruct{n, "hello world"})
	}
}

func BenchmarkStoreStruct100(b *testing.B) {
	for n := 0; n < b.N; n++ {
		map100.Store(101, TestStruct{n, "hello world"})
	}
}

func BenchmarkStoreStruct1000(b *testing.B) {
	for n := 0; n < b.N; n++ {
		map1000.Store(1001, TestStruct{n, "hello world"})
	}
}

const N = 10000

var lookupmap = func() *Map {
	m := New()
	for n := 0; n < N; n++ {
		m = m.Store(n, fmt.Sprintf("hello world_%d", n))
	}
	return m
}()

func TestLookup10K(t *testing.T) {
	assertCount(t, lookupmap, N)
	for n := 0; n < N; n++ {
		v, ok := lookupmap.Load(n)
		require.True(t, ok)
		require.Equal(t, v, fmt.Sprintf("hello world_%d", n))
	}
}

var (
	map10   = storeN(10, "hello world")
	map100  = storeN(100, "hello world")
	map1000 = storeN(1000, "hello world")
)

func BenchmarkLookup10(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, _ = map10.Load(8)
	}
}

func BenchmarkLookup100(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, _ = map100.Load(87)
	}
}

func BenchmarkLookup1000(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, _ = map1000.Load(950)
	}
}

func assertCount(t *testing.T, m *Map, n int) {
	require.Equal(t, uint64(n), m.rootHashTable.count)
}

func assertAllNodesNil(t *testing.T, m *Map) {
	for _, n := range m.rootHashTable.nodes {
		require.Nil(t, n)
	}
}

func assertSomeNodesFull(t *testing.T, m *Map) {
	var nilNodes uint64
	for _, n := range m.rootHashTable.nodes {
		if isNil(n) {
			nilNodes++
		}
	}

	if m.rootHashTable.count == nilNodes {
		t.FailNow()
	}
}
