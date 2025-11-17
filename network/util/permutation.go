package util

type RcPermutation struct {
	virtToPhys []uint16
	physToVirt []uint16
}

// newRCPermutation은 `size` 크기의 항등 순열(identity permutation)을 생성합니다.
func newRCPermutation(size int) *RcPermutation {
	s := uint16(size)
	virtToPhys := make([]uint16, size)
	physToVirt := make([]uint16, size)
	for i := uint16(0); i < s; i++ {
		virtToPhys[i] = i
		physToVirt[i] = i
	}
	return &RcPermutation{
		virtToPhys: virtToPhys,
		physToVirt: physToVirt,
	}
}

// index는 '논리적' 인덱스 `a`에 해당하는 '물리적' 인덱스를 반환합니다.
func (p *RcPermutation) index(a int) int {
	return int(p.virtToPhys[a])
}

// swap은 '논리적' 인덱스 `a`와 `b`를 교환합니다.
func (p *RcPermutation) swap(a, b int) {
	p.virtToPhys[a], p.virtToPhys[b] = p.virtToPhys[b], p.virtToPhys[a]

	physB := p.virtToPhys[a]
	physA := p.virtToPhys[b]

	p.physToVirt[physA], p.physToVirt[physB] = p.physToVirt[physB], p.physToVirt[physA]
}
