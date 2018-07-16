#!/usr/bin/env python3

import fileinput

from pretty import Pretty

if __name__ == '__main__':
    p = Pretty()

    for line in fileinput.input():
        p.run(line)
