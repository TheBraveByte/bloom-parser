"""Vision-LLM refinement of low-confidence cells (OpenAI-compatible API).

Every flagged cell goes to the fast model; answers failing sanity guards (too
long for the crop, column-type mismatch, or wildly different from a moderately
confident OCR read) get a second opinion from the bigger model.
"""

import base64
import csv
import difflib
import json
import os
import re
import sys
import time
import urllib.request
from concurrent.futures import ThreadPoolExecutor

import cv2

from .ocr import NUMERIC

URL = "https://integrate.api.nvidia.com/v1/chat/completions"
MODEL_FAST = "meta/llama-3.2-11b-vision-instruct"
MODEL_SLOW = "meta/llama-3.2-90b-vision-instruct"
WORKERS = 4
TIMEOUT = 25
RETRIES = 2
CHAR_PX = 8

_SENTENCE = re.compile(
    r'(?i)(?:the (?:text|image|cell|number|value)[^."]*?'
    r'(?:is|shows|reads|contains))[:.]?\s*(.*)')
_META = re.compile(r'(?i)transcribed text|note:|^\*+$')
_NUMERIC_TOKEN = re.compile(r'[\d][\d,.]*')


def load_key(env_var="NVIDIA_API_KEY"):
    """Return the API key from the environment or .env, or None if unset."""
    key = os.environ.get(env_var)
    if key:
        return key
    env = os.path.join(os.path.dirname(__file__), "..", "..", ".env")
    if os.path.exists(env):
        for line in open(env):
            if line.startswith(env_var + "="):
                return line.split("=", 1)[1].strip() or None
    return None


def normalize_reply(content, numeric=False):
    """Turn a chatty model reply into a bare cell value; '' means empty."""
    if isinstance(content, list):
        content = "".join(p.get("text", "") for p in content)
    s = content.strip().splitlines()[0].strip()
    m = re.search(r'"([^"]*)"', s)
    if m:
        s = m.group(1)
    else:
        m = _SENTENCE.match(s)
        if m:
            s = m.group(1)
    s = re.sub(r'(?i)^the (?:number|text|value|image)\s+(?:is\s+)?', '', s)
    s = s.replace("**", "").strip().strip('"').strip("'").strip().rstrip(".")
    if not s or s.upper() == "EMPTY" or _META.search(s):
        return ""
    if numeric and not NUMERIC.match(s):
        toks = _NUMERIC_TOKEN.findall(s)
        if toks:
            s = toks[-1]
    return s


def encode_crop(path):
    """Upscale a saved cell crop and return (base64_png, original_width)."""
    img = cv2.imread(path)
    if img is None:
        return None, 0
    w = img.shape[1]
    big = cv2.resize(img, None, fx=3, fy=3, interpolation=cv2.INTER_CUBIC)
    _, png = cv2.imencode(".png", big)
    return base64.b64encode(png.tobytes()).decode(), w


def ask(key, model, b64, label, numeric=False):
    """Transcribe one cell crop via the chat-completions API. Retries on any
    failure; returns None if all attempts fail."""
    payload = {
        "model": model, "max_tokens": 48, "temperature": 0,
        "messages": [{"role": "user", "content": [
            {"type": "text", "text": (
                "This image is one cell from a scanned financial table."
                + (f" Its row is labelled '{label}'." if label else "")
                + " Transcribe exactly the characters visible - typically a"
                  " numeric amount or a short text label. Never invent text:"
                  " if it shows only table borders or noise, reply EMPTY.")},
            {"type": "image_url",
             "image_url": {"url": f"data:image/png;base64,{b64}"}},
        ]}],
    }
    req = urllib.request.Request(
        URL, data=json.dumps(payload).encode(),
        headers={"Authorization": f"Bearer {key}",
                 "Content-Type": "application/json"})
    for attempt in range(RETRIES):
        try:
            with urllib.request.urlopen(req, timeout=TIMEOUT) as resp:
                body = json.load(resp)
                return normalize_reply(
                    body["choices"][0]["message"]["content"], numeric)
        except Exception as e:  # noqa: BLE001 - retried, then logged per cell
            if attempt < RETRIES - 1:
                time.sleep(3)
                continue
            print(f"  api error ({model}): {e}", file=sys.stderr)
            return None


def suspect(new, old, old_conf, crop_w, numeric_col):
    """Decide whether the fast model's answer needs a second opinion."""
    conf = float(old_conf or 0)
    if not new:
        return bool(re.search(r"[A-Za-z0-9]", old)) and conf >= 50
    if len(new) * CHAR_PX > crop_w:
        return True
    if numeric_col and not NUMERIC.match(new):
        return True
    if len(old) <= 3 and len(new) > len(old) + 2:
        return True
    if old and conf >= 50 and difflib.SequenceMatcher(
            None, old.lower(), new.lower()).ratio() < 0.3:
        return True
    return False


def run(out_dir):
    key = load_key()
    if not key:
        print("NVIDIA_API_KEY not set; skipping vision-LLM refine", file=sys.stderr)
        return
    with open(os.path.join(out_dir, "crops.csv"), newline="") as f:
        flagged = list(csv.DictReader(f))
    with open(os.path.join(out_dir, "tables.csv"), newline="") as f:
        tables = list(csv.reader(f))
    data = tables[1:]
    idx = {(r[0], r[1]): r for r in data}

    numeric = {}
    for r in data:
        for i, v in enumerate(r[3:]):
            if v:
                n = numeric.setdefault((r[0], i), [0, 0])
                n[0] += 1
                n[1] += bool(NUMERIC.match(v))
    is_num = {k: v[0] >= 3 and v[1] > v[0] * 0.6 for k, v in numeric.items()}

    def refine(rec):
        b64, w = encode_crop(os.path.join(out_dir, "crops", rec["crop"]))
        if b64 is None:
            return rec, None, ""
        numcol = is_num.get((rec["file"], int(rec["col"]) - 1), False)
        new = ask(key, MODEL_FAST, b64, rec["label"], numcol)
        model = "fast"
        if new is not None and suspect(new, rec["text"], rec["conf"], w, numcol):
            second = ask(key, MODEL_SLOW, b64, rec["label"], numcol)
            if second is not None:
                new, model = second, "slow"
        return rec, new, model

    changed = 0
    with open(os.path.join(out_dir, "refine_log.csv"), "w", newline="") as lf:
        lw = csv.writer(lf)
        lw.writerow(["file", "row", "col", "old", "conf", "new", "model"])
        with ThreadPoolExecutor(max_workers=WORKERS) as ex:
            for i, (rec, new, model) in enumerate(ex.map(refine, flagged), 1):
                if new is None:
                    continue
                row = idx.get((rec["file"], rec["row"]))
                col = int(rec["col"])
                if row and new != rec["text"]:
                    if len(row) <= 2 + col:
                        row.extend([""] * (3 + col - len(row)))
                    lw.writerow([rec["file"], rec["row"], col,
                                 rec["text"], rec["conf"], new, model])
                    row[2 + col] = new
                    row[2] = (row[2] + ";" + model if row[2] else model)
                    changed += 1
                if i % 100 == 0 or i == len(flagged):
                    print(f"{i}/{len(flagged)} refined, {changed} changed",
                          flush=True)

    with open(os.path.join(out_dir, "tables_refined.csv"), "w", newline="") as tf:
        csv.writer(tf).writerows(tables)
    print(f"{len(flagged)} cells sent, {changed} corrected")
    print(f"wrote {out_dir}/tables_refined.csv and refine_log.csv")


def main(argv):
    if len(argv) < 1:
        sys.exit("usage: refine <out_dir>")
    run(argv[0])
