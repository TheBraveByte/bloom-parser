"""Normalize extracted grid rows into canonical facts.

Each numeric cell becomes one row: file, row, section, line_item, period,
value, raw, flag. Period comes from the sheet's date header (Standalone /
Consolidated group x year); sheets without a detectable header fall back to
positional period labels (col4, col5, ...).
"""

import csv
import os
import re
import sys

from .ocr import NUMERIC, is_noise

_MONTHS = ("jan|feb|mar|apr|may|jun|jul|aug|sep|oct|nov|dec")
_DMY = re.compile(rf'\b(\d{{1,2}})[-./](\d{{1,2}}|{_MONTHS})[a-z]*[-./ ]'
                  r'((?:19|20)\d{2})', re.I)
_MDY = re.compile(rf'\b({_MONTHS})[a-z]*\s+(\d{{1,2}})[a-z]*[,\s]+'
                  r'((?:19|20)\d{2})', re.I)
_MY = re.compile(rf'\b({_MONTHS})[a-z]*\s+((?:19|20)\d{{2}})', re.I)
_YEAR = re.compile(r'\b(?:19|20)\d{2}\b')
_GROUP = re.compile(r'(?i)standalone|consolidated')
_SECTION = re.compile(
    r'(?i)^(assets|equity|liabilities|non-current|current|financial'
    r'|equity and liabilities|total|capital|profit|loss|income'
    r'|revenue|expenses|cash flow|notes|b\s|a\s|[a-z]\))')
_MNUM = {"jan": 1, "feb": 2, "mar": 3, "apr": 4, "may": 5, "jun": 6,
         "jul": 7, "aug": 8, "sep": 9, "oct": 10, "nov": 11, "dec": 12}


def parse_period(cell):
    """Extract a period label from a header cell: ISO date > year-month > year.
    Returns '' if the cell has no date/year."""
    m = _DMY.search(cell)
    if m:
        d, mo, y = m.groups()
        mo = _MNUM.get(mo[:3].lower(), int(mo) if mo.isdigit() else 0)
        return f"{y}-{mo:02d}-{int(d):02d}" if mo else y
    m = _MDY.search(cell)
    if m:
        mo, d, y = m.groups()
        return f"{y}-{_MNUM[mo[:3].lower()]:02d}-{int(d):02d}"
    m = _MY.search(cell)
    if m:
        mo, y = m.groups()
        return f"{y}-{_MNUM[mo[:3].lower()]:02d}"
    m = _YEAR.search(cell)
    return m.group(0) if m else ""


def normalize_value(raw):
    """Parse an OCR'd amount into a canonical numeric string. Indian-format
    thousands use ','; a '.' before a 3-digit group is treated as thousands
    (OCR comma/period confusion), otherwise it's a decimal point. Parentheses
    mean negative. Returns '' when not parseable."""
    s = raw.strip().replace(" ", "")
    if not s:
        return ""
    neg = s.startswith("-") or (s.startswith("(") and s.endswith(")"))
    s = s.strip("()").lstrip("-")
    if not re.fullmatch(r"[\d,.]+", s):
        return ""
    if "." in s and "," in s:
        if s.rsplit(".", 1)[-1].isdigit() and len(s.rsplit(".", 1)[-1]) <= 2 \
                and s.rindex(".") > s.rindex(","):
            s = s.replace(",", "")
        else:
            s = s.replace(",", "").replace(".", "")
    elif "," in s:
        s = s.replace(",", "")
    elif s.count(".") > 1 or (s.count(".") == 1 and len(s.rsplit(".", 1)[1]) == 3):
        s = s.replace(".", "")
    if not re.fullmatch(r"\d+(\.\d+)?", s):
        return ""
    return ("-" if neg else "") + s


def is_section_label(text):
    return bool(_SECTION.match(text.strip()))


def parse_periods(cell):
    """All period labels in a cell, left to right (merged header cells can
    hold several, e.g. 'March 2018 March 2017')."""
    out = []
    rest = cell
    while True:
        p = parse_period(rest)
        if not p:
            return out
        out.append(p)
        for rx in (_DMY, _MDY, _MY, _YEAR):
            m = rx.search(rest)
            if m:
                rest = rest[m.end():]
                break


