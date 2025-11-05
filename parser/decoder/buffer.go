package decoder

import (
	"errors"
	"fmt"
	"monad-flow/util"
)

// buffer는 '두뇌'가 관리하는 단일 버퍼(방정식)의 상태입니다.
// Rust의 `Buffer` struct를 포팅한 것입니다.
type buffer struct {
	// 이 버퍼에 XOR로 연결된 모든 중간 심볼(변수)의 ID 목록
	intermediateSymbolIDs util.OrderedSet

	// intermediateSymbolIDs 중 Active 또는 Used 상태인 심볼의 개수 (필링용)
	activeUsedWeight uint16

	// 이 버퍼가 Used 상태의 심볼 하나만 가리키는지 여부
	used bool
}

// newBuffer는 비어있는 새 버퍼 상태를 생성합니다.
// Rust의 `Buffer::new()`에 해당합니다.
func newBuffer() *buffer {
	return &buffer{
		intermediateSymbolIDs: util.NewOrderedSet(),
		activeUsedWeight:      0,
		used:                  false,
	}
}

// appendIntermediateSymbolID는 버퍼에 심볼 ID를 추가합니다.
// Rust의 `append_intermediate_symbol_id`에 해당합니다.
func (b *buffer) appendIntermediateSymbolID(id uint16, incrementActiveUsedWeight bool) {
	// [최적화] Rust의 append는 정렬된 상태로 삽입됨을 가정합니다.
	b.intermediateSymbolIDs.Append(id)
	if incrementActiveUsedWeight {
		b.activeUsedWeight++
	}
}

// firstIntermediateSymbolID는 이 버퍼와 연결된 첫 번째 심볼 ID를 반환합니다.
// (weight=1일 때 호출됨을 가정)
func (b *buffer) firstIntermediateSymbolID() (uint16, bool) {
	return b.intermediateSymbolIDs.First()
}

// xorEq는 'other' 버퍼의 심볼 목록을 'b'의 심볼 목록과 Set-XOR합니다.
// Rust의 `Buffer::xor_eq`에 해당합니다.
// [주의] Rust 주석에 따라, 이 함수는 `activeUsedWeight`는 건드리지 않습니다.
//
//	호출자(lowLevelDecoder)가 `activeUsedWeight`를 별도로 갱신해야 합니다.
func (b *buffer) xorEq(other *buffer) {
	for _, id := range other.intermediateSymbolIDs.Values() {
		b.intermediateSymbolIDs.InsertOrRemove(id)
	}
}

// [새 함수 추가] isPaired는 버퍼의 active_used_weight가 1인지 확인합니다.
// (Rust의 `is_paired` 헬퍼 함수로 추정)
func (b *buffer) isPaired() bool {
	return b.activeUsedWeight == 1
}

// --- BufferId ---

// bufferIdType은 버퍼가 임시 버퍼인지 수신된 청크 버퍼인지 구분합니다.
// Rust의 `BufferId` enum (TempBuffer, ReceiveBuffer)에 해당합니다.
type bufferIdType int

const (
	// tempBuffer는 가우스 소거법 등에 사용되는 임시 버퍼입니다.
	tempBuffer bufferIdType = iota
	// receiveBuffer는 네트워크에서 수신된 실제 청크 페이로드를 저장한 버퍼입니다.
	receiveBuffer
)

// bufferId는 bufferSet 내의 특정 버퍼를 가리키는 핸들입니다.
// Rust의 `BufferId` enum의 두 variant를 Go struct로 표현합니다.
type bufferId struct {
	Type  bufferIdType
	Index int // TempBuffer의 경우 0..numTempBuffers-1
	// ReceiveBuffer의 경우 0..num_received_buffers-1 (상대 인덱스)
}

// --- BufferSet ---

