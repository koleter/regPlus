// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// backtrack is a regular expression search with submatch
// tracking for small regular expressions and texts. It allocates
// a bit vector with (length of input) * (length of prog) bits,
// to make sure it never explores the same (character position, instruction)
// state multiple times. This limits the search to run in time linear in
// the length of the test.
//
// backtrack is a fast replacement for the NFA code on small
// regexps when onepass cannot be used.

package regPlus

import (
	"github.com/koleter/regPlus/syntax"
)

// A job is an entry on the backtracker's job stack. It holds
// the instruction pc and the position in the input.
type job struct {
	pc  uint32
	arg bool
	pos int

	aux interface{}
	f   func()
}

const (
	visitedBits        = 32
	maxBacktrackProg   = 500        // len(prog.Inst) <= max
	maxBacktrackVector = 256 * 1024 // bit vector size <= max (bits)
)

// bitState holds state for the backtracker.
type bitState struct {
	end      int
	cap      []int
	matchcap []int
	jobs     []job
	visited  []uint32

	inputs inputs
}

//var bitStatePool sync.Pool

func newBitState() *bitState {
	return new(bitState)
}

func (re *Regexp) freeBitState() {
	b := re.b
	jobs := b.jobs
	for i := len(b.jobs) - 1; i >= 0; i-- {
		if jobs[i].f != nil {
			jobs[i].f()
		}
	}
	re.finalFree()
}

// maxBitStateLen returns the maximum length of a string to search with
// the backtracker using prog.
func maxBitStateLen(prog *syntax.Prog) int {
	if !shouldBacktrack(prog) {
		return 0
	}
	return maxBacktrackVector / len(prog.Inst)
}

// shouldBacktrack reports whether the program is too
// long for the backtracker to run.
func shouldBacktrack(prog *syntax.Prog) bool {
	return len(prog.Inst) <= maxBacktrackProg
}

// reset resets the state of the backtracker.
// end is the end position in the input.
// ncap is the number of captures.
func (b *bitState) reset(prog *syntax.Prog, end int, ncap int) {
	b.end = end

	if cap(b.jobs) == 0 {
		b.jobs = make([]job, 0, 256)
	} else {
		b.jobs = b.jobs[:0]
	}

	visitedSize := (len(prog.Inst)*(end+1) + visitedBits - 1) / visitedBits
	if cap(b.visited) < visitedSize {
		b.visited = make([]uint32, visitedSize, maxBacktrackVector/visitedBits)
	} else {
		b.visited = b.visited[:visitedSize]
		for i := range b.visited {
			b.visited[i] = 0
		}
	}

	if cap(b.cap) < ncap {
		b.cap = make([]int, ncap)
	} else {
		b.cap = b.cap[:ncap]
	}
	for i := range b.cap {
		b.cap[i] = -1
	}

	if cap(b.matchcap) < ncap {
		b.matchcap = make([]int, ncap)
	} else {
		b.matchcap = b.matchcap[:ncap]
	}
	for i := range b.matchcap {
		b.matchcap[i] = -1
	}
}

// shouldVisit reports whether the combination of (pc, pos) has not
// been visited yet.
func (b *bitState) shouldVisit(pc uint32, pos int) bool {
	n := uint(int(pc)*(b.end+1) + pos)
	if b.visited[n/visitedBits]&(1<<(n&(visitedBits-1))) != 0 {
		return false
	}
	b.visited[n/visitedBits] |= 1 << (n & (visitedBits - 1))
	return true
}

// push pushes (pc, pos, arg) onto the job stack if it should be
// visited.
func (b *bitState) push(re *Regexp, pc uint32, pos int, arg bool) {
	// Only check shouldVisit when arg is false.
	// When arg is true, we are continuing a previous visit.
	if re.prog.Inst[pc].Op != syntax.InstFail && (arg || b.shouldVisit(pc, pos)) {
		b.jobs = append(b.jobs, job{pc: pc, arg: arg, pos: pos})
	}
}