def detect_header(rows):
    """Find the period header row: the row (within the first 12) with >=2
    periods across its cells. Returns (row_idx, [periods in column order])."""
    best, best_periods = -1, []
    for i, r in enumerate(rows[:12]):
        periods = [p for cell in r for p in parse_periods(cell)]
        if len(periods) >= 2 and len(periods) > len(best_periods):
            best, best_periods = i, periods
    return best, best_periods


def detect_groups(rows, upto, val_cols):
    """Map each value column to a Standalone/Consolidated group using group
    labels in the rows above the period header."""
    groups = {}
    for r in rows[:upto + 1]:
        seen = ""
        for ci, cell in enumerate(r):
            g = _GROUP.search(cell)
            if g:
                seen = g.group(0).lower()
            if ci in val_cols and seen:
                groups[ci] = seen
    return groups


def numeric_value_columns(rows, hdr):
    """Columns whose body cells are mostly numeric = the value columns."""
    counts, nums = {}, {}
    for _, _, r in rows[hdr + 1:]:
        for ci, cell in enumerate(r):
            if not cell or is_noise(cell):
                continue
            counts[ci] = counts.get(ci, 0) + 1
            if NUMERIC.match(cell):
                nums[ci] = nums.get(ci, 0) + 1
    return [c for c in sorted(counts)
            if nums.get(c, 0) >= 3 and nums[c] > counts[c] * 0.6]


def facts_for_sheet(file, rows):
    """rows: list of (rowno, flag, [cells]). Yields normalized fact dicts."""
    cells = [r[2] for r in rows]
    hdr, periods = detect_header(cells)
    val_cols = numeric_value_columns(rows, hdr)
    groups = detect_groups(cells, hdr, set(val_cols))
    queue = list(periods)
    period = {}
    for c in val_cols:
        p = queue.pop(0) if queue else f"col{c + 1}"
        period[c] = (groups.get(c, "") + ":" if groups.get(c) else "") + p
    section, pending_label = "", None
    last_facts = []
    for idx, (rowno, flag, r) in enumerate(rows):
        if idx <= hdr:
            continue
        line_item = next((c for c in r if c and not NUMERIC.match(c)
                          and not is_noise(c)), "")
        numbers = [(ci, c) for ci, c in enumerate(r)
                   if c and NUMERIC.match(c) and not is_noise(c)]
        if not numbers:
            if not line_item:
                continue
            nxt = next((x for x in rows if x[0] > rowno), None)
            nxt_vals = [c for c in nxt[2] if NUMERIC.match(c)] if nxt else []
            nxt_label = next((c for c in nxt[2] if c
                              and not NUMERIC.match(c)), "") if nxt else ""
            if nxt_vals and not nxt_label:
                pending_label = line_item
            elif last_facts and len(line_item.split()) <= 3 \
                    and not is_section_label(line_item):
                for f in last_facts:
                    f["line_item"] += " " + line_item
            elif is_section_label(line_item):
                section = line_item
            continue
        label = line_item or pending_label or ""
        pending_label = None
        last_facts = []
        for ci, raw in numbers:
            if ci in period:
                p = period[ci]
            elif val_cols and min(abs(c - ci) for c in val_cols) == 1:
                p = period[min(val_cols, key=lambda c: abs(c - ci))]
            else:
                p = f"col{ci + 1}"
            fact = {"file": file, "row": rowno, "section": section,
                    "line_item": label, "period": p,
                    "value": normalize_value(raw), "raw": raw, "flag": flag}
            last_facts.append(fact)
            yield fact


def run(out_dir):
    src = os.path.join(out_dir, "tables_refined.csv")
    if not os.path.exists(src):
        src = os.path.join(out_dir, "tables.csv")
    rows = list(csv.reader(open(src, newline="")))
    sheets = {}
    for r in rows[1:]:
        sheets.setdefault(r[0], []).append((int(r[1]), r[2], r[3:]))
    out = os.path.join(out_dir, "normalized.csv")
    facts = 0
    with open(out, "w", newline="") as f:
        w = csv.DictWriter(f, fieldnames=["file", "row", "section",
                                          "line_item", "period",
                                          "value", "raw", "flag"])
        w.writeheader()
        for file, srows in sheets.items():
            srows.sort(key=lambda x: x[0])
            for fact in facts_for_sheet(file, srows):
                w.writerow(fact)
                facts += 1
    print(f"{facts} facts across {len(sheets)} sheets -> {out}")


def main(argv):
    if len(argv) < 1:
        sys.exit("usage: normalize <out_dir>")
    run(argv[0])
