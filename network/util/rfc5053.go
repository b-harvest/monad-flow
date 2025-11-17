package util

import (
	"errors"
	"fmt"
	"math/bits"
	"sort"
)

const MAX_DEGREE = 40

func deg(v uint32) uint8 {
	// RFC 5053 section 5.4.4.2
	switch {
	case v <= 10240:
		return 1
	case v <= 491581:
		return 2
	case v <= 712793:
		return 3
	case v <= 831694:
		return 4
	case v <= 948445:
		return 10
	case v <= 1032188:
		return 11
	case v <= 1048575: // MAX_V - 1
		return 40
	default:
		panic(fmt.Sprintf("Can't find Deg(%d)", v))
	}
}

func rand(x uint16, i uint8, m uint32) uint32 {
	// Rand[X, i, m] = (V0[(X + i) % 256] ^ V1[(floor(X/256)+ i) % 256]) % m
	xU32 := uint32(x)
	iU32 := uint32(i)

	v0Index := (xU32 + iU32) & 0xFF
	v1Index := ((xU32 >> 8) + iU32) & 0xFF

	if m == 0 {
		return 0
	}

	return (V0[v0Index] ^ V1[v1Index]) % m
}

func NewCodeParameters(numSourceSymbols int) (*CodeParameters, error) {
	if numSourceSymbols < SourceSymbolsMin || numSourceSymbols > SourceSymbolsMax {
		return nil, fmt.Errorf("numSourceSymbols %d not in range %d..%d",
			numSourceSymbols, SourceSymbolsMin, SourceSymbolsMax)
	}

	k := uint16(numSourceSymbols)

	// 1. X 결정
	x, err := determineX(k)
	if err != nil {
		return nil, fmt.Errorf("failed to determine X: %w", err)
	}

	// 2. S (NumLdpcSymbols) 결정
	s, err := determineNumLdpcSymbols(k, x)
	if err != nil {
		return nil, fmt.Errorf("failed to determine S: %w", err)
	}

	// 3. H (NumHalfSymbols) 결정
	h, err := determineNumHalfSymbols(k, s)
	if err != nil {
		return nil, fmt.Errorf("failed to determine H: %w", err)
	}

	// 4. L (NumIntermediateSymbols) 계산
	// L = K + S + H
	l := k + s + uint16(h)

	// 5. L' (NumIntermediateSymbolsPrime) 결정
	lp, err := smallestPrimeGreaterOrEqual(l)
	if err != nil {
		return nil, fmt.Errorf("failed to determine L': %w", err)
	}

	// 6. J (SystematicIndex) 결정
	j, err := determineSystematicIndex(k)
	if err != nil {
		return nil, fmt.Errorf("failed to determine J: %w", err)
	}

	return &CodeParameters{
		NumSourceSymbols:            k,
		NumLdpcSymbols:              s,
		NumHalfSymbols:              h,
		NumIntermediateSymbols:      l,
		NumIntermediateSymbolsPrime: lp,
		SystematicIndex:             j,
	}, nil
}

func determineX(numSourceSymbols uint16) (uint8, error) {
	k2 := uint64(numSourceSymbols) * 2
	for x := uint64(xMin); x <= uint64(xMax); x++ {
		if x*(x-1) >= k2 {
			return uint8(x), nil
		}
	}
	return 0, fmt.Errorf("can't find x for num_source_symbols = %d", numSourceSymbols)
}

func determineNumLdpcSymbols(numSourceSymbols uint16, x uint8) (uint16, error) {
	sMin := ((numSourceSymbols + 99) / 100) + uint16(x)
	return smallestPrimeGreaterOrEqual(sMin)
}

func choose(n, k uint8) uint16 {
	if k > n {
		return 0
	}

	// c = (n! / k!)
	var c uint32 = 1
	for val := uint32(k + 1); val <= uint32(n); val++ {
		c *= val
	}

	// c = c / (n - k)!
	for val := uint32(2); val <= uint32(n-k); val++ {
		c /= val
	}

	return uint16(c)
}

func determineNumHalfSymbols(numSourceSymbols, numLdpcSymbols uint16) (uint8, error) {
	ks := numSourceSymbols + numLdpcSymbols

	for h := uint8(halfMin); h <= uint8(halfMax); h++ {
		if choose(h, (h+1)/2) >= ks {
			return h, nil
		}
	}
	return 0, fmt.Errorf("can't find H for K=%d, S=%d", numSourceSymbols, numLdpcSymbols)
}

func smallestPrimeGreaterOrEqual(primeMin uint16) (uint16, error) {
	index := sort.Search(len(SMALL_PRIMES), func(i int) bool {
		return SMALL_PRIMES[i] >= primeMin
	})

	if index == len(SMALL_PRIMES) {
		return 0, fmt.Errorf("can't find small prime >= %d in table", primeMin)
	}
	return SMALL_PRIMES[index], nil
}

func determineSystematicIndex(numSourceSymbols uint16) (uint16, error) {
	index := int(numSourceSymbols) - SourceSymbolsMin

	if index < 0 || index >= len(SYSTEMATIC_INDEX) {
		return 0, fmt.Errorf("can't find systematic index for num_source_symbols = %d (index %d out of bounds)",
			numSourceSymbols, index)
	}

	return SYSTEMATIC_INDEX[index], nil
}

