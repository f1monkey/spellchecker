package dictionary

type Index map[bitmap][]uint32

func (i Index) get(b bitmap) []uint32 {
	return i[b]
}

type bitmap int64

func (b *bitmap) or(id uint32) {
	*b |= (1 << id)
}

func (b *bitmap) xor(id uint32) {
	*b ^= (1 << id)
}

func (b *bitmap) isEmpty() bool {
	return *b == 0
}

func (b *bitmap) countDiff(b2 bitmap) int {
	if *b == b2 {
		return 0
	}

	cnt := 0
	for i := 0; i < 64; i++ {
		if *b&(1<<i) != b2&(1<<i) {
			cnt++
		}
	}

	return cnt
}

func (b *bitmap) clone() bitmap {
	return *b
}
