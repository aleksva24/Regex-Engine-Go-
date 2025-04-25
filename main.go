package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

type Regexp struct {
	matchers []Matcher
}

type Matcher interface {
	Match(s string, index int) int
}

type Literal struct {
	literal string
}

type AtLeastOneLiteral struct {
	matcher Matcher
	next    *Matcher
}
type ZeroOrOneLiterals struct {
	matcher Matcher
}
type ZeroOrMoreLiterals struct {
	matcher Matcher
	next    *Matcher
}
type StartsWith struct {
	matcher Matcher
}
type AnyLiteral Literal
type EndsWith struct {
	matcher Matcher
}

func (alo AtLeastOneLiteral) Match(s string, ind int) int {
	if alo.matcher.Match(s, ind) == -1 {
		return -1
	}

	var nextMatcher Matcher
	if alo.next != nil {
		nextMatcher = *alo.next
	}
	ind++

	for ind < len(s) {
		indOfNextMatcher := -1
		if nextMatcher != nil {
			indOfNextMatcher = nextMatcher.Match(s, ind)
		}

		newInd := alo.matcher.Match(s, ind)

		if newInd == -1 || indOfNextMatcher != -1 {
			return ind
		}
		ind = newInd
	}

	return ind - 1
}

func (ooz ZeroOrOneLiterals) Match(s string, ind int) int {
	if (ooz.matcher).Match(s, ind) == -1 {
		return ind
	} else {
		return ind + 1
	}
}

func (mto ZeroOrMoreLiterals) Match(s string, ind int) int {
	if (mto.matcher).Match(s, ind) == -1 {
		return ind
	}

	var nextMatcher Matcher
	if mto.next != nil {
		nextMatcher = *mto.next
	}

	ind++
	for ind < len(s) {
		indOfNextMatcher := -1
		if nextMatcher != nil {
			indOfNextMatcher = nextMatcher.Match(s, ind)
		}

		newInd := (mto.matcher).Match(s, ind)
		if newInd == -1 || indOfNextMatcher != -1 {
			return ind
		}
		ind = newInd
	}

	return ind - 1
}

func (l Literal) Match(s string, ind int) int {
	if l.literal != string(s[ind]) {
		return -1
	}

	return ind + 1
}

func (sw StartsWith) Match(s string, ind int) int {
	newInd := sw.matcher.Match(s, ind)
	if newInd == -1 {
		return -1
	}

	if ind != 0 {
		return -1
	}

	return newInd
}

func (ew EndsWith) Match(s string, ind int) int {
	if (ew.matcher).Match(s, ind) == -1 {
		return -1
	}

	if ind < len(s)-1 {
		return -1
	}

	return ind
}

func (al AnyLiteral) Match(s string, ind int) int {
	return ind + 1
}

func (re Regexp) Match(s string) bool {
	if re.matchers == nil {
		return true
	}

	i := 0

	for i < len(s) {
		j := i
		ind := i

		for _, matcher := range re.matchers {
			ind = matcher.Match(s, j)
			if ind >= len(s) {
				continue
			}

			if ind != -1 {
				j = ind
				continue
			}
			break
		}

		if ind != -1 {
			return true
		}

		i++
	}

	return false
}

func NewRegexp(s string) Regexp {
	var matchers []Matcher
	var nextLetter string
	var matcher Matcher
	var hasStartMarker, hasEndMarker bool

	hasStartMarker = strings.HasPrefix(s, "^")

	if hasStartMarker {
		s = strings.TrimPrefix(s, "^")
	}

	hasEndMarker = strings.HasSuffix(s, "$")
	if hasEndMarker {
		s = strings.TrimSuffix(s, "$")
	}

	i := 0

	for i < len(s) {
		letter := string(s[i])

		if i+1 == len(s) {
			nextLetter = ""
		} else {
			nextLetter = string(s[i+1])
		}

		switch letter {
		case `\`:
			matchers = append(matchers, Literal{literal: nextLetter})
			i++
		default:
			if letter == "." {
				matcher = AnyLiteral{}
			} else {
				matcher = Literal{literal: letter}
			}

			switch nextLetter {
			case "+":
				matchers = append(matchers, AtLeastOneLiteral{matcher: matcher})
				i++
			case "*":
				matchers = append(matchers, ZeroOrMoreLiterals{matcher: matcher})
				i++
			case "?":
				matchers = append(matchers, ZeroOrOneLiterals{matcher: matcher})
				i++
			default:
				matchers = append(matchers, matcher)
			}
		}
		i++
	}

	for i := 0; i < len(matchers)-1; i++ {

		switch matchers[i].(type) {
		case ZeroOrMoreLiterals:
			m := matchers[i].(ZeroOrMoreLiterals)
			m.next = &matchers[i+1]
			matchers[i] = m
		case AtLeastOneLiteral:
			m := matchers[i].(AtLeastOneLiteral)
			m.next = &matchers[i+1]
			matchers[i] = m
		}
	}

	if hasStartMarker {
		startsWith := []Matcher{StartsWith{matcher: matchers[0]}}
		matchers = append(startsWith, matchers[1:]...)
	}

	if hasEndMarker {
		var endsWithMatcher Matcher = EndsWith{matcher: matchers[len(matchers)-1]}
		matchers = append(matchers[:len(matchers)-1], endsWithMatcher)

		if len(matchers) > 1 {
			prevInd := len(matchers) - 2
			prevMatcher := matchers[prevInd]
			switch prev := prevMatcher.(type) {
			case StartsWith:
				switch m := prev.matcher.(type) {
				case AtLeastOneLiteral:
					m.next = &endsWithMatcher
					prev.matcher = m
					matchers[prevInd] = prev
				case ZeroOrMoreLiterals:
					m.next = &endsWithMatcher
					prev.matcher = m
					matchers[prevInd] = prev
				}
			case AtLeastOneLiteral:
				prev.next = &endsWithMatcher
				matchers[prevInd] = prev
			case ZeroOrMoreLiterals:
				prev.next = &endsWithMatcher
				matchers[prevInd] = prev
			}
		}
	}

	return Regexp{matchers: matchers}
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal("Error reading string")
	}
	parts := strings.Split(strings.TrimSuffix(input, "\n"), "|")

	if len(parts) != 2 {
		log.Fatal("Please enter a valid input")
	}

	re := NewRegexp(parts[0])
	fmt.Println(re.Match(parts[1]))
}
