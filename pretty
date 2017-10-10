#!/usr/bin/env python

import fileinput
import string
from functools import partial

lparens = "{(["
rparens = "])}"
parens = lparens + rparens

sibling_delims = ",;"
kv_delims = ":"

rprint = partial(print, end='', sep='')

def process(line):
    indent = ''
    shift = '    '

    i = 0
    while i < len(line):
        if line[i] in lparens:
            indent += shift
            rprint(line[i] + "\n" + indent)
        elif line[i] in rparens:
            indent = indent[:(len(indent) - len(shift))]
            rprint("\n" + indent + line[i])
        elif line[i] in sibling_delims:
            rprint(line[i] + "\n" + indent)
        elif line[i] in string.whitespace:
            pass
        elif line[i] in kv_delims:
            rprint(line[i] + " ")
        elif line[i] == '"':
            rprint(line[i])
            i += 1
            while line[i] != '"':
                rprint(line[i])
                i += 1
            rprint(line[i])
        else:
            rprint(line[i])

        i += 1

for line in fileinput.input():
    process(line)
    print()
