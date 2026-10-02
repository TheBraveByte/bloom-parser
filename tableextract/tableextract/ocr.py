"""Tesseract word OCR, cell binning, and per-cell re-OCR."""

import re

import cv2
import numpy as np
import pytesseract
from pytesseract import Output

CONF_RETRY = 65  # re-OCR cells below this mean word confidence
CONF_FLAG = 70   # still flag cells below this after re-OCR
DIGITS_CFG = "-c tessedit_char_whitelist=0123456789,.-()"

# Border glyphs tesseract reads from table rules. Parens are stripped only when
# touching a digit so "(Net)" in labels survives.
_EDGE = r'[\[\]{}|/\\_`~\'"“”‘’]'
_L_PAREN = re.compile(r'^\((?=[\d.,])')
_R_PAREN = re.compile(r'(?<=[\d.,])\)$')
NUMERIC = re.compile(r'^[(\-]?[\d,.\s]+[)]?$')


def clean_cell(text):
    text = re.sub(rf'^{_EDGE}+|{_EDGE}+$', '', text.strip())
    text = _L_PAREN.sub('', _R_PAREN.sub('', text))
    return text.strip()


def is_noise(text):
    return not re.search(r'[A-Za-z0-9]', text)


def ocr_words(img):
    d = pytesseract.image_to_data(img, config="--psm 6", output_type=Output.DICT)
    words = []
    for i, text in enumerate(d["text"]):
        text = text.strip()
        conf = float(d["conf"][i])
        if text and conf > 0:
            words.append((d["left"][i], d["top"][i], d["width"][i],
                          d["height"][i], text, conf))
    return words


def bin_words(words, row_ys, col_xs):
    """Assign words to (row, col) cells by centre point. Returns
    {row: {col: (text, conf)}} with words joined in reading order."""
    def band(v, bounds):
        for i in range(len(bounds) - 1):
            if bounds[i] <= v < bounds[i + 1]:
                return i
        return None

    cells = {}
    for x, y, w, h, text, conf in words:
        r, c = band(y + h / 2, row_ys), band(x + w / 2, col_xs)
        if r is not None and c is not None:
            cells.setdefault(r, {}).setdefault(c, []).append((x, y, text, conf))
    out = {}
    for r, cols in cells.items():
        out[r] = {}
        for c, ws in cols.items():
            ws.sort(key=lambda t: (t[1] // 8, t[0]))
            out[r][c] = (clean_cell(" ".join(t[2] for t in ws)),
                         float(np.mean([t[3] for t in ws])))
    return out


def recell(gray, row_ys, col_xs, r, c, numeric):
    """Crop one cell, upscale, and re-OCR it as a single line."""
    pad = 4
    x0 = max(0, col_xs[c] + pad)
    x1 = min(gray.shape[1], col_xs[c + 1] - pad)
    y0 = max(0, row_ys[r] + pad)
    y1 = min(gray.shape[0], row_ys[r + 1] - pad)
    if x1 - x0 < 8 or y1 - y0 < 8:
        return "", 0.0
    crop = gray[y0:y1, x0:x1]
    crop = cv2.resize(crop, None, fx=3, fy=3, interpolation=cv2.INTER_CUBIC)
    cfg = "--psm 7 " + (DIGITS_CFG if numeric else "")
    d = pytesseract.image_to_data(crop, config=cfg, output_type=Output.DICT)
    words = [(t.strip(), float(d["conf"][i]))
             for i, t in enumerate(d["text"])
             if t.strip() and float(d["conf"][i]) > 0]
    if not words:
        return "", 0.0
    return (clean_cell(" ".join(w for w, _ in words)),
            float(np.mean([cf for _, cf in words])))


def numeric_columns(grid, ncols):
    """A column is numeric if over half its populated cells look numeric."""
    counts, nums = [0] * ncols, [0] * ncols
    for cols in grid.values():
        for c, (text, _) in cols.items():
            if text:
                counts[c] += 1
                if NUMERIC.match(text):
                    nums[c] += 1
    return [counts[c] >= 3 and nums[c] > counts[c] * 0.6 for c in range(ncols)]