// tryBacktrack runs a backtracking search starting at pos.
func (re *Regexp) tryBacktrack(b *bitState, i input, pc uint32, pos int) bool {
	longest := re.longest

	b.push(re, pc, pos, false)
	return re.run(b, i, longest)
}

func (re *Regexp) run(b *bitState, i input, longest bool) bool {
Loop:
	for len(b.jobs) > 0 {
		l := len(b.jobs) - 1
		// Pop job off the stack.
		curjob := b.jobs[l]
		b.jobs = b.jobs[:l]
		if curjob.f != nil {
			curjob.f()
			continue
		}
		pc := curjob.pc
		pos := curjob.pos
		arg := curjob.arg

		// Optimization: rather than push and pop,
		// code that is going to Push and continue
		// the loop simply updates ip, p, and arg
		// and jumps to CheckAndLoop. We have to
		// do the ShouldVisit check that Push
		// would have, but we avoid the stack
		// manipulation.
		goto Skip
	CheckAndLoop:
		if !b.shouldVisit(pc, pos) {
			continue
		}
	Skip:

		inst := re.prog.Inst[pc]

		switch inst.Op {
		default:
			panic("bad inst")
		case syntax.InstFail:
			panic("unexpected InstFail")
		case syntax.InstAlt:
			// Cannot just
			//   b.push(inst.Out, pos, false)
			//   b.push(inst.Arg, pos, false)
			// If during the processing of inst.Out, we encounter
			// inst.Arg via another path, we want to process it then.
			// Pushing it here will inhibit that. Instead, re-push
			// inst with arg==true as a reminder to push inst.Arg out
			// later.
			if arg {
				// Finished inst.Out; try inst.Arg.
				arg = false
				pc = inst.Arg
				goto CheckAndLoop
			} else {
				b.push(re, pc, pos, true)
				pc = inst.Out
				goto CheckAndLoop
			}

		case syntax.InstAltMatch:
			// One opcode consumes runes; the other leads to match.
			switch re.prog.Inst[inst.Out].Op {
			case syntax.InstRune, syntax.InstRune1, syntax.InstRuneAny, syntax.InstRuneAnyNotNL:
				// inst.Arg is the match.
				b.push(re, inst.Arg, pos, false)
				pc = inst.Arg
				pos = b.end
				goto CheckAndLoop
			}
			// inst.Out is the match - non-greedy
			b.push(re, inst.Out, b.end, false)
			pc = inst.Out
			goto CheckAndLoop

		case syntax.InstRune:
			r, width := i.step(pos)
			if !inst.MatchRune(r) {
				continue
			}
			pos += width
			pc = inst.Out
			goto CheckAndLoop

		case syntax.InstRune1:
			r, width := i.step(pos)
			if r != inst.Rune[0] {
				continue
			}
			pos += width
			pc = inst.Out
			goto CheckAndLoop

		case syntax.InstRuneAnyNotNL:
			r, width := i.step(pos)
			if r == '\n' || r == endOfText {
				continue
			}
			pos += width
			pc = inst.Out
			goto CheckAndLoop

		case syntax.InstRuneAny:
			r, width := i.step(pos)
			if r == endOfText {
				continue
			}
			pos += width
			pc = inst.Out
			goto CheckAndLoop

		case syntax.InstCapture:
			if arg {
				// Finished inst.Out; restore the old value.
				b.cap[inst.Arg] = pos
				continue
			} else {
				if inst.Arg < uint32(len(b.cap)) {
					// Capture pos to register, but save old value.
					b.push(re, pc, b.cap[inst.Arg], true) // come back when we're done.
					b.cap[inst.Arg] = pos
				}
				pc = inst.Out
				goto CheckAndLoop
			}

		case syntax.InstEmptyWidth:
			flag := i.context(pos)
			if !flag.match(syntax.EmptyOp(inst.Arg)) {
				continue
			}
			pc = inst.Out
			goto CheckAndLoop

		case syntax.InstNop:
			pc = inst.Out
			goto CheckAndLoop

		case syntax.InstMatch:
			for _, treeNode := range re.stringVar {
				if treeNode.count < treeNode.min {
					continue Loop
				}
			}

			for _, regNode := range re.regVar {
				if regNode.count < regNode.min {
					continue Loop
				}
			}

			// We found a match. If the caller doesn't care
			// where the match is, no point going further.
			if len(b.cap) == 0 {
				return true
			}

			// Record best match so far.
			// Only need to check end point, because this entire
			// call is only considering one start position.
			if len(b.cap) > 1 {
				b.cap[1] = pos
			}
			if old := b.matchcap[1]; old == -1 || (longest && pos > 0 && pos > old) {
				copy(b.matchcap, b.cap)
			}

			// If going for first match, we're done.
			if !longest {
				return true
			}

			// If we used the entire text, no longer match is possible.
			if pos == b.end {
				return true
			}

			// Otherwise, continue on in hope of a longer match.
			continue
		case syntax.InstStringVar:
			if arg {
				arg = false
				node := curjob.aux.(*node)
				for r, width := i.step(pos); r != endOfText; r, width = i.step(pos) {
					node = node.Next[r]
					if node == nil {
						continue Loop
					}
					pos += width
					if node.Cnt > 0 {
						node.Cnt--
						b.jobs = append(b.jobs, job{pc: pc, arg: true, pos: pos, aux: node})
						b.jobs = append(b.jobs, job{f: func() {
							node.Cnt++
						}})
						pc = inst.Out
						goto CheckAndLoop
					}
				}
			} else {
				treeNode := re.stringVar[inst.Str]
				if treeNode == nil {
					panic("string var " + inst.Str + " is unregistered")
				}
				if treeNode.count >= treeNode.max {
					continue
				}
				treeNode.count++
				b.jobs = append(b.jobs, job{f: func() {
					treeNode.count--
				}})
				b.jobs = append(b.jobs, job{pc: pc, arg: true, pos: pos, aux: treeNode.root})
			}
			continue
		case syntax.InstRegVar:
			type researchReg struct {
				re           *Regexp
				b            *bitState
				maxSearchEnd int
			}
			if arg {
				arg = false
				switch node := curjob.aux.(type) {
				case *researchReg:
					regexp := node.re
					bit := node.b
					end := b.end

					if !regexp.run(bit, i, regexp.longest) {
						for j := node.maxSearchEnd; j < end; j++ {
							bit = regexp.backtrackForRegVar(i, j, end, 2, []int{})
							if bit != nil {
								node.b = bit
								goto match
							}
						}
						continue
					}
				match:
					if bit.cap[1] > node.maxSearchEnd {
						node.maxSearchEnd = bit.cap[1]
					}
					if bit.cap[0] == bit.cap[1] {
						node.maxSearchEnd++
					}
					b.jobs = append(b.jobs, job{pc: pc, arg: true, pos: pos, aux: node})
					pc = inst.Out
					pos = bit.cap[1]
					goto CheckAndLoop
				case *Element:
					for ; node != nil; node = node.Next() {
						regexp := node.Value.(*Regexp)

						bit := regexp.backtrackForRegVar(i, pos, b.end, 2, []int{})
						if bit == nil {
							continue
						}
						b.jobs = append(b.jobs, job{pc: pc, arg: true, pos: pos, aux: node.Next()})

						if prev := node.Prev(); prev != nil {
							b.jobs = append(b.jobs, job{f: func() {
								prev.InsertElementAfter(node)
							}})
						} else if next := node.Next(); next != nil {
							b.jobs = append(b.jobs, job{f: func() {
								next.InsertElementBefore(node)
							}})
						} else {
							l := node.list
							b.jobs = append(b.jobs, job{f: func() {
								l.root.InsertElementBefore(node)
							}})
						}
						if bit.cap[0] != bit.cap[1] {
							b.jobs = append(b.jobs, job{pc: pc, arg: true, pos: pos, aux: &researchReg{re: regexp, b: bit, maxSearchEnd: bit.matchcap[1]}})
						}
						node.RemoveSelf()
						pc = inst.Out
						pos = bit.matchcap[1]
						goto CheckAndLoop
					}
				}
			} else {
				regNode := re.regVar[inst.Str]
				if regNode == nil {
					panic("string var " + inst.Str + " is unregistered")
				}
				if regNode.count >= regNode.max {
					continue
				}
				regNode.count++
				b.jobs = append(b.jobs, job{f: func() {
					regNode.count--
				}})
				b.jobs = append(b.jobs, job{pc: pc, arg: true, pos: pos, aux: regNode.l.Front()})
			}
			continue
		}
	}

	return longest && len(b.matchcap) > 1 && b.matchcap[1] >= 0
}