// bufferSet은 디코딩 과정에 필요한 모든 바이트 버퍼를 관리합니다.
// Rust의 `BufferSet` 구조체에 해당하며, 실제 XOR 연산을 수행합니다.
type bufferSet struct {
	// buffers는 모든 버퍼를 저장하는 슬라이스입니다.
	// [0 .. numTempBuffers-1] : 임시 버퍼
	// [numTempBuffers .. ] : 수신된 청크 버퍼
	buffers        [][]byte
	symbolSize     int // T, 각 버퍼(청크)의 고정된 바이트 크기
	numTempBuffers int // 임시 버퍼의 개수
}

// newBufferSet는 새 버퍼셋을 생성하고, 필요한 임시 버퍼를 미리 할당합니다.
// Rust의 `BufferSet::new`에 해당합니다.
func newBufferSet(symbolSize int, numTempBuffers int) *bufferSet {
	// 임시 버퍼들을 미리 할당하고 0으로 초기화합니다.
	buffers := make([][]byte, numTempBuffers)
	for i := 0; i < numTempBuffers; i++ {
		buffers[i] = make([]byte, symbolSize)
	}

	return &bufferSet{
		buffers:        buffers,
		symbolSize:     symbolSize,
		numTempBuffers: numTempBuffers,
	}
}

// bufferIndex는 `bufferId`를 `buffers` 슬라이스의 실제 인덱스로 변환합니다.
// Rust의 `buffer_index` 헬퍼 함수에 해당합니다.
func (bs *bufferSet) bufferIndex(id bufferId) int {
	if id.Type == tempBuffer {
		// 임시 버퍼의 인덱스는 0부터 시작
		return id.Index
	}
	// 수신 버퍼의 인덱스는 임시 버퍼 개수 이후부터 시작
	return bs.numTempBuffers + id.Index
}

// addReceiveBuffer는 수신된 새 청크 페이로드(payload)를 버퍼셋에 추가합니다.
// [!] 중요: 이 함수는 payload의 *복사본*을 만들어 저장합니다.
//
//	(원본 eBPF 버퍼가 재사용될 수 있으므로)
//
// Rust의 `push_buffer`에 해당하며, `managedDecoder`가 사용할
// *상대 인덱스* (relative index)를 반환합니다.
func (bs *bufferSet) addReceiveBuffer(payload []byte) (int, error) {
	if len(payload) != bs.symbolSize {
		return 0, fmt.Errorf("invalid symbol size: expected %d, got %d",
			bs.symbolSize, len(payload))
	}

	// 페이로드의 복사본을 만들어 슬라이스에 추가합니다.
	buf := make([]byte, bs.symbolSize)
	copy(buf, payload)

	bs.buffers = append(bs.buffers, buf)

	// 방금 추가된 버퍼의 '상대 인덱스'를 반환합니다.
	// (전체 길이) - (0-based 인덱싱 1) - (임시 버퍼 개수)
	relativeIndex := len(bs.buffers) - 1 - bs.numTempBuffers
	return relativeIndex, nil
}

// xorBuffers는 `a = a ^ b` 연산을 수행합니다.
// `bufferId` (핸들)를 받아 실제 버퍼를 찾아 바이트 단위 XOR를 실행합니다.
// Rust의 `xor_buffers`에 해당합니다.
func (bs *bufferSet) xorBuffers(a, b bufferId) error {
	aIndex := bs.bufferIndex(a)
	bIndex := bs.bufferIndex(b)

	if aIndex == bIndex {
		return errors.New("xorBuffers: cannot XOR buffer with itself")
	}

	// Bounds check
	if aIndex < 0 || aIndex >= len(bs.buffers) {
		return fmt.Errorf("xorBuffers: bufferId 'a' is out of bounds (index %d)", aIndex)
	}
	if bIndex < 0 || bIndex >= len(bs.buffers) {
		return fmt.Errorf("xorBuffers: bufferId 'b' is out of bounds (index %d)", bIndex)
	}

	dst := bs.buffers[aIndex] // Destination (a)
	src := bs.buffers[bIndex] // Source (b)

	if len(dst) != bs.symbolSize || len(src) != bs.symbolSize {
		return fmt.Errorf("xorBuffers: buffer size mismatch (dst: %d, src: %d, expected: %d)",
			len(dst), len(src), bs.symbolSize)
	}

	// TODO: 이 루프는 향후 SIMD 최적화(e.g., []uint64 캐스팅)가 가능합니다.
	// 현재는 바이트 단위로 정확하게 구현합니다.
	for i := 0; i < bs.symbolSize; i++ {
		dst[i] ^= src[i]
	}

	return nil
}