type CodeParameters struct {
	NumSourceSymbols            uint16 // K (원본 심볼 수)
	NumLdpcSymbols              uint16 // S (LDPC 심볼 수)
	NumHalfSymbols              uint8  // H (Half 심볼 수)
	NumIntermediateSymbols      uint16 // L (중간 심볼 수, K + S + H)
	NumIntermediateSymbolsPrime uint16 // L' (L보다 크거나 같은 가장 작은 소수)
	SystematicIndex             uint16 // J(K)
}

// GHalf는 G_Half 행렬의 요소를 생성합니다 (RFC 5053 섹션 5.4.2.3).
func (cp *CodeParameters) GHalf(setElement func(h, j int)) {
	// h_prime = ceil(H / 2)
	hPrime := (cp.NumHalfSymbols + 1) >> 1

	var i uint64 = 0
	hPrimeInt := int(hPrime)

	mNext := func() uint64 {
		for {
			gI := i ^ (i >> 1)
			i++

			if bits.OnesCount64(gI) == hPrimeInt {
				return gI
			}
		}
	}

	jLimit := int(cp.NumSourceSymbols) + int(cp.NumLdpcSymbols)

	for j := 0; j < jLimit; j++ {
		m := mNext()

		for h := 0; h < int(cp.NumHalfSymbols); h++ {
			if (m & (1 << h)) != 0 {
				setElement(h, j)
			}
		}
	}
}

// ldpcTriple은 RFC 5053 섹션 5.4.2.3의 트리플을 생성합니다.
func (cp *CodeParameters) ldpcTriple(sourceSymbol int) (int, int, int) {
	s := int(cp.NumLdpcSymbols)
	if s <= 1 {
		if s == 0 {
			return 0, 0, 0
		}
		return 0, sourceSymbol % s, (sourceSymbol % s) % s
	}

	a := 1 + ((sourceSymbol / s) % (s - 1))
	b1 := sourceSymbol % s
	b2 := (b1 + a) % s
	b3 := (b2 + a) % s

	return b1, b2, b3
}

// GLdpc는 G_LDPC 행렬의 요소를 생성합니다 (RFC 5053 섹션 5.4.2.3).
func (cp *CodeParameters) GLdpc(setElement func(bufferIndex, symbolIndex int)) {
	k := int(cp.NumSourceSymbols)
	b := make([]int, 3)

	for i := 0; i < k; i++ {
		b1, b2, b3 := cp.ldpcTriple(i)
		b[0], b[1], b[2] = b1, b2, b3
		sort.Ints(b)
		for _, el := range b {
			setElement(el, i)
		}
	}
}

// trip은 RFC 5053 섹션 5.4.4.4의 Triple Generator 함수입니다.
func (cp *CodeParameters) trip(encodingSymbolID uint16) (d uint8, a uint16, b uint16, err error) {
	const q = 65521

	j := uint64(cp.SystematicIndex)
	valA := (53591 + j*997) % q

	valB := (10267 * (j + 1)) % q

	x := uint64(encodingSymbolID)
	y := uint16((valB + x*valA) % q)

	v := rand(y, 0, 1<<20)
	d = deg(v)

	lp := uint32(cp.NumIntermediateSymbolsPrime)
	if lp == 0 {
		return 0, 0, 0, errors.New("trip: NumIntermediateSymbolsPrime is 0")
	}

	valALT := 1 + rand(y, 1, lp-1)

	valBLT := rand(y, 2, lp)

	return d, uint16(valALT), uint16(valBLT), nil
}

// LTSequenceOp는 RFC 5053 섹션 5.4.4.3의 LT 시퀀스를 계산합니다.
func (cp *CodeParameters) LTSequenceOp(encodingSymbolID int, setElement func(intermediateSymbolID int)) error {
	d, a, b, err := cp.trip(uint16(encodingSymbolID))
	if err != nil {
		return fmt.Errorf("trip failed: %w", err)
	}

	dInt := int(d)
	aInt := int(a)
	bInt := int(b)

	l := int(cp.NumIntermediateSymbols)
	lp := int(cp.NumIntermediateSymbolsPrime)

	numSymbols := dInt
	if numSymbols > l {
		numSymbols = l
	}

	if numSymbols > MAX_DEGREE {
		numSymbols = MAX_DEGREE
	}

	var symbols [MAX_DEGREE]uint16

	for i := 0; i < numSymbols; i++ {
		for bInt >= l {
			bInt = (bInt + aInt) % lp
		}
		symbols[i] = uint16(bInt)
		bInt = (bInt + aInt) % lp
	}

	sort.Slice(symbols[:numSymbols], func(i, j int) bool {
		return symbols[i] < symbols[j]
	})

	for i := 0; i < numSymbols; i++ {
		setElement(int(symbols[i]))
	}
	return nil
}

func (cp *CodeParameters) GLT(setElement func(i, el int), nrows int) error {
	for i := 0; i < nrows; i++ {
		err := cp.LTSequenceOp(i, func(el int) {
			setElement(i, el)
		})
		if err != nil {
			return fmt.Errorf("LTSequenceOp failed for row %d: %w", i, err)
		}
	}
	return nil
}
