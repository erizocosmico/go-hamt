# go-hamt

Hash array mapped trie implementation for Go. This is a persistent hash map, so whenever an element is added, removed, or replaced, a new hash map is returned instead of modifying the previous map. This makes the hash map immutable.

The implementation is based on the paper [Ideal Hash Trees](http://lampwww.epfl.ch/papers/idealhashtrees.pdf).

**IMPORTANT:** it's not, and it's not meant to be, thread safe.

## Usage

```go
m := hamt.New()

m.Store("key", "value")

val, ok := m.Load("key")
if ok {
    // use val
}
```

## Status

It's in a very early stage and performance is still pretty bad for real life usage.

## Benchmarks

```
BenchmarkStore10-4           	 2000000	       834 ns/op	     632 B/op	       8 allocs/op
BenchmarkStore100-4          	 1000000	      1014 ns/op	     696 B/op	      10 allocs/op
BenchmarkStore1000-4         	 1000000	      1178 ns/op	     920 B/op	      10 allocs/op
BenchmarkStoreStruct10-4     	 2000000	       915 ns/op	     664 B/op	       9 allocs/op
BenchmarkStoreStruct100-4    	 1000000	      1077 ns/op	     728 B/op	      11 allocs/op
BenchmarkStoreStruct1000-4   	 1000000	      1298 ns/op	     952 B/op	      11 allocs/op
BenchmarkLookup10-4          	10000000	       245 ns/op	      32 B/op	       4 allocs/op
BenchmarkLookup100-4         	 5000000	       253 ns/op	      32 B/op	       4 allocs/op
BenchmarkLookup1000-4        	 5000000	       250 ns/op	      32 B/op	       4 allocs/op
```

These are the benchmarks of how much it takes to perform a Store or Load operation depending on the size of the map we're applying the operation to.

## Pending

- Remove
- Range
- Hopefully performance improvements