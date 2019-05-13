package hamt

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"math"
	"reflect"
)

func hash(x interface{}) uint32 {
	return rehash(x, 0)
}

func rehash(x interface{}, level uint32) uint32 {
	var h = crc32.NewIEEE()
	hashUint32(h, level)
	hashValue(h, x)
	return h.Sum32()
}

func hashValue(w io.Writer, x interface{}) {
	hashType(w, x)

	switch x := x.(type) {
	case byte:
		writebyte(w, x)
	case int8:
		ux := byte(x) << 1
		if x < 0 {
			ux = ^ux
		}
		writebyte(w, ux)
	case uint16:
		var bs = make([]byte, 2)
		binary.LittleEndian.PutUint16(bs, x)
		assertwrite(w.Write(bs))
	case int16:
		ux := uint16(x) << 1
		if x < 0 {
			ux = ^ux
		}
		bs := make([]byte, 2)
		binary.LittleEndian.PutUint16(bs, ux)
		assertwrite(w.Write(bs))
	case uint32:
		hashUint32(w, x)
	case int32:
		ux := uint32(x) << 1
		if x < 0 {
			ux = ^ux
		}
		bs := make([]byte, 4)
		binary.LittleEndian.PutUint32(bs, ux)
		assertwrite(w.Write(bs))
	case uint64:
		var bs = make([]byte, 8)
		binary.LittleEndian.PutUint64(bs, x)
		assertwrite(w.Write(bs))
	case int64:
		ux := uint64(x) << 1
		if x < 0 {
			ux = ^ux
		}
		bs := make([]byte, 8)
		binary.LittleEndian.PutUint64(bs, ux)
		assertwrite(w.Write(bs))
	case uint:
		var bs = make([]byte, 8)
		binary.LittleEndian.PutUint64(bs, uint64(x))
		assertwrite(w.Write(bs))
	case int:
		ux := uint64(x) << 1
		if x < 0 {
			ux = ^ux
		}
		bs := make([]byte, 8)
		binary.LittleEndian.PutUint64(bs, ux)
		assertwrite(w.Write(bs))
	case float32:
		hashUint32(w, math.Float32bits(x))
	case float64:
		var bs = make([]byte, 8)
		binary.LittleEndian.PutUint64(bs, math.Float64bits(x))
		assertwrite(w.Write(bs))
	case string:
		assertwrite(w.Write([]byte(x)))
	case []byte:
		assertwrite(w.Write(x))
	case bool:
		if x {
			writebyte(w, 1)
		} else {
			writebyte(w, 0)
		}
	case uintptr:
		hashPtr(w, x)
	case nil:
		// don't write anything
	default:
		v := reflect.ValueOf(x)
		switch v.Kind() {
		case reflect.Map, reflect.Interface, reflect.Ptr:
			hashPtr(w, v.Pointer())
		case reflect.Slice, reflect.Array:
			for i := 0; i < v.Len(); i++ {
				hashValue(w, v.Index(i).Interface())
				assertwrite(w.Write([]byte{0}))
			}
		case reflect.Struct:
			t := reflect.TypeOf(x)
			assertwrite(w.Write([]byte(t.Name())))
			for i := 0; i < t.NumField(); i++ {
				hashValue(w, v.Field(i).Interface())
				assertwrite(w.Write([]byte{0}))
			}
		default:
			panic(fmt.Sprintf("hamt: cannot index value of type %T", x))
		}
	}
}

func hashType(w io.Writer, x interface{}) {
	switch x := x.(type) {
	case byte:
		writebyte(w, 1)
	case int8:
		writebyte(w, 2)
	case uint16:
		writebyte(w, 3)
	case int16:
		writebyte(w, 4)
	case uint32:
		writebyte(w, 5)
	case int32:
		writebyte(w, 6)
	case uint64:
		writebyte(w, 7)
	case int64:
		writebyte(w, 8)
	case uint:
		writebyte(w, 9)
	case int:
		writebyte(w, 10)
	case float32:
		writebyte(w, 11)
	case float64:
		writebyte(w, 12)
	case string:
		writebyte(w, 13)
	case []byte:
		writebyte(w, 14)
	case bool:
		writebyte(w, 15)
	case uintptr:
		writebyte(w, 16)
	case nil:
		writebyte(w, 17)
		// don't write anything
	default:
		v := reflect.ValueOf(x)
		switch v.Kind() {
		case reflect.Map:
			writebyte(w, 18)
		case reflect.Interface:
			writebyte(w, 19)
		case reflect.Ptr:
			writebyte(w, 20)
		case reflect.Slice:
			writebyte(w, 21)
		case reflect.Array:
			writebyte(w, 22)
		case reflect.Struct:
			writebyte(w, 23)
		}
	}
}

func writebyte(w io.Writer, b byte) {
	assertwrite(w.Write([]byte{b}))
}

func hashUint32(w io.Writer, x uint32) {
	var bs = make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, x)
	assertwrite(w.Write(bs))
}

func hashPtr(w io.Writer, ptr uintptr) {
	var bs = make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, uint64(ptr))
	assertwrite(w.Write(bs))
}

func assertwrite(_ int, err error) {
	if err != nil {
		panic(err)
	}
}
