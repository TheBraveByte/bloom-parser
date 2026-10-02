"""CLI entry point: `tableextract <extract|refine|normalize> ...`"""

import sys

from . import extract, normalize, refine

_COMMANDS = {"extract": extract, "refine": refine, "normalize": normalize}


def main():
    if len(sys.argv) < 2 or sys.argv[1] not in _COMMANDS:
        sys.exit("usage: tableextract <extract|refine|normalize> <args...>")
    _COMMANDS[sys.argv[1]].main(sys.argv[2:])


if __name__ == "__main__":
    main()
