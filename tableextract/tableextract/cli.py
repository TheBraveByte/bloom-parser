"""CLI entry point: `tableextract <extract|refine|normalize|pipeline> ...`"""

import sys

from . import extract, normalize, refine

_COMMANDS = {"extract": extract, "refine": refine, "normalize": normalize}


def _pipeline(argv):
    """extract -> refine (optional) -> normalize in a single process.

    The server used to spawn three Python processes per file; each pays the
    interpreter + cv2 import cost again. Fusing the stages keeps the same
    outputs while cutting most of that startup time. refine stays
    best-effort: any failure or an exhausted --refine-budget falls back to the
    un-refined tables rather than losing the file."""
    paths = [a for a in argv if not a.startswith("--")]
    if len(paths) < 2:
        sys.exit("usage: pipeline <out_dir> <image>... "
                 "[--refine] [--refine-budget=SECONDS]")
    out_dir, images = paths[0], paths[1:]
    budget = next((float(a.split("=", 1)[1]) for a in argv
                   if a.startswith("--refine-budget=")), None)
    extract.run(out_dir, images)
    if "--refine" in argv:
        try:
            refine.run(out_dir, budget=budget)
        except Exception as e:  # noqa: BLE001 - refine is best-effort
            print(f"refine failed, using un-refined tables: {e}",
                  file=sys.stderr)
    normalize.run(out_dir)


def main():
    if len(sys.argv) < 2:
        sys.exit("usage: tableextract <extract|refine|normalize|pipeline>"
                 " <args...>")
    if sys.argv[1] == "pipeline":
        _pipeline(sys.argv[2:])
    elif sys.argv[1] in _COMMANDS:
        _COMMANDS[sys.argv[1]].main(sys.argv[2:])
    else:
        sys.exit(f"unknown command: {sys.argv[1]}")


if __name__ == "__main__":
    main()
