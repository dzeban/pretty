import sys
from pprint import pformat

from fsm import FSM


class Pretty:
    output = sys.stdout

    def error(self, fsm: FSM):
        print("state: {}, symbol: '{}'".format(fsm.current_state, fsm.input_symbol))
        print("transitions:\n {}".format(pformat(fsm.state_transitions)))
        raise Exception("error while parsing input")

    def print(self, fsm: FSM):
        print(fsm.input_symbol, end="", sep="", file=self.output)

    def __init__(self):
        self.fsm = FSM("router", {})
        self.fsm.default_transition = self.error
        self.fsm.add_transition_list("\"'`", "router", self.print, "string")
        self.fsm.add_transition_list("\"'`", "string", self.print, "router")
        self.fsm.add_transition_any("string", action=self.print, next_state=None)

        self.fsm.add_transition_list(";\n", "router", self.print, "newline")
        self.fsm.add_transition_list(" \t\n", "newline", None)
        self.fsm.add_transition_list("\"'`", "newline", self.print, "string")
        self.fsm.add_transition_any("newline", action=None, next_state="router")

    def run(self, data):
        for c in data:
            self.fsm.process(c)