// backtrack runs a backtracking search of prog on the input starting at pos.
func (re *Regexp) backtrack(ib []byte, is string, pos int, ncap int, dstCap []int) []int {
	startCond := re.cond
	if startCond == ^syntax.EmptyOp(0) { // impossible
		return nil
	}
	if startCond&syntax.EmptyBeginText != 0 && pos != 0 {
		// Anchored match, past beginning of text.
		return nil
	}

	b := newBitState()
	i, end := b.inputs.init(nil, ib, is)
	b.reset(re.prog, end, ncap)

	re.b = b

	// Anchored search must start at the beginning of the input
	if startCond&syntax.EmptyBeginText != 0 {
		if len(b.cap) > 0 {
			b.cap[0] = pos
		}
		if !re.tryBacktrack(b, i, uint32(re.prog.Start), pos) {
			re.freeBitState()
			return nil
		}
	} else {

		// Unanchored search, starting from each possible text position.
		// Notice that we have to try the empty string at the end of
		// the text, so the loop condition is pos <= end, not pos < end.
		// This looks like it's quadratic in the size of the text,
		// but we are not clearing visited between calls to TrySearch,
		// so no work is duplicated and it ends up still being linear.
		width := -1
		for ; pos <= end && width != 0; pos += width {
			if len(re.prefix) > 0 {
				// Match requires literal prefix; fast search for it.
				advance := i.index(re, pos)
				if advance < 0 {
					re.freeBitState()
					return nil
				}
				pos += advance
			}

			if len(b.cap) > 0 {
				b.cap[0] = pos
			}
			if re.tryBacktrack(b, i, uint32(re.prog.Start), pos) {
				// Match must be leftmost; done.
				goto Match
			}
			_, width = i.step(pos)
		}
		re.freeBitState()
		return nil
	}

Match:
	dstCap = append(dstCap, b.matchcap...)
	re.freeBitState()
	return dstCap
}

