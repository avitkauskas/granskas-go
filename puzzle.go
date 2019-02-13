package main

import (
	"fmt"
	"sort"
	"strings"
)

const boardSize = 6
const topRow uint64 = 0xFF00000000000000
const bottomRow = topRow >> (8 * (boardSize - 1))
const leftColumn uint64 = 0x8080808080808080
const rightColumn = leftColumn >> (boardSize - 1)

type solution map[int]uint64 // pieceIndex -> piecePosition

var board = [...]string{
	"K", "L", " ", "U", "V", "A",
	"C", " ", "T", " ", " ", "E",
	"D", "L", " ", "U", "U", " ",
	"I", " ", "S", " ", "M", "G",
	" ", "K", "M", " ", " ", "Š",
	"C", " ", "K", "L", "R", "U",
}

var pieces = [...]uint64{
	0xE0C0800000000000, // 11100000_11000000_10000000...
	0xF060000000000000, // 11110000_00110000_10000000...
	0xF080800000000000, // 11110000_10000000_10000000...
	0x70E0000000000000, // 01110000_11100000_00000000...
	0xF0C0000000000000, // 11110000_11000000_00000000...
	0x4070C00000000000, // 01000000_01110000_11000000...
}

var pieceSymbols = [...]rune{'-', ':', '+', '=', '*', '@'}

var puzzles = []string{
	"ASU", "ISM", "KSU", "KTU", "KU", "LCC", "LEU", "LKA", "LMTA",
	"LMSU", "LSU", "MRU", "ŠU", "VDA", "VDU", "VGTU", "VU",
}

var sortedPuzzles []string
var emptyField uint64

func setSortedPuzzles() {
	for _, puzzle := range puzzles {
		splitPuzzle := strings.Split(puzzle, "")
		sort.Strings(splitPuzzle)
		sortedPuzzle := strings.Join(splitPuzzle, "")
		sortedPuzzles = append(sortedPuzzles, sortedPuzzle)
	}
}

func setEmptyField() {
	emptyField = 0x0000000000000000
	for i := uint(1); i <= 8 - boardSize; i++ {
		emptyField |= rightColumn >> i
		emptyField |= bottomRow >> (8 * i)
	}
}

func canShiftLeft(u uint64) bool {
	return u & leftColumn == 0
}

func canShiftRight(u uint64) bool {
	return u & rightColumn == 0
}

func canShiftUp(u uint64) bool {
	return u & topRow == 0
}

func canShiftDown(u uint64) bool {
	return u & bottomRow == 0
}

func shiftRight(u uint64) uint64 {
	return u >> 1
}

func shiftDown(u uint64) uint64 {
	return u >> 8
}

func flushLeft(u uint64) uint64 {
	for canShiftLeft(u) {
		u <<= 1
	}
	return u
}

func flushTop(u uint64) uint64 {
	for canShiftUp(u) {
		u <<= 8
	}
	return u
}

func flipDiagonal(u uint64) uint64 {
	const k1 uint64 = 0xaa00aa00aa00aa00
	const k2 uint64 = 0xcccc0000cccc0000
	const k4 uint64 = 0xf0f0f0f00f0f0f0f
	t := u ^ (u << 36)
	u = u ^ (k4 & (t ^ (u >> 36)))
	t = k2 & (u ^ (u << 18))
	u = u ^ (t ^ (t >> 18))
	t = k1 & (u ^ (u << 9))
	u = u ^ (t ^ (t >> 9))
	return u
}

func flipVertical(u uint64) uint64 {
	const k1 uint64 = 0x00FF00FF00FF00FF
	const k2 uint64 = 0x0000FFFF0000FFFF
	u = ((u >> 8) & k1) | ((u & k1) << 8)
	u = ((u >> 16) & k2) | ((u & k2) << 16)
	u = (u >> 32) | (u << 32)
	return u
}

func rotate90(u uint64) uint64 {
	return flipVertical(flipDiagonal(u))
}

func flipAndFlush(u uint64) uint64 {
	return flushLeft(flushTop(flipVertical(u)))
}

func rotateAndFlush(u uint64) uint64 {
	return flushLeft(flushTop(rotate90(u)))
}

func getPiecePositions(u uint64) map[uint64]struct{} {
	member := struct{}{}
	positions := make(map[uint64]struct{})

	// put the piece itself flushed to top left
	u = flushTop(flushLeft(u))
	positions[u] = member

	// put flipped version
	u = flipAndFlush(u)
	positions[u] = member

	copyPositions := make(map[uint64]struct{}, len(positions))
	for k, v := range positions {
		copyPositions[k] = v
	}

	// get all rotations of both flip sides
	for p := range copyPositions {
		for i := 0; i < 3; i++ {
			p = rotateAndFlush(p)
			positions[p] = member
		}
	}

	copyPositions = make(map[uint64]struct{}, len(positions))
	for k, v := range positions {
		copyPositions[k] = v
	}

	// get all shifts down and right of all rotations
	for p := range copyPositions {
		for a := p; canShiftRight(a); {
			a = shiftRight(a)
			positions[a] = member
		}
		for a := p; canShiftDown(a); {
			a = shiftDown(a)
			positions[a] = member
			for b := a; canShiftRight(b); {
				b = shiftRight(b)
				positions[b] = member
			}
		}
	}
	return positions
}

func getSolutionsWithoutPiece(p int) []solution {
	return make([]solution, 1) // DUMMY
}

func checkForSolution(field uint64) (puzzleIndex int, found bool) {
	var uncoveredCells []string
	for i := uint(1); i < 64; i++ {
		if field & (1 << i) != 0 {
			idx := 63 - i
			idx = (idx / 8) * 6 + idx % 8
			uncoveredCells = append(uncoveredCells, board[idx])
		}
	}
	sort.Strings(uncoveredCells)
	uncoveredLetters := strings.Join(uncoveredCells, "")
	uncoveredLetters = strings.TrimSpace(uncoveredLetters)
	for idx, puzzle := range sortedPuzzles {
		if puzzle == uncoveredLetters {
			return idx, true
		}
	}
	return 0, false
}

//func printField(u uint64) {
//	for i := uint(0); i < 8; i++ {
//		mask := uint64(0xFF00000000000000) >> (i * 8)
//		fmt.Printf("%08b\n", (u & mask) >> ((7 - i) * 8))
//	}
//	fmt.Println()
//}

func main() {

	// initialize globals
	setSortedPuzzles()
	setEmptyField()

	piecePositions := make([]map[uint64]struct{}, len(pieces))
	for i, v := range pieces {
		piecePositions[i] = getPiecePositions(v)
	}

	solutions := make([][]solution, len(pieces))
	for i := range pieces {
		solutions[i] = getSolutionsWithoutPiece(i)
	}

	fmt.Println(checkForSolution(0x000000F800000000)) // DUMMY
}