package util

import (
	// (g_half 포팅을 위해 이 파일 상단에 추가)
	"errors"
	"fmt"
	"math/bits"
	"sort"
)

// [수정] Rust의 `deg.rs`에서 실제 MAX_DEGREE 값(40)을 확인했습니다.
const MAX_DEGREE = 40

// [수정] `deg` 함수의 스텁을 실제 코드로 교체합니다.
// (Rust의 `deg` 함수 포팅)
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
		// `rand` 함수가 `m = 1<<20` (1048576)으로 호출되므로,
		// `v`는 항상 0..1048575 범위에 있어야 합니다.
		panic(fmt.Sprintf("Can't find Deg(%d)", v))
	}
}

// [수정] `rand` 함수의 스텁을 실제 코드로 교체합니다.
// (Rust의 `rand` 함수 포팅)
func rand(x uint16, i uint8, m uint32) uint32 {
	// Rand[X, i, m] = (V0[(X + i) % 256] ^ V1[(floor(X/256)+ i) % 256]) % m

	// Rust의 `x.into()`와 `i.into()`는 Go에서 단순 캐스팅입니다.
	xU32 := uint32(x)
	iU32 := uint32(i)

	// (X + i) % 256
	v0Index := (xU32 + iU32) & 0xFF
	// (floor(X/256) + i) % 256
	v1Index := ((xU32 >> 8) + iU32) & 0xFF

	if m == 0 {
		// (방어 코드) 0으로 나누기 방지
		return 0
	}

	// [!!] V0, V1 테이블의 데이터가 필요합니다.
	return (V0[v0Index] ^ V1[v1Index]) % m
}

// NewCodeParameters는 K(numSourceSymbols)로부터 모든 파라미터를 계산합니다.
// Rust의 CodeParameters::new() 함수를 포팅한 것입니다.
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
	// [!!] 이 함수의 실제 구현은 Rust 코드에서 제공되지 않았습니다. (아래 주의사항 참고)
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

// --- 내부 헬퍼 함수 (Internal Helper Functions) ---

// determineX는 RFC 5053 5.4.2.3의 X 값을 계산합니다.
// "Let X be the smallest positive integer such that X*(X-1) >= 2*K"
// (Rust의 determine_x_slow 테스트 로직을 포팅, 범위가 작아 성능 문제 없음)
func determineX(numSourceSymbols uint16) (uint8, error) {
	k2 := uint64(numSourceSymbols) * 2 // u16 * 2는 u16 범위를 넘지 않지만, 비교를 위해 u64 사용
	for x := uint64(xMin); x <= uint64(xMax); x++ {
		if x*(x-1) >= k2 {
			return uint8(x), nil
		}
	}
	return 0, fmt.Errorf("can't find x for num_source_symbols = %d", numSourceSymbols)
}

// determineNumLdpcSymbols는 S (NumLdpcSymbols) 값을 계산합니다.
// "Let S be the smallest prime integer such that S >= ceil(0.01*K) + X"
func determineNumLdpcSymbols(numSourceSymbols uint16, x uint8) (uint16, error) {
	// ceil(0.01*K)는 (K + 99) / 100 와 동일 (정수 나눗셈)
	sMin := ((numSourceSymbols + 99) / 100) + uint16(x)
	return smallestPrimeGreaterOrEqual(sMin)
}

// choose는 조합(Combination) C(n, k)를 계산합니다.
// (Rust의 choose 함수를 직접 포팅)
func choose(n, k uint8) uint16 {
	if k > n {
		return 0
	}

	// Rust 구현은 C(n, k) = n! / (k! * (n-k)!) 를 다음과 같이 계산:
	// c = (n! / k!)
	var c uint32 = 1
	for val := uint32(k + 1); val <= uint32(n); val++ {
		c *= val
	}

	// c = c / (n - k)!
	for val := uint32(2); val <= uint32(n-k); val++ {
		c /= val
	}

	// Rust 주석에 따르면 결과는 항상 u16에 맞습니다.
	return uint16(c)
}

// determineNumHalfSymbols는 H (NumHalfSymbols) 값을 계산합니다.
// "Let H be the smallest integer such that choose(H,ceil(H/2)) >= K + S"
// (Rust의 determine_num_half_symbols_slow 테스트 로직을 포팅)
func determineNumHalfSymbols(numSourceSymbols, numLdpcSymbols uint16) (uint8, error) {
	ks := numSourceSymbols + numLdpcSymbols // u16 + u16, 오버플로우 가능성 있으나 K, S의 최대값(8192, 211)으론 문제 없음

	for h := uint8(halfMin); h <= uint8(halfMax); h++ {
		// C(h, ceil(h/2))
		// (h+1)/2 는 h/2의 정수 올림(ceiling)입니다.
		if choose(h, (h+1)/2) >= ks {
			return h, nil
		}
	}
	return 0, fmt.Errorf("can't find H for K=%d, S=%d", numSourceSymbols, numLdpcSymbols)
}

