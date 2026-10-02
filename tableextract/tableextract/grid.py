"""OpenCV helpers: skew correction and ruled-grid line detection."""

import cv2
import numpy as np


def cluster(positions, tol):
    """Merge 1-D positions within tol of each other into their means."""
    out = []
    for p in sorted(positions):
        if out and p - out[-1][-1] <= tol:
            out[-1].append(p)
        else:
            out.append([p])
    return [int(np.mean(g)) for g in out]


def deskew(gray):
    """Estimate page skew from the minAreaRect over ink pixels and rotate."""
    inv = cv2.threshold(gray, 0, 255, cv2.THRESH_BINARY_INV + cv2.THRESH_OTSU)[1]
    pts = cv2.findNonZero(inv)
    if pts is None:
        return gray
    angle = cv2.minAreaRect(pts)[2]
    if angle > 45:
        angle -= 90
    if abs(angle) < 0.1 or abs(angle) > 5:
        return gray
    h, w = gray.shape
    m = cv2.getRotationMatrix2D((w / 2, h / 2), angle, 1.0)
    return cv2.warpAffine(gray, m, (w, h), flags=cv2.INTER_CUBIC,
                          borderMode=cv2.BORDER_CONSTANT, borderValue=255)


def grid_lines(binary):
    """Return (row_ys, col_xs) grid boundaries, or ([], []) without a usable
    ruled grid. Columns come from vertical-line segment contours since
    financial tables often rule only part of a column's height."""
    h, w = binary.shape
    horiz = cv2.morphologyEx(
        binary, cv2.MORPH_OPEN,
        cv2.getStructuringElement(cv2.MORPH_RECT, (max(20, w // 25), 1)))
    vert = cv2.morphologyEx(
        binary, cv2.MORPH_OPEN,
        cv2.getStructuringElement(cv2.MORPH_RECT, (1, max(15, h // 60))))
    row_ys = cluster(np.where(horiz.sum(axis=1) > 255 * w * 0.15)[0], tol=4)
    segs = []
    for c in cv2.findContours(vert, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)[0]:
        x, y, cw, ch = cv2.boundingRect(c)
        if ch >= h * 0.015 and cw <= w * 0.02:
            segs.append(x + cw // 2)
    col_xs = cluster(segs, tol=6)
    if len(row_ys) < 2 or len(col_xs) < 2:
        return [], []
    return row_ys, col_xs


def gap_columns(binary, row_ys):
    """Fallback for borderless tables: column boundaries from vertical
    whitespace gaps in the word-pixel projection inside the table body."""
    top, bot = row_ys[0], row_ys[-1]
    filled = binary[top:bot, :].sum(axis=0) > 0
    gaps, run = [], 0
    for x, f in enumerate(filled):
        if f:
            if run >= 8:
                gaps.append(x - run // 2)
            run = 0
        else:
            run += 1
    return cluster(gaps, tol=6)