// buffer는 `bufferId`에 해당하는 버퍼의 (읽기 전용) 슬라이스를 반환합니다.
// Rust의 `buffer` 메서드에 해당합니다.
func (bs *bufferSet) buffer(id bufferId) ([]byte, error) {
	index := bs.bufferIndex(id)

	if index < 0 || index >= len(bs.buffers) {
		return nil, fmt.Errorf("buffer: bufferId is out of bounds (index %d)", index)
	}

	return bs.buffers[index], nil
}

// ========================================================================
//  bufferWeightMap (이번 단계에서 새로 추가됨)
// ========================================================================

// bufferWeightMap은 'buffer'의 가중치를 기반으로 하는 Min-Heap (우선순위 큐)입니다.
// Rust의 `BufferWeightMap`을 포팅한 것입니다.
type bufferWeightMap struct {
	// 힙 배열 (i번째 노드에 buffer_index 저장)
	heapIndexToBufferIndex []uint16

	// 룩업 테이블: buffer_index -> weight (0 = 힙에 없음)
	bufferIndexToWeight []uint16

	// 룩업 테이블: buffer_index -> heap_index (힙 배열 내 위치)
	bufferIndexToHeapIndex []uint16
}

// newBufferWeightMap은 힙을 초기화합니다.
func newBufferWeightMap(initialCapacity int) *bufferWeightMap {
	if initialCapacity < 16 {
		initialCapacity = 16
	}
	return &bufferWeightMap{
		heapIndexToBufferIndex: make([]uint16, 0, initialCapacity),
		bufferIndexToWeight:    make([]uint16, initialCapacity),
		bufferIndexToHeapIndex: make([]uint16, initialCapacity),
	}
}

// ensureCapacity는 룩업 테이블이 bufferIndex에 접근할 수 있도록 크기를 보장합니다.
func (bwm *bufferWeightMap) ensureCapacity(bufferIndex int) {
	if bufferIndex >= len(bwm.bufferIndexToWeight) {
		newCap := bufferIndex + 1
		if newCap < (len(bwm.bufferIndexToWeight) * 2) {
			newCap = len(bwm.bufferIndexToWeight) * 2
		}

		newWeights := make([]uint16, newCap)
		copy(newWeights, bwm.bufferIndexToWeight)
		bwm.bufferIndexToWeight = newWeights

		newHeapIndices := make([]uint16, newCap)
		copy(newHeapIndices, bwm.bufferIndexToHeapIndex)
		bwm.bufferIndexToHeapIndex = newHeapIndices
	}
}

// --- 힙 내부 헬퍼 ---

// weightAtHeapIndex는 힙 인덱스를 이용해 가중치를 조회합니다.
func (bwm *bufferWeightMap) weightAtHeapIndex(heapIndex int) uint16 {
	bufferIndex := bwm.heapIndexToBufferIndex[heapIndex]
	weight := bwm.bufferIndexToWeight[bufferIndex]
	if weight == 0 {
		panic(fmt.Sprintf("weightAtHeapIndex: found weight 0 for heap index %d (buffer index %d)", heapIndex, bufferIndex))
	}
	return weight
}

