package hamt

// Map implements a persistent hash array mapped trie.
type Map struct {
	*rootHashTable
}

// New creates a new map.
func New() *Map {
	return &Map{emptyRootHashTable()}
}

// Load retrieves the value from the map with the given key and whether it
// found a value or not.
func (m *Map) Load(key interface{}) (v interface{}, ok bool) {
	return m.lookup(hash(key), key)
}

// Store a key and a value. If the key is found then the value is not inserted
// but replaced.
func (m *Map) Store(key, value interface{}) *Map {
	return &Map{m.assoc(hash(key), key, value, true)}
}

type rootHashTable struct {
	count uint64
	nodes []interface{}
}

func emptyRootHashTable() *rootHashTable {
	return &rootHashTable{0, make([]interface{}, 32)}
}

func (t *rootHashTable) clone() *rootHashTable {
	nt := &rootHashTable{t.count, make([]interface{}, 32)}
	copy(nt.nodes, t.nodes)
	return nt
}

func (t *rootHashTable) assoc(hash uint32, key, value interface{}, persistent bool) *rootHashTable {
	idx := mask(hash, 0)

	newRoot := t
	if persistent {
		newRoot = t.clone()
	}
	newRoot.count++
	switch n := t.nodes[idx]; n.(type) {
	case *keyValuePair:
		kv := n.(*keyValuePair)
		if kv.keyEquals(key) {
			newRoot.nodes[idx] = &keyValuePair{hash, key, value}
			newRoot.count--
			return newRoot
		}

		// there is a hash collision
		sht := resolveHashCollision(5, kv, &keyValuePair{hash, key, value})
		newRoot.nodes[idx] = sht
	case *subHashTable:
		sht := n.(*subHashTable)
		sht, replaced := sht.assoc(5, hash, key, value, persistent)

		newRoot.nodes[idx] = sht
		if replaced {
			newRoot.count--
		}
	case nil:
		newRoot.nodes[idx] = &keyValuePair{hash, key, value}
	}

	return newRoot
}

func (t *rootHashTable) lookup(hash uint32, key interface{}) (interface{}, bool) {
	idx := mask(hash, 0)
	switch n := t.nodes[idx]; n.(type) {
	case *keyValuePair:
		if kv := n.(*keyValuePair); kv.keyEquals(key) {
			return kv.value, true
		}
	case *subHashTable:
		return n.(*subHashTable).lookup(5, hash, key)
	}
	return nil, false
}

type subHashTable struct {
	bitmap uint32
	base   []interface{}
}

func newSubHashTable() *subHashTable {
	return &subHashTable{0, nil}
}

func (t *subHashTable) clone() *subHashTable {
	return &subHashTable{t.bitmap, t.base}
}

func (t *subHashTable) pos(idx uint32) uint32 {
	return uint32(len(t.base)+1) - popcnt(t.bitmap>>idx)
}

func (t *subHashTable) lookupPos(idx uint32) uint32 {
	return uint32(len(t.base)) - popcnt(t.bitmap>>idx)
}

func (t *subHashTable) setBitmap(idx uint32) {
	t.bitmap = t.bitmap | 1<<idx
}

// setAtCopying sets the node at the given position performing copies of the previous elements.
func (t *subHashTable) setAtCopying(idx uint32, node interface{}) {
	t.setBitmap(idx)
	startIdx := t.pos(idx)
	base := make([]interface{}, len(t.base)+1)
	if startIdx > 0 {
		copy(base[:startIdx], t.base[:startIdx])
	}
	base[startIdx] = node
	if uint32(len(t.base)) > startIdx {
		copy(base[startIdx+1:], t.base[startIdx:])
	}
	t.base = base
}

func (t *subHashTable) replace(pos uint32, node interface{}) {
	base := make([]interface{}, len(t.base))
	copy(base, t.base)
	base[pos] = node
	t.base = base
}

func (t *subHashTable) assoc(shift, hash uint32, key, value interface{}, persistent bool) (*subHashTable, bool) {
	// bits of the hash have been exhausted, we need a rehash
	if shift >= 32 {
		hash = rehash(key, shift/5)
		shift = 0
	}

	idx := mask(hash, shift)
	bit := (t.bitmap >> idx) & 0x01
	if bit == 0 {
		if t.bitmap == 0 || !persistent {
			t.setBitmap(idx)
			t.base = []interface{}{&keyValuePair{hash, key, value}}
			return t, false
		}

		sht := t.clone()
		sht.setAtCopying(idx, &keyValuePair{hash, key, value})
		return sht, false
	}

	pos := t.lookupPos(idx)
	switch n := t.base[pos]; n.(type) {
	case *keyValuePair:
		return t.assocKeyValuePair(n.(*keyValuePair), idx, pos, shift, hash, key, value, persistent)
	case *subHashTable:
		return t.assocSubHashTable(n.(*subHashTable), idx, pos, shift, hash, key, value, persistent)
	}

	return t, false
}