// backtrack runs a backtracking search of prog on the input starting at pos.
func (re *Regexp) backtrackForRegVar(i input, pos, end int, ncap int, dstCap []int) *bitState {
	startCond := re.cond
	if startCond == ^syntax.EmptyOp(0) { // impossible
		return nil
	}

	b := newBitState()
	b.reset(re.prog, end, ncap)
	re.b = b

	// Anchored search must start at the beginning of the input
	if startCond&syntax.EmptyBeginText != 0 {
		if len(b.cap) > 0 {
			b.cap[0] = pos
		}
		if !re.tryBacktrack(b, i, uint32(re.prog.Start), pos) {
			return nil
		}
	} else {

		// Unanchored search, starting from each possible text position.
		// Notice that we have to try the empty string at the end of
		// the text, so the loop condition is pos <= end, not pos < end.
		// This looks like it's quadratic in the size of the text,
		// but we are not clearing visited between calls to TrySearch,
		// so no work is duplicated and it ends up still being linear.
		width := -1
		for ; pos <= end && width != 0; pos += width {
			if len(re.prefix) > 0 {
				// Match requires literal prefix; fast search for it.
				advance := i.index(re, pos)
				if advance < 0 {
					return nil
				}
				pos += advance
			}

			if len(b.cap) > 0 {
				b.cap[0] = pos
			}
			if re.tryBacktrack(b, i, uint32(re.prog.Start), pos) {
				// Match must be leftmost; done.
				goto Match
			}
			_, width = i.step(pos)
		}
		return nil
	}

Match:
	dstCap = append(dstCap, b.matchcap...)
	return b
}