// swap은 힙 내의 두 노드를 교환하고 룩업 테이블을 갱신합니다.
func (bwm *bufferWeightMap) swap(heapIndexA, heapIndexB int) {
	bufferIndexA := bwm.heapIndexToBufferIndex[heapIndexA]
	bufferIndexB := bwm.heapIndexToBufferIndex[heapIndexB]

	// 1. 힙 배열 교환
	bwm.heapIndexToBufferIndex[heapIndexA], bwm.heapIndexToBufferIndex[heapIndexB] = bufferIndexB, bufferIndexA

	// 2. 역방향 룩업 테이블 갱신
	bwm.bufferIndexToHeapIndex[bufferIndexA] = uint16(heapIndexB)
	bwm.bufferIndexToHeapIndex[bufferIndexB] = uint16(heapIndexA)
}

// pullUp은 노드를 힙 위로(부모 방향) 올립니다. (Heapify-up)
func (bwm *bufferWeightMap) pullUp(heapIndex int) {
	for heapIndex > 0 {
		parentHeapIndex := (heapIndex - 1) / 2
		if bwm.weightAtHeapIndex(parentHeapIndex) <= bwm.weightAtHeapIndex(heapIndex) {
			break // 부모가 더 작거나 같으면 힙 속성 만족
		}
		bwm.swap(heapIndex, parentHeapIndex)
		heapIndex = parentHeapIndex
	}
}

// pushDown은 노드를 힙 아래로(자식 방향) 내립니다. (Heapify-down)
func (bwm *bufferWeightMap) pushDown(heapIndex int) {
	n := len(bwm.heapIndexToBufferIndex)
	for {
		heapIndexMin := heapIndex
		childHeapIndex1 := 2*heapIndex + 1
		if childHeapIndex1 < n && bwm.weightAtHeapIndex(childHeapIndex1) < bwm.weightAtHeapIndex(heapIndexMin) {
			heapIndexMin = childHeapIndex1
		}

		childHeapIndex2 := 2*heapIndex + 2
		if childHeapIndex2 < n && bwm.weightAtHeapIndex(childHeapIndex2) < bwm.weightAtHeapIndex(heapIndexMin) {
			heapIndexMin = childHeapIndex2
		}

		if heapIndex == heapIndexMin {
			break // 자식들이 더 크거나 같으면 힙 속성 만족
		}

		bwm.swap(heapIndex, heapIndexMin)
		heapIndex = heapIndexMin
	}
}

// removeHeapIndex는 힙의 특정 인덱스에 있는 노드를 제거합니다.
func (bwm *bufferWeightMap) removeHeapIndex(heapIndex int, bufferIndex uint16) {
	lastHeapIndex := len(bwm.heapIndexToBufferIndex) - 1

	// 1. 제거할 노드를 힙의 마지막 노드와 교환 (마지막 노드가 아니라면)
	if heapIndex != lastHeapIndex {
		bwm.swap(heapIndex, lastHeapIndex)
	}

	// 2. 힙 배열에서 마지막 노드(제거 대상) 제거
	bwm.heapIndexToBufferIndex = bwm.heapIndexToBufferIndex[:lastHeapIndex]

	// 3. 룩업 테이블에서 제거
	prevWeight := bwm.bufferIndexToWeight[bufferIndex]
	bwm.bufferIndexToWeight[bufferIndex] = 0 // 0 = None (힙에 없음)

	// 4. (제거 대상이 마지막 노드가 아니었을 경우)
	//    교환되어 heapIndex로 이동한 노드의 힙 속성을 복원
	if heapIndex != lastHeapIndex {
		// Rust: `match weight.cmp(&prev_weight)`
		// 새로 이동한 노드의 가중치가 이전에 있던 노드의 가중치와 비교
		// 1. 새로 이동한 노드의 (현재) 가중치
		currentWeight := bwm.weightAtHeapIndex(heapIndex)
		// 2. 이전에 있던 노드의 가중치

		if currentWeight < prevWeight {
			bwm.pullUp(heapIndex)
		} else if currentWeight > prevWeight {
			bwm.pushDown(heapIndex)
		}
		// 같으면 아무것도 안 함
	}
}

