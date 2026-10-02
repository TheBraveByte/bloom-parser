"""Grid extraction: image -> cell grid -> CSVs + flagged-cell crops."""

import csv
import os
import sys

import cv2

from . import grid as g
from . import ocr


def extract(path, crops_dir=None):
    """Process one image. Returns ((rows, flags, crops), mode) where crops is a
    list of (row, col, text, conf, label, filename) for low-confidence cells."""
    img = cv2.imread(path)
    if img is None:
        return None, "undecodable"
    scale = 2 if img.shape[1] < 1400 else 1
    if scale > 1:
        img = cv2.resize(img, None, fx=scale, fy=scale,
                         interpolation=cv2.INTER_CUBIC)
    gray = g.deskew(cv2.cvtColor(img, cv2.COLOR_BGR2GRAY))
    binary = cv2.adaptiveThreshold(gray, 255, cv2.ADAPTIVE_THRESH_GAUSSIAN_C,
                                   cv2.THRESH_BINARY_INV, 25, 10)
    row_ys, col_xs = g.grid_lines(binary)
    mode = "ruled"
    if not col_xs:
        mode = "borderless"
        h, w = binary.shape
        horiz = cv2.morphologyEx(
            binary, cv2.MORPH_OPEN,
            cv2.getStructuringElement(cv2.MORPH_RECT, (max(20, w // 25), 1)))
        row_ys = g.cluster(
            [y for y in range(h) if horiz.sum(axis=1)[y] > 255 * w * 0.15], tol=4)
        if len(row_ys) < 2:
            row_ys = [0, h]
        col_xs = g.gap_columns(binary, row_ys)
    if row_ys[0] > 0:
        row_ys = [0] + row_ys
    if row_ys[-1] < img.shape[0]:
        row_ys.append(img.shape[0])
    if col_xs[0] > 0:
        col_xs = [0] + col_xs
    if col_xs[-1] < img.shape[1]:
        col_xs.append(img.shape[1])

    ncols = len(col_xs) - 1
    cells = ocr.bin_words(ocr.ocr_words(gray), row_ys, col_xs)
    numeric = ocr.numeric_columns(cells, ncols)

    for r in sorted(cells):
        if not any(not ocr.is_noise(t) for t, _ in cells[r].values()):
            continue
        for c in range(ncols):
            text, conf = cells[r].get(c, ("", 0.0))
            if conf < ocr.CONF_RETRY and (text or numeric[c]):
                ntext, nconf = ocr.recell(gray, row_ys, col_xs, r, c, numeric[c])
                if ntext and nconf > conf:
                    cells[r][c] = (ntext, nconf)

    rows, flags, crops = [], [], []
    for r in sorted(cells):
        row, flag = [], []
        for c in range(ncols):
            text, conf = cells[r].get(c, ("", 0.0))
            if text and conf < ocr.CONF_FLAG:
                flag.append(f"c{c+1}:{int(conf)}")
            row.append(text)
        while row and ocr.is_noise(row[-1]):
            row.pop()
        if any(not ocr.is_noise(x) for x in row):
            rows.append(row)
            flags.append(",".join(flag))
            if crops_dir:
                label = next((x for x in row if not ocr.NUMERIC.match(x)), "")
                for c in range(len(row)):
                    text, conf = cells[r].get(c, ("", 0.0))
                    if text and conf < ocr.CONF_FLAG:
                        crop = gray[max(0, row_ys[r]):row_ys[r + 1],
                                    max(0, col_xs[c]):col_xs[c + 1]]
                        fn = (f"{os.path.splitext(os.path.basename(path))[0]}"
                              f"_r{len(rows)}_c{c+1}.png")
                        cv2.imwrite(os.path.join(crops_dir, fn), crop)
                        crops.append((len(rows), c + 1, text, int(conf),
                                      label, fn))
    return (rows, flags, crops), mode


def run(out_dir, paths):
    crops_dir = os.path.join(out_dir, "crops")
    os.makedirs(crops_dir, exist_ok=True)
    with open(os.path.join(out_dir, "pages.csv"), "w", newline="") as pf, \
         open(os.path.join(out_dir, "tables.csv"), "w", newline="") as tf, \
         open(os.path.join(out_dir, "crops.csv"), "w", newline="") as cf:
        pw, tw, cw = csv.writer(pf), csv.writer(tf), csv.writer(cf)
        pw.writerow(["file", "mode", "rows", "cells", "flagged_cells", "error"])
        tw.writerow(["file", "row", "flag", "c1..."])
        cw.writerow(["file", "row", "col", "text", "conf", "label", "crop"])
        for i, path in enumerate(paths, 1):
            name = os.path.basename(path)
            try:
                result, mode = extract(path, crops_dir)
            except Exception as e:  # noqa: BLE001 - report per file, keep going
                pw.writerow([name, "", 0, 0, 0, str(e)])
                continue
            if result is None:
                pw.writerow([name, "", 0, 0, 0, mode])
                continue
            rows, flags, crops = result
            for r, (row, flag) in enumerate(zip(rows, flags, strict=True), 1):
                tw.writerow([name, r, flag] + row)
            for r, c, text, conf, label, fn in crops:
                cw.writerow([name, r, c, text, conf, label, fn])
            pw.writerow([name, mode, len(rows), sum(len(r) for r in rows),
                         len(crops), ""])
            if i % 10 == 0 or i == len(paths):
                print(f"{i}/{len(paths)}", flush=True)
    print(f"wrote {out_dir}/pages.csv, tables.csv and crops.csv + crops/")


def main(argv):
    if len(argv) < 2:
        sys.exit("usage: extract <out_dir> <image>...")
    run(argv[0], argv[1:])
