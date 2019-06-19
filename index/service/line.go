package service

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
)

type liner struct {
	scanner    *bufio.Scanner
	lineNumber int
	lines      map[int]string
	done       bool
}

func newLiner(content []byte) *liner {
	return &liner{
		scanner:    bufio.NewScanner(bytes.NewReader(content)),
		lineNumber: 0,
		lines:      make(map[int]string),
	}
}

var errLineNotExist = errors.New("")

// we only need to keep #[n-2, n+1] line
func (l *liner) advanceToLineNumber(n int) error {
	if l.done {
		return nil
	}

	for l.lineNumber < n-2 {
		if !l.scanner.Scan() {
			return errors.New(fmt.Sprintf("unexpect error when scan line:%d", l.lineNumber+1))
		}
		l.lineNumber += 1
	}

	for ln := n - 2; ln <= n; ln++ {
		if ln <= l.lineNumber {
			continue
		}
		if _, ok := l.lines[ln]; ok {
			continue
		}
		if !l.scanner.Scan() {
			return errors.New(fmt.Sprintf("unexpect error when scan line:%d", l.lineNumber+1))
		}
		l.lineNumber += 1
		l.lines[l.lineNumber] = l.scanner.Text()
	}

	// #n+1 line, it may not exist
	if !l.scanner.Scan() {
		err := l.scanner.Err()
		if err == nil {
			l.done = true
			err = nil
		}
		return err
	}
	l.lineNumber += 1
	l.lines[l.lineNumber] = l.scanner.Text()
	return nil
}

func (l *liner) getLine(n int) (string, error) {
	if n <= 0 {
		return "", errors.New("unexpected line number <= 0")
	}

	l.advanceToLineNumber(n)
	if n > l.lineNumber {
		return "", errLineNotExist
	}

	return l.lines[n], nil
}
