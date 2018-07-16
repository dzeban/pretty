import io

from pretty import Pretty


def test_string():
    test = """
"asd
zz"
'qq'
`zxc
asd
qwe`"""

    check_input(test, test)


def test_word():
    tests = [("a  b   c", "a b c"), ("a	b	c", "a b c"), ("123	45  6", "123 45 6")]

    for test in tests:
        check_input(test[0], test[1])


def test_indent():
    tests = [("fn some(param) { hi }", "fn some(param) {\n    hi \n}")]

    for test in tests:
        check_input(test[0], test[1])


def check_input(inp: str, expected: str):
    p = Pretty()

    # Output pretty printed text to string instead of stdout
    # to compare with expected result
    s = io.StringIO()
    p.output = s

    p.run(inp)

    # Read pretty printed text
    # Seek is needed here because it is a stream over string
    s.seek(0)
    output = s.read()

    assert output == expected