// smallestPrimeGreaterOrEqual는 primeMin보다 크거나 같은 가장 작은 소수를
// SMALL_PRIMES 테이블에서 이진 검색으로 찾습니다.
func smallestPrimeGreaterOrEqual(primeMin uint16) (uint16, error) {
	// sort.Search는 f(i)가 true인 가장 작은 인덱스 i를 찾습니다.
	index := sort.Search(len(SMALL_PRIMES), func(i int) bool {
		return SMALL_PRIMES[i] >= primeMin
	})

	if index == len(SMALL_PRIMES) {
		// 테이블의 모든 소수보다 큰 값이 요청됨
		return 0, fmt.Errorf("can't find small prime >= %d in table", primeMin)
	}
	return SMALL_PRIMES[index], nil
}

// determineSystematicIndex는 J(K) 값을 결정합니다.
// Rust의 SYSTEMATIC_INDEX 룩업 테이블을 기반으로 값을 찾습니다.
func determineSystematicIndex(numSourceSymbols uint16) (uint16, error) {
	// SourceSymbolsMin은 1입니다.
	// K=1일 때 인덱스 0, K=2일 때 인덱스 1...
	index := int(numSourceSymbols) - SourceSymbolsMin

	// SYSTEMATIC_INDEX 테이블의 경계값 확인
	if index < 0 || index >= len(SYSTEMATIC_INDEX) {
		return 0, fmt.Errorf("can't find systematic index for num_source_symbols = %d (index %d out of bounds)",
			numSourceSymbols, index)
	}

	return SYSTEMATIC_INDEX[index], nil
}

// CodeParameters는 RFC 5053의 핵심 파라미터를 저장합니다.
// 이 값들은 모두 K(NumSourceSymbols)로부터 파생됩니다.
type CodeParameters struct {
	NumSourceSymbols            uint16 // K (원본 심볼 수)
	NumLdpcSymbols              uint16 // S (LDPC 심볼 수)
	NumHalfSymbols              uint8  // H (Half 심볼 수)
	NumIntermediateSymbols      uint16 // L (중간 심볼 수, K + S + H)
	NumIntermediateSymbolsPrime uint16 // L' (L보다 크거나 같은 가장 작은 소수)
	SystematicIndex             uint16 // J(K)
}

// GHalf는 G_Half 행렬의 요소를 생성합니다 (RFC 5053 섹션 5.4.2.3).
// Rust의 `CodeParameters::g_half`를 포팅한 것입니다.
// `setElement` 콜백은 `setElement(h, j)` 형태로 호출됩니다 (h=행, j=열).
func (cp *CodeParameters) GHalf(setElement func(h, j int)) {
	// h_prime = ceil(H / 2)
	// (cp.NumHalfSymbols + 1) >> 1은 (H+1)/2의 정수 나눗셈으로, ceil(H/2)와 동일합니다.
	hPrime := (cp.NumHalfSymbols + 1) >> 1

	// Rust: `m_next` 클로저 로직
	// 그레이 코드를 생성하여, 1-비트의 개수가 hPrime과 같은 첫 번째 값을 찾습니다.
	var i uint64 = 0
	hPrimeInt := int(hPrime) // bits.OnesCount64가 int를 반환하므로 비교를 위해 캐스팅

	mNext := func() uint64 {
		for {
			gI := i ^ (i >> 1) // 그레이 코드 계산
			i++

			if bits.OnesCount64(gI) == hPrimeInt {
				return gI
			}
		}
	}

	// j = 0..K+S-1 까지 반복
	jLimit := int(cp.NumSourceSymbols) + int(cp.NumLdpcSymbols)

	for j := 0; j < jLimit; j++ {
		// m = m[j, H'] 계산
		m := mNext()

		for h := 0; h < int(cp.NumHalfSymbols); h++ {
			// m의 h번째 비트가 1인지 확인
			if (m & (1 << h)) != 0 {
				setElement(h, j) // (h = 행, j = 열)
			}
		}
	}
}