// --- 힙 공개 API ---

func (bwm *bufferWeightMap) isEmpty() bool {
	return len(bwm.heapIndexToBufferIndex) == 0
}

// peekMin은 힙의 최소값(루트)을 (제거하지 않고) 반환합니다.
// (bufferIndex, weight, ok)
func (bwm *bufferWeightMap) peekMin() (uint16, uint16, bool) {
	if bwm.isEmpty() {
		return 0, 0, false
	}
	bufferIndex := bwm.heapIndexToBufferIndex[0]
	weight := bwm.bufferIndexToWeight[bufferIndex]
	return bufferIndex, weight, true
}

// insertBufferWeight는 새 버퍼와 가중치를 힙에 추가합니다.
func (bwm *bufferWeightMap) insertBufferWeight(bufferIndex int, weight uint16) {
	if weight == 0 {
		panic("insertBufferWeight: weight cannot be 0")
	}

	heapIndex := len(bwm.heapIndexToBufferIndex)
	bwm.heapIndexToBufferIndex = append(bwm.heapIndexToBufferIndex, uint16(bufferIndex))

	bwm.ensureCapacity(bufferIndex)

	if bwm.bufferIndexToWeight[bufferIndex] != 0 {
		panic(fmt.Sprintf("insertBufferWeight: buffer %d already in heap", bufferIndex))
	}
	bwm.bufferIndexToWeight[bufferIndex] = weight
	bwm.bufferIndexToHeapIndex[bufferIndex] = uint16(heapIndex)

	bwm.pullUp(heapIndex)
}

// removeMin은 힙의 최소값(루트)을 제거합니다.
func (bwm *bufferWeightMap) removeMin() {
	if !bwm.isEmpty() {
		bufferIndex := bwm.heapIndexToBufferIndex[0]
		bwm.removeHeapIndex(0, bufferIndex)
	}
}

// removeBufferWeight는 특정 bufferIndex를 힙에서 제거합니다.
func (bwm *bufferWeightMap) removeBufferWeight(bufferIndex int) (uint16, bool) {
	bwm.ensureCapacity(bufferIndex) // 룩업 테이블 접근 전 용량 확보
	weight := bwm.bufferIndexToWeight[bufferIndex]
	if weight == 0 {
		return 0, false // 힙에 없음
	}

	heapIndex := bwm.bufferIndexToHeapIndex[bufferIndex]
	bwm.removeHeapIndex(int(heapIndex), uint16(bufferIndex))
	return weight, true
}

// updateBufferWeight는 힙에 있는 버퍼의 가중치를 갱신합니다.
func (bwm *bufferWeightMap) updateBufferWeight(bufferIndex int, weight uint16) {
	if weight == 0 {
		panic("updateBufferWeight: weight cannot be 0")
	}

	bwm.ensureCapacity(bufferIndex)
	prevWeight := bwm.bufferIndexToWeight[bufferIndex]
	if prevWeight == 0 {
		panic(fmt.Sprintf("updateBufferWeight: buffer %d not in heap", bufferIndex))
	}

	if prevWeight == weight {
		return // 가중치 변경 없음
	}

	bwm.bufferIndexToWeight[bufferIndex] = weight
	heapIndex := int(bwm.bufferIndexToHeapIndex[bufferIndex])

	if weight < prevWeight {
		bwm.pullUp(heapIndex)
	} else {
		bwm.pushDown(heapIndex)
	}
}

// enumerate는 힙의 모든 항목에 대해 함수를 실행합니다 (순서 보장 없음).
func (bwm *bufferWeightMap) enumerate(callback func(bufferIndex uint16, weight uint16)) {
	for _, bufferIndex := range bwm.heapIndexToBufferIndex {
		weight := bwm.bufferIndexToWeight[bufferIndex]
		callback(bufferIndex, weight)
	}
}
