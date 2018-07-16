import io

import pretty

def test_string():
    test = """
"asd
zz"
'qq'
`zxc
asd
qwe`"""

    p = pretty.Pretty()

    s = io.StringIO()
    p.output = s

    p.run(test)

    s.seek(0)
    output = s.read()

    assert output == test