// [새 함수 추가] ldpcTriple은 RFC 5053 섹션 5.4.2.3의 트리플을 생성합니다.
// (Rust의 `ldpc_triple` 포팅)
func (cp *CodeParameters) ldpcTriple(sourceSymbol int) (int, int, int) {
	s := int(cp.NumLdpcSymbols)
	if s <= 1 {
		// (방어 코드) NumLdpcSymbols는 RFC 5053에 따라 1보다 큽니다.
		// (s-1)로 인한 0으로 나누기 방지
		if s == 0 {
			return 0, 0, 0
		}
		// s == 1
		return 0, sourceSymbol % s, (sourceSymbol % s) % s
	}

	a := 1 + ((sourceSymbol / s) % (s - 1))
	b1 := sourceSymbol % s
	b2 := (b1 + a) % s
	b3 := (b2 + a) % s

	return b1, b2, b3
}

// [새 함수 추가] GLdpc는 G_LDPC 행렬의 요소를 생성합니다 (RFC 5053 섹션 5.4.2.3).
// (Rust의 `g_ldpc` 포팅)
func (cp *CodeParameters) GLdpc(setElement func(bufferIndex, symbolIndex int)) {
	k := int(cp.NumSourceSymbols)
	// b는 3개 요소를 담는 슬라이스로 재사용
	b := make([]int, 3)

	for i := 0; i < k; i++ { // i = symbolIndex (0..K-1)
		b1, b2, b3 := cp.ldpcTriple(i)
		b[0], b[1], b[2] = b1, b2, b3

		// Rust: b.sort()
		sort.Ints(b)

		for _, el := range b { // el = bufferIndex (0..S-1)
			setElement(el, i)
		}
	}
}

// trip은 RFC 5053 섹션 5.4.4.4의 Triple Generator 함수입니다.
// Rust의 `CodeParameters::trip`을 포팅한 것입니다.
func (cp *CodeParameters) trip(encodingSymbolID uint16) (d uint8, a uint16, b uint16, err error) {
	// Q = 65521, a prime
	const q = 65521

	// A = (53591 + J(K)*997) % Q
	// (중간 계산 시 오버플로우를 방지하기 위해 uint64 사용)
	j := uint64(cp.SystematicIndex)
	valA := (53591 + j*997) % q

	// B = 10267*(J(K)+1) % Q
	valB := (10267 * (j + 1)) % q

	// Y = (B + X*A) % Q
	x := uint64(encodingSymbolID)
	y := uint16((valB + x*valA) % q)

	// --- (d 계산) ---
	// v = Rand[Y, 0, 2^^20]
	v := rand(y, 0, 1<<20)
	// d = Deg[v]
	d = deg(v)

	// --- (a, b 계산) ---
	lp := uint32(cp.NumIntermediateSymbolsPrime)
	if lp == 0 {
		return 0, 0, 0, errors.New("trip: NumIntermediateSymbolsPrime is 0")
	}

	// a = 1 + Rand[Y, 1, L’-1]
	valALT := 1 + rand(y, 1, lp-1)

	// b = Rand[Y, 2, L’]
	valBLT := rand(y, 2, lp)

	return d, uint16(valALT), uint16(valBLT), nil
}

// LTSequenceOp는 RFC 5053 섹션 5.4.4.3의 LT 시퀀스를 계산합니다.
// Rust의 `CodeParameters::lt_sequence_op`을 포팅한 것입니다.
func (cp *CodeParameters) LTSequenceOp(encodingSymbolID int, setElement func(intermediateSymbolID int)) error {
	d, a, b, err := cp.trip(uint16(encodingSymbolID))
	if err != nil {
		return fmt.Errorf("trip failed: %w", err)
	}

	dInt := int(d)
	aInt := int(a) // (uint16 -> int)
	bInt := int(b) // (uint16 -> int)

	l := int(cp.NumIntermediateSymbols)
	lp := int(cp.NumIntermediateSymbolsPrime)

	numSymbols := dInt
	if numSymbols > l {
		numSymbols = l
	}

	if numSymbols > MAX_DEGREE {
		// [방어 코드] trip이 MAX_DEGREE보다 큰 d를 반환하는 경우
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

	// Rust의 `symbols[0..num_symbols].sort()`
	sort.Slice(symbols[:numSymbols], func(i, j int) bool {
		return symbols[i] < symbols[j]
	})

	for i := 0; i < numSymbols; i++ {
		setElement(int(symbols[i]))
	}
	return nil
}

// GLT는 G_LT 행렬의 요소를 생성합니다. (현재 사용처 없음, 추후 receive_symbol에서 사용)
// Rust의 `CodeParameters::g_lt`을 포팅한 것입니다.
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
