package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
)

type StateId string

const (
	StateUnknown     StateId = "Unknown"
	StateMain        StateId = "Main"
	StateAlphanum    StateId = "Alphanum"
	StateLineEnd     StateId = "LineEnd"
	StateWhitespace  StateId = "Whitespace"
	StateWhitespace1 StateId = "Whitespace1"
)

type Event int

const (
	EventUnknown Event = iota
	EventStop
	EventAny
	EventAlphanum
	EventNonAlphanum
	EventWhitespace
	EventNonWhitespace
	EventLineSeparator
)

type State struct {
	Action      func(runner *Runner) (Event, error)
	Transitions map[Event]StateId
}

type StateMachine map[StateId]State

type Runner struct {
	stateMachine StateMachine
	input        *bufio.Reader
	cur          byte
	indentLevel  int
	indent       string
}

func NewRunner(sm StateMachine, input io.Reader) (*Runner, error) {
	smr := Runner{
		stateMachine: sm,
		input:        bufio.NewReader(input),
		indentLevel:  0,
		indent:       "  ", // TODO: make configurable
	}

	// Initialize cur
	err := smr.Advance()
	if err != nil {
		return nil, fmt.Errorf("failed to read first byte: %w", err)
	}

	return &smr, nil
}

func (r *Runner) Run(next StateId) error {
	state := r.stateMachine[next]

	var (
		event Event
		err   error
	)

	for {
		event, err = state.Action(r)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("halt on error: %w", err)
		}

		if event == EventUnknown {
			return fmt.Errorf("halt on unknown event in state %v, current input is %q", next, r.cur)
		}

		if event == EventStop {
			break
		}
		next = state.Transitions[event]
		state = r.stateMachine[next]
	}
	return nil
}

func (r *Runner) Print() {
	os.Stdout.Write([]byte{r.cur})
}

func (r *Runner) Advance() error {
	var err error
	r.cur, err = r.input.ReadByte()
	if err != nil {
		return fmt.Errorf("reading failed: %w", err)
	}
	return nil
}

func (r *Runner) Indent() {
	for i := 0; i < r.indentLevel; i++ {
		fmt.Print(r.indent)
	}
}

func (r *Runner) IncreaseIndent() {
	r.indentLevel++
}

func (r *Runner) DecreaseIndent() {
	r.indentLevel--
}

var (
	RegexAlphanum   = regexp.MustCompile("[a-zA-Z0-9]")
	RegexLine       = regexp.MustCompile("[,;\n]")
	RegexWhitespace = regexp.MustCompile("[ \t]")
)

func Main(runner *Runner) (Event, error) {
	cur := []byte{runner.cur}
	if RegexAlphanum.Match(cur) {
		return EventAlphanum, nil
	}

	if RegexLine.Match(cur) {
		return EventLineSeparator, nil
	}

	if RegexWhitespace.Match(cur) {
		return EventWhitespace, nil
	}

	return EventUnknown, nil
}

func Alphanum(runner *Runner) (Event, error) {
	runner.Print()

	err := runner.Advance()
	if err != nil {
		return EventUnknown, fmt.Errorf("Alphanum: %w", err)
	}

	if RegexAlphanum.Match([]byte{runner.cur}) {
		return EventAlphanum, nil
	}
	return EventNonAlphanum, nil
}

func Line(runner *Runner) (Event, error) {
	if runner.cur != '\n' {
		runner.Print()
	}

	runner.Print()
	runner.Indent()
	err := runner.Advance()
	if err != nil {
		return EventUnknown, fmt.Errorf("Line: %w", err)
	}

	if RegexWhitespace.Match([]byte{runner.cur}) || runner.cur == '\n' {
		return EventWhitespace, nil
	}
	return EventNonWhitespace, nil
}

func Whitespace1(runner *Runner) (Event, error) {
	runner.Print()

	err := runner.Advance()
	if err != nil {
		return EventUnknown, fmt.Errorf("Alphanum: %w", err)
	}

	if RegexWhitespace.Match([]byte{runner.cur}) {
		return EventWhitespace, nil
	}
	return EventNonWhitespace, nil
}

func Whitespace(runner *Runner) (Event, error) {
	err := runner.Advance()
	if err != nil {
		return EventUnknown, fmt.Errorf("Line: %w", err)
	}

	if RegexWhitespace.Match([]byte{runner.cur}) {
		return EventWhitespace, nil
	}
	return EventNonWhitespace, nil
}

var PrettyStateMachine = StateMachine{
	StateMain: {
		Main,
		map[Event]StateId{
			EventAlphanum:      StateAlphanum,
			EventLineSeparator: StateLineEnd,
			EventWhitespace:    StateWhitespace1,
		},
	},
	StateAlphanum: {
		Alphanum,
		map[Event]StateId{
			EventAlphanum:    StateAlphanum,
			EventNonAlphanum: StateMain,
		},
	},
	StateLineEnd: {
		Line,
		map[Event]StateId{
			EventWhitespace:    StateWhitespace,
			EventNonWhitespace: StateMain,
		},
	},
	StateWhitespace1: {
		Whitespace1,
		map[Event]StateId{
			EventWhitespace:    StateWhitespace,
			EventNonWhitespace: StateMain,
		},
	},
	StateWhitespace: {
		Whitespace,
		map[Event]StateId{
			EventWhitespace:    StateWhitespace,
			EventNonWhitespace: StateMain,
		},
	},
}

func main() {
	runner, err := NewRunner(PrettyStateMachine, os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	err = runner.Run(StateMain)
	if err != nil {
		log.Fatal(err)
	}
}