func (t *subHashTable) assocSubHashTable(
	sht *subHashTable,
	idx, pos, shift, hash uint32,
	key, value interface{},
	persistent bool,
) (*subHashTable, bool) {
	sht, replaced := sht.assoc(shift+5, hash, key, value, persistent)
	if persistent {
		parent := t.clone()
		parent.replace(pos, sht)
		return parent, replaced
	}
	return t, replaced
}

func (t *subHashTable) assocKeyValuePair(
	kv *keyValuePair,
	idx, pos, shift, hash uint32,
	key, value interface{},
	persistent bool,
) (*subHashTable, bool) {
	if kv.keyEquals(key) {
		if !persistent {
			kv.value = value
			return t, true
		}
		sht := t.clone()
		sht.replace(pos, &keyValuePair{hash, key, value})
		return sht, true
	}

	// there is a hash collision
	parent := t
	if persistent {
		parent = t.clone()
	}
	sht := resolveHashCollision(shift+5, kv, &keyValuePair{hash, key, value})
	if persistent {
		parent.replace(pos, sht)
	} else {
		parent.base[pos] = sht
	}

	return parent, false
}

func (t *subHashTable) lookup(shift, hash uint32, key interface{}) (interface{}, bool) {
	if shift >= 32 {
		hash = rehash(key, shift/5)
		shift = 0
	}

	idx := mask(hash, shift)
	if bit := (t.bitmap >> idx) & 0x01; bit == 0 {
		return nil, false
	}

	pos := t.lookupPos(idx)
	switch n := t.base[pos]; n.(type) {
	case *subHashTable:
		return n.(*subHashTable).lookup(shift+5, hash, key)
	case *keyValuePair:
		if kv := n.(*keyValuePair); kv.keyEquals(key) {
			return kv.value, true
		}
	}
	return nil, false
}

// resolveHashCollision resolves a hash collision in the most efficient way possible.
// kv1 is assumed to be the kv that already was in the tree and kv2 the one that is being inserted.
func resolveHashCollision(shift uint32, kv1, kv2 *keyValuePair) *subHashTable {
	if shift >= 32 {
		kv1 = &keyValuePair{rehash(kv1.key, shift/5), kv1.key, kv1.value}
		kv2.hash = rehash(kv2.key, shift/5)
		shift = 0
	}

	idx1 := mask(kv1.hash, shift)
	idx2 := mask(kv2.hash, shift)

	t := newSubHashTable()
	if idx1 != idx2 {
		t.base = make([]interface{}, 2)
		t.setBitmap(idx1)
		t.setBitmap(idx2)
		if idx1 > idx2 {
			t.base[0] = kv2
			t.base[1] = kv1
		} else {
			t.base[0] = kv1
			t.base[1] = kv2
		}
		return t
	}

	sht := resolveHashCollision(shift+5, kv1, kv2)
	t.setBitmap(idx1)
	t.base = []interface{}{sht}
	return t
}

type keyValuePair struct {
	hash  uint32
	key   interface{}
	value interface{}
}

func (k *keyValuePair) keyEquals(key interface{}) bool {
	return equals(k.key, key)
}

func mask(hash, shift uint32) uint32 {
	return (hash >> shift) & 0x01f
}

// popcnt implementation for 32 bit integers
// see: http://graphics.stanford.edu/~seander/bithacks.html#CountBitsSetParallel

const (
	sk5  uint32 = 0x55555555
	sk3  uint32 = 0x33333333
	skf0 uint32 = 0x0F0F0F0F
	skff uint32 = 0x00FF00FF
	sk0f uint32 = 0x0000FFFF
)

func popcnt(m uint32) uint32 {
	m -= (m >> 1) & sk5
	m = ((m >> 2) & sk3) + (m & sk3)
	m = ((m >> 4) + m) & skf0
	m = ((m >> 8) + m) & skff
	return ((m >> 16) + m) & sk0f
}
