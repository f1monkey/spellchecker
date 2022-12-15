package spellchecker

import "math/bits"

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
	return bits.OnesCount64(uint64(*b) ^ uint64(b2))
}

func (b *bitmap) clone() bitmap {
	return *b
}
