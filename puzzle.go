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

var board = [...]string{
	"K", "L", " ", "U", "V", "A",
	"C", " ", "T", " ", " ", "E",
	"D", "L", " ", "U", "U", " ",
	"I", " ", "S", " ", "M", "G",
	" ", "K", "M", " ", " ", "Š",
	"C", " ", "K", "L", "R", "U",
}

//var pieces = [...]uint64{ // original Granskas set
//	0xE0C0800000000000, // 11100000_11000000_10000000...
//	0xF060000000000000, // 11110000_00110000_00000000...
//	0xF080800000000000, // 11110000_10000000_10000000...
//	0x70E0000000000000, // 01110000_11100000_00000000...
//	0xF0C0000000000000, // 11110000_11000000_00000000...
//	0x4070C00000000000, // 01000000_01110000_11000000...
//}

var pieces = [...]uint64{
	0x80F0800000000000, // 10000000_11110000_10000000... 01) cube unfolding
	0x80F0400000000000, // 10000000_11110000_01000000... 02) cube unfolding
	0x80F0200000000000, // 10000000_11110000_00100000... 03) cube unfolding
	0x80F0100000000000, // 10000000_11110000_00010000... 04) cube unfolding
	0x40F0200000000000, // 01000000_11110000_00100000... 05) cube unfolding
	0x40F0400000000000, // 01000000_11110000_01000000... 06) cube unfolding
	0xC070400000000000, // 11000000_01110000_01000000... 07) cube unfolding
	0xC070200000000000, // 11000000_01110000_00100000... 08) cube unfolding
	0x80E0300000000000, // 10000000_11100000_00110000... 09) cube unfolding
	0xC060300000000000, // 11000000_01100000_00110000... 10) cube unfolding
	0xE038000000000000, // 11100000_00111000_00000000... 11) cube unfolding
	0xFC00000000000000, // 11111100_00000000_00000000... 12)
	0x80F8000000000000, // 10000000_11111000_00000000... 13)
	0x40F8000000000000, // 01000000_11111000_00000000... 14)
	0x20F8000000000000, // 00100000_11111000_00000000... 15)
	0xC078000000000000, // 11000000_01111000_00000000... 16)
	0xC0F0000000000000, // 11000000_11110000_00000000... 17)
	0xA0F0000000000000, // 10100000_11110000_00000000... 18)
	0x90F0000000000000, // 10010000_11110000_00000000... 19)
	0x60F0000000000000, // 01100000_11110000_00000000... 20)
	0x8080F00000000000, // 10000000_10000000_11110000... 21)
	0x4040F00000000000, // 01000000_01000000_11110000... 22)
	0x40C0700000000000, // 01000000_11000000_01110000... 23)
	0xD070000000000000, // 11010000_01110000_00000000... 24)
	0xE070000000000000, // 11100000_01110000_00000000... 25)
	0xE0E0000000000000, // 11100000_11100000_00000000... 26)
	0xC0E0800000000000, // 11000000_11100000_10000000... 27)
	0xC040700000000000, // 11000000_01000000_01110000... 28)
	0x80C0700000000000, // 10000000_11000000_01110000... 29)
	0xC080E00000000000, // 11000000_10000000_11100000... 30)
	0xC040E00000000000, // 11000000_01000000_11100000... 31)
	0xC060C00000000000, // 11000000_01100000_11000000... 32)
	0xE0C0800000000000, // 11100000_11000000_10000000... 33)
	0xC0E0400000000000, // 11000000_11100000_01000000... 34)
	0x20E0C00000000000, // 00100000_11100000_11000000... 35)
}

var pieceSymbols = [...]rune{'-', ':', '+', '=', '*', '@'}

var piecePositions = make([]map[uint64]struct{}, len(pieces))
func setPiecePositions() {
	for i, v := range pieces {
		piecePositions[i] = getPiecePositions(v)
	}
}

var puzzles = []string{
	"ASU", "ISM", "KSU", "KTU", "KU", "LCC", "LEU", "LKA", "LMTA",
	"LMSU", "LSU", "MRU", "ŠU", "VDA", "VDU", "VGTU", "VU",
}

var sortedPuzzles []string
func setSortedPuzzles() {
	for _, puzzle := range puzzles {
		splitPuzzle := strings.Split(puzzle, "")
		sort.Strings(splitPuzzle)
		sortedPuzzle := strings.Join(splitPuzzle, "")
		sortedPuzzles = append(sortedPuzzles, sortedPuzzle)
	}
}

var emptyField uint64
func setEmptyField() {
	emptyField = 0x0000000000000000
	for i := uint(1); i <= 8 - boardSize; i++ {
		emptyField |= rightColumn >> i
		emptyField |= bottomRow >> (8 * i)
	}
}

type solution map[int]uint64 // pieceIndex -> piecePosition
var solutions = make([][][]solution, len(pieces))
func makeEmptySolutions() {
	for i := range solutions {
		solutions[i] = make([][]solution, len(puzzles))
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

func makeRange(min, max int) []int {
	a := make([]int, max - min + 1)
	for i := range a {
		a[i] = min + i
	}
	return a
}

func getSolutionsWithoutPiece(withoutPiece int) {
	usedPieces := makeRange(0, len(pieces) - 1)
	usedPieces = append(usedPieces[:withoutPiece], usedPieces[withoutPiece+1:]...)

	field := emptyField
	findSolutions(field, withoutPiece, len(usedPieces),0, usedPieces, make(solution))
}

func findSolutions(field uint64, withoutPiece, totalPieces,
		pieceIndex int, usedPieces []int, usedPositions solution) {

	piece := usedPieces[pieceIndex]
	for pos := range piecePositions[piece] {
		if field & pos == 0 { // can put piece?
			field |= pos // put piece
			usedPositions[piece] = pos
			if pieceIndex == totalPieces - 1 { // was it last piece?
				puzzleIndex, found := checkForSolution(field)
				if found {
					solutions[withoutPiece][puzzleIndex] =
						append(solutions[withoutPiece][puzzleIndex], usedPositions)
				}
				field ^= pos // remove piece and try it's next position
				delete(usedPositions, piece)
			} else { // it was not the last piece
				findSolutions(field, withoutPiece, totalPieces, pieceIndex + 1, usedPieces, usedPositions)
				field ^= pos // remove piece and try it's next position
				delete(usedPositions, piece)
			}
		}
	}
}

func checkForSolution(field uint64) (puzzleIndex int, found bool) {
	field ^= 0xFFFFFFFFFFFFFFFF
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
	setPiecePositions()
	makeEmptySolutions()

	for i := range pieces {
		getSolutionsWithoutPiece(i)
	}

	for i, wp := range solutions {
		fmt.Println("Without piece", i)
		for j, s := range wp {
			fmt.Println("Puzzle", puzzles[j], "has", len(s), "solutions")
		}
	}
}