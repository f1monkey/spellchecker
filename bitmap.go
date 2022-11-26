package spellchecker

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

func (b *bitmap) has(id uint32) bool {
	return *b&(1<<id) > 0
}

func (b *bitmap) countDiff(b2 bitmap) int {
	if *b == b2 {
		return 0
	}

	cnt := 0
	for i := 0; i < 64; i++ {
		if b.has(uint32(i)) != b2.has(uint32(i)) {
			cnt++
		}
	}

	return cnt
}

func (b *bitmap) clone() bitmap {
	return *b
}
