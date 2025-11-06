package util

import (
	"bytes"
	"fmt"
)

// DenseMatrix는 가우스 소거법에 사용되는 밀집 행렬입니다.
type DenseMatrix struct {
	data  []bool
	nrows int
	ncols int
}

// NewDenseMatrix는 모든 원소가 `elem` 값으로 채워진 행렬을 생성합니다.
func NewDenseMatrix(nrows, ncols int, elem bool) *DenseMatrix {
	data := make([]bool, nrows*ncols)
	if elem {
		for i := range data {
			data[i] = true
		}
	}
	return &DenseMatrix{
		data:  data,
		nrows: nrows,
		ncols: ncols,
	}
}

// NewDenseMatrixFromFn은 `f(i, j)` 함수로 행렬을 생성합니다.
func NewDenseMatrixFromFn(nrows, ncols int, f func(i, j int) bool) *DenseMatrix {
	data := make([]bool, 0, nrows*ncols)
	for i := 0; i < nrows; i++ {
		for j := 0; j < ncols; j++ {
			data = append(data, f(i, j))
		}
	}
	return &DenseMatrix{
		data:  data,
		nrows: nrows,
		ncols: ncols,
	}
}

func (m *DenseMatrix) Rows() int { return m.nrows }
func (m *DenseMatrix) Cols() int { return m.ncols }

func (m *DenseMatrix) checkBounds(i, j int) {
	if i < 0 || i >= m.nrows {
		panic(fmt.Sprintf("matrix: row index %d out of bounds (nrows: %d)", i, m.nrows))
	}
	if j < 0 || j >= m.ncols {
		panic(fmt.Sprintf("matrix: col index %d out of bounds (ncols: %d)", j, m.ncols))
	}
}

func (m *DenseMatrix) At(i, j int) bool {
	m.checkBounds(i, j)
	return m.data[i*m.ncols+j]
}

func (m *DenseMatrix) Set(i, j int, val bool) {
	m.checkBounds(i, j)
	m.data[i*m.ncols+j] = val
}

func (m *DenseMatrix) String() string {
	var buf bytes.Buffer
	buf.WriteString("\n")
	for i := 0; i < m.nrows; i++ {
		buf.WriteString("  |")
		for j := 0; j < m.ncols; j++ {
			if m.At(i, j) {
				buf.WriteString(" 1")
			} else {
				buf.WriteString(" 0")
			}
		}
		buf.WriteString(" |\n")
	}
	return buf.String()
}

// RowwiseEliminationGaussianFullPivot는 '전체 피벗팅' 전략을 사용합니다.
func (m *DenseMatrix) RowwiseEliminationGaussianFullPivot(
	rowOperation func(op RowOperation),
) error {
	// `eliminationStrategyFn` (전체 피벗팅 전략)
	eliminationStrategyFn := func(a *rcSwapMatrix, step int) (row int, col int, ok bool) {
		var bestRow, bestLeadCol int
		bestWeight := -1 // `None`

		// '논리적' 행 `step`부터 `nrows-1`까지 순회
		for r := step; r < a.Rows(); r++ {
			weight := 0
			leadColumn := -1 // `None`

			// '논리적' 열 `0`부터 `ncols-1`까지 순회
			for c := 0; c < a.Cols(); c++ {
				if a.at(r, c) { // `at`이 논리적 인덱싱(permuted)을 수행
					weight++
					if leadColumn == -1 {
						leadColumn = c
					}
				}
			}

			if weight != 0 {
				if bestWeight == -1 || weight < bestWeight {
					bestRow = r
					bestWeight = weight
					bestLeadCol = leadColumn
				}
			}
		}

		if bestWeight != -1 {
			return bestRow, bestLeadCol, true // (row, lead_column)
		}
		return 0, 0, false // 피벗을 찾지 못함
	}

	return m.rowwiseEliminationSchedule(rowOperation, eliminationStrategyFn)
}

// rowwiseEliminationSchedule는 가우스 소거법의 메인 스케줄러입니다.
func (m *DenseMatrix) rowwiseEliminationSchedule(
	rowOperation func(op RowOperation),
	eliminationStrategyFn func(a *rcSwapMatrix, step int) (row int, col int, ok bool),
) error {
	if m.nrows < m.ncols {
		return fmt.Errorf("rowwiseElimination: nrows (%d) must be >= ncols (%d)", m.nrows, m.ncols)
	}

	a := newRCSwapMatrix(m)

	for step := 0; step < a.Cols(); step++ {
		// 1. 피벗 찾기
		row, col, ok := eliminationStrategyFn(a, step)
		if !ok {
			return nil
		}

		// 2. 피벗을 (step, step) 위치로 이동 (논리적 교환)
		if row != step {
			a.swapRows(row, step)
		}
		if col != step {
			a.swapColumns(col, step)
		}

		// 3. `step` 행을 사용하여 다른 모든 행을 소거
		for i := 0; i < a.Rows(); i++ {
			if i != step && a.at(i, step) {
				// `Row[i] = Row[i] XOR Row[step]` (물리적 연산)
				a.rowSubAssign(i, step)

				// 수행된 연산을 *물리적* 인덱스로 보고
				rowOperation(RowOperationSubAssign{
					I: a.rowPermutation.index(i),
					J: a.rowPermutation.index(step),
				})
			}
		}
	}

	return nil
}

// rcSwapMatrix는 행/열 교환(Pivoting)을 효율적으로 추적하는 래퍼입니다.
type rcSwapMatrix struct {
	mat               *DenseMatrix
	rowPermutation    *RcPermutation
	columnPermutation *RcPermutation
}

// newRCSwapMatrix는 `DenseMatrix`로부터 `RCSwapMatrix`를 생성합니다.
func newRCSwapMatrix(mat *DenseMatrix) *rcSwapMatrix {
	return &rcSwapMatrix{
		mat:               mat,
		rowPermutation:    newRCPermutation(mat.Rows()),
		columnPermutation: newRCPermutation(mat.Cols()),
	}
}

func (rcs *rcSwapMatrix) Rows() int { return rcs.mat.Rows() }
func (rcs *rcSwapMatrix) Cols() int { return rcs.mat.Cols() }

func (rcs *rcSwapMatrix) swapRows(a, b int) {
	rcs.rowPermutation.swap(a, b)
}

func (rcs *rcSwapMatrix) swapColumns(a, b int) {
	rcs.columnPermutation.swap(a, b)
}

// rowSubAssign는 `Row[a] = Row[a] XOR Row[b]` 연산을 수행합니다.
func (rcs *rcSwapMatrix) rowSubAssign(a, b int) {
	aPhys := rcs.rowPermutation.index(a)
	bPhys := rcs.rowPermutation.index(b)

	for k := 0; k < rcs.mat.Cols(); k++ {
		// (열(k)은 물리적 인덱스로 순회)
		aVal := rcs.mat.At(aPhys, k)
		bVal := rcs.mat.At(bPhys, k)
		rcs.mat.Set(aPhys, k, aVal != bVal) // aVal ^ bVal
	}
}

// at은 '논리적' 인덱스 (i, j)의 값을 반환합니다.
func (rcs *rcSwapMatrix) at(i, j int) bool {
	physRow := rcs.rowPermutation.index(i)
	physCol := rcs.columnPermutation.index(j)
	return rcs.mat.At(physRow, physCol)
}

type RowOperation interface{ isRowOperation() }
type RowOperationSubAssign struct {
	I int
	J int
}

func (r RowOperationSubAssign) isRowOperation() {}
