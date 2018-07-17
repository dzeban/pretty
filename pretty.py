import sys
from pprint import pformat

from fsm import FSM


class Pretty:
    output = sys.stdout
    indent = ""
    indent_level = 0
    indent_size = 4  # spaces

    def error(self, fsm: FSM):
        print("state: {}, symbol: '{}'".format(fsm.current_state, fsm.input_symbol))
        print("transitions:\n {}".format(pformat(fsm.state_transitions)))
        raise Exception("error while parsing input")

    def print(self, fsm: FSM):
        print(fsm.input_symbol, end="", sep="", file=self.output)

    def print_space(self, fsm: FSM):
        print(" ", end="", sep="", file=self.output)

    def print_indent(self, fsm: FSM):
        print(fsm.input_symbol, end="", sep="", file=self.output)
        print("\n", end="", sep="", file=self.output)
        self.indent_level += 1
        self.indent = " " * self.indent_level * self.indent_size
        print(self.indent, end="", sep="", file=self.output)

    def print_unindent(self, fsm: FSM):
        print("\n", end="", sep="", file=self.output)
        self.indent_level -= 1
        self.indent = " " * self.indent_level * self.indent_size
        print(self.indent + fsm.input_symbol, end="", sep="", file=self.output)

    def print_newline(self, fsm: FSM):
        print(fsm.input_symbol, end="", sep="", file=self.output)
        if fsm.input_symbol != "\n":
            print("\n", end="", sep="", file=self.output)

        print(self.indent, end="", sep="", file=self.output)

    def __init__(self):
        self.fsm = FSM("router", {})
        self.fsm.default_transition = self.error

        self.fsm.add_transition_any("router", action=self.print, next_state=None)

        self.fsm.add_transition_list("\"'`", "router", self.print, "string")
        self.fsm.add_transition_list("\"'`", "string", self.print, "router")
        self.fsm.add_transition_any("string", action=self.print, next_state=None)

        self.fsm.add_transition_list(",;\n", "router", self.print_newline, "newline")
        self.fsm.add_transition_list("{[", "router", self.print_indent, "newline")
        self.fsm.add_transition_list("}]", "router", self.print_unindent, "newline")

        self.fsm.add_transition_list(" \t\n", "newline", None)
        self.fsm.add_transition_list("\"'`", "newline", self.print, "string")
        self.fsm.add_transition_list("{[", "newline", self.print_indent, "newline")
        self.fsm.add_transition_list("}]", "newline", self.print_unindent, "newline")
        self.fsm.add_transition_list(",;", "newline", self.print_newline, "newline")
        self.fsm.add_transition_any("newline", action=self.print, next_state="router")

        self.fsm.add_transition_list(" \t", "router", self.print_space, "word")
        self.fsm.add_transition_list(" \t", "word", action=None, next_state=None)
        self.fsm.add_transition_list(" \t", "word", action=None, next_state=None)
        self.fsm.add_transition_list("{[", "word", self.print_indent, "newline")
        self.fsm.add_transition_list("}]", "word", self.print_unindent, "newline")
        self.fsm.add_transition_list("\"'`", "word", self.print, "string")
        self.fsm.add_transition_any("word", action=self.print, next_state="router")

    def run(self, data):
        for c in data:
            self.fsm.process(c)
