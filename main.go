package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
)

type StateId string

const (
	StateUnknown         StateId = "Unknown"
	StateMain            StateId = "Main"
	StateAlphanum        StateId = "Alphanum"
	StateLineEnd         StateId = "LineEnd"
	StateWhitespace      StateId = "Whitespace"
	StateWhitespace1     StateId = "Whitespace1"
	StateBlockOpen       StateId = "BlockOpen"
	StateBlockClose      StateId = "BlockClose"
	StateBlockEnd        StateId = "BlockEnd"
	StateBlockEndNewline StateId = "BlockEndNewline"
	StateQuoteOpen       StateId = "QuoteOpen"
	StateQuoteClose      StateId = "QuoteClose"
	StateString          StateId = "String"
)

type Event string

const (
	EventUnknown       Event = "EventUnknown"
	EventStop          Event = "EventStop"
	EventAny           Event = "EventAny"
	EventAlphanum      Event = "EventAlphanum"
	EventNonAlphanum   Event = "EventNonAlphanum"
	EventWhitespace    Event = "EventWhitespace"
	EventNonWhitespace Event = "EventNonWhitespace"
	EventLineSeparator Event = "EventLineSeparator"
	EventBlockOpen     Event = "EventBlockOpen"
	EventBlockClose    Event = "EventBlockClose"
	EventQuote         Event = "EventQuote"
)

type State struct {
	Action      func(runner *Runner) (Event, error)
	Transitions map[Event]StateId
}

type StateMachine map[StateId]State

type Runner struct {
	stateMachine StateMachine
	input        *bufio.Reader
	indent       string
	indentLevel  int
	cur          byte
	quote        byte
	debug        bool
}

func NewRunner(sm StateMachine, input io.Reader) (*Runner, error) {
	smr := Runner{
		stateMachine: sm,
		input:        bufio.NewReader(input),
		indentLevel:  0,
		indent:       "    ", // TODO: make configurable
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
				break
			}
			return fmt.Errorf("halt on error: %w", err)
		}

		if event == EventUnknown {
			return fmt.Errorf("halt on unknown event in state %v, current input is %q", next, r.cur)
		}

		if event == EventStop {
			break
		}

		curState := next // only for logging

		next = state.Transitions[event]
		state = r.stateMachine[next]

		if r.debug {
			fmt.Fprintf(os.Stderr, "%v(%q): %v -> %v\n", curState, r.cur, event, next)
		}
	}
	r.Newline()
	return nil
}

func (r *Runner) Print() {
	os.Stdout.Write([]byte{r.cur})
}

func (r *Runner) Newline() {
	os.Stdout.WriteString("\n")
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
	RegexAlphanum          = regexp.MustCompile("[^{(\\[\\])} \t\n,;'\"`]")
	RegexLine              = regexp.MustCompile(`[,;\n]`)
	RegexWhitespace        = regexp.MustCompile(`[ \t]`)
	RegexWhitespaceNewline = regexp.MustCompile(`[ \t\n]`)
	RegexBlockOpen         = regexp.MustCompile(`[{(\[]`)
	RegexBlockClose        = regexp.MustCompile(`[})\]]`)
	RegexQuote             = regexp.MustCompile("['\"`]")
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

	if RegexBlockOpen.Match(cur) {
		return EventBlockOpen, nil
	}

	if RegexBlockClose.Match(cur) {
		return EventBlockClose, nil
	}

	if RegexQuote.Match(cur) {
		return EventQuote, nil
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

	runner.Newline()
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

func BlockOpen(runner *Runner) (Event, error) {
	runner.Print()
	runner.Newline()
	runner.IncreaseIndent()
	runner.Indent()
	err := runner.Advance()
	if err != nil {
		return EventUnknown, fmt.Errorf("BlockOpen: %w", err)
	}

	if RegexWhitespaceNewline.Match([]byte{runner.cur}) {
		return EventWhitespace, nil
	}

	return EventNonWhitespace, nil
}

func BlockClose(runner *Runner) (Event, error) {
	runner.Newline()
	runner.DecreaseIndent()
	runner.Indent()
	runner.Print()
	runner.Advance()
	if RegexLine.Match([]byte{runner.cur}) {
		return EventLineSeparator, nil
	}
	return EventAny, nil
}

func BlockEnd(runner *Runner) (Event, error) {
	if !RegexWhitespaceNewline.Match([]byte{runner.cur}) {
		runner.Print()
	}

	err := runner.Advance()
	if err != nil {
		return EventUnknown, fmt.Errorf("BlockEnd: %w", err)
	}

	if RegexLine.Match([]byte{runner.cur}) {
		return EventLineSeparator, nil
	}

	if RegexWhitespaceNewline.Match([]byte{runner.cur}) {
		return EventWhitespace, nil
	}

	if RegexBlockClose.Match([]byte{runner.cur}) {
		return EventBlockClose, nil
	}

	return EventAny, nil
}

func BlockEndNewline(runner *Runner) (Event, error) {
	runner.Newline()
	runner.Indent()
	return EventAny, nil
}

func QuoteOpen(runner *Runner) (Event, error) {
	// Save open quote to match closing
	runner.quote = runner.cur

	runner.Print()
	runner.Advance()

	if runner.cur == runner.quote {
		return EventQuote, nil
	}

	return EventAny, nil
}

func QuoteClose(runner *Runner) (Event, error) {
	runner.Print()
	runner.Advance()

	runner.quote = 0

	return EventAny, nil
}

func String(runner *Runner) (Event, error) {
	runner.Print()
	runner.Advance()

	if runner.cur == runner.quote {
		return EventQuote, nil
	}

	return EventAny, nil
}

var PrettyStateMachine = StateMachine{
	StateMain: {
		Main,
		map[Event]StateId{
			EventAlphanum:      StateAlphanum,
			EventLineSeparator: StateLineEnd,
			EventWhitespace:    StateWhitespace1,
			EventBlockOpen:     StateBlockOpen,
			EventBlockClose:    StateBlockClose,
			EventQuote:         StateQuoteOpen,
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
	StateBlockOpen: {
		BlockOpen,
		map[Event]StateId{
			EventWhitespace:    StateWhitespace,
			EventNonWhitespace: StateMain,
		},
	},
	StateBlockClose: {
		BlockClose,
		map[Event]StateId{
			EventLineSeparator: StateBlockEnd,
			EventAny:           StateMain,
		},
	},
	StateBlockEnd: {
		BlockEnd,
		map[Event]StateId{
			EventLineSeparator: StateBlockEnd,
			EventWhitespace:    StateBlockEnd,
			EventBlockClose:    StateBlockClose,
			EventAny:           StateBlockEndNewline,
		},
	},
	StateBlockEndNewline: {
		BlockEndNewline,
		map[Event]StateId{
			EventAny: StateMain,
		},
	},
	StateQuoteOpen: {
		QuoteOpen,
		map[Event]StateId{
			EventQuote: StateQuoteClose,
			EventAny:   StateString,
		},
	},
	StateQuoteClose: {
		QuoteClose,
		map[Event]StateId{
			EventAny: StateMain,
		},
	},
	StateString: {
		String,
		map[Event]StateId{
			EventQuote: StateQuoteClose,
			EventAny:   StateString,
		},
	},
}

func main() {
	debug := flag.Bool("debug", false, "Enable debug printing")
	flag.Parse()

	runner, err := NewRunner(PrettyStateMachine, os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	runner.debug = *debug

	err = runner.Run(StateMain)
	if err != nil {
		log.Fatal(err)
	}
}
