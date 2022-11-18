package dictionary

type Index map[bitmap][]uint32

func (i Index) get(bm bitmap) []uint32 {
	return i[bm]
}

type bitmap int64

func (b *bitmap) or(id uint32) {
	*b |= (1 << id)
}

func (b *bitmap) xor(id uint32) {
	*b ^= (1 << id)
}

func (b *bitmap) clone() bitmap {
	return *b
}
