"""Vernacular floor-bot parser for supervisor WhatsApp messages (Hindi / Marathi / Hinglish / English)."""
from __future__ import annotations

import re

NUM_WORDS = {"ek": 1, "do": 2, "teen": 3, "char": 4, "chaar": 4, "paanch": 5, "panch": 5, "che": 6, "saat": 7, "aath": 8, "nau": 9, "das": 10,
             "एक": 1, "दो": 2, "तीन": 3, "चार": 4, "पांच": 5, "छह": 6, "सात": 7, "आठ": 8, "नौ": 9, "दस": 10}
DEVANAGARI_DIGITS = str.maketrans("०१२३४५६७८९", "0123456789")

LINE_RE = re.compile(r"\b(?:line|लाइन)\s*([ab])\b", re.I)
TONNES_RE = re.compile(r"(\d+(?:\.\d+)?)\s*(?:ton|tonne|tonnes|t|टन)\b", re.I)
BATCH_RE = re.compile(r"\b([A-Z]-\d{3})\b", re.I)
WORKERS_RE = re.compile(r"(\d+)\s*(?:log|logo|worker|workers|staff|लोग|कामगार|कर्मचारी)\b", re.I)
DIESEL_RE = re.compile(r"(\d+(?:\.\d+)?)\s*(?:litre|liter|ltr|l|लीटर)\b", re.I)
SHIFT_RE = re.compile(r"\bshift\s*[-]?\s*(\d)\b", re.I)


def normalise(msg: str) -> str:
    msg = msg.translate(DEVANAGARI_DIGITS)
    for word, n in NUM_WORDS.items():
        msg = re.sub(rf"(?<!\w){re.escape(word)}(?!\w)", str(n), msg, flags=re.I)
    return msg


def parse(message: str) -> dict:
    text = normalise(message)
    out: dict = {"intents": []}
    if m := LINE_RE.search(text):
        out["zone_code"] = f"PROD-{m.group(1).upper()}"
    if m := BATCH_RE.search(text):
        out["batch_code"] = m.group(1).upper()
    if m := TONNES_RE.search(text):
        out["produced_t"] = float(m.group(1))
        out["intents"].append("log_production")
    if m := WORKERS_RE.search(text):
        out["active_workers"] = int(m.group(1))
        out["intents"].append("update_headcount")
    if m := DIESEL_RE.search(text):
        litres = float(m.group(1))
        out["diesel_l"] = litres
        out["scope1_tco2e"] = round(litres * 0.00268, 3)
        out["intents"].append("log_fuel")
    if m := SHIFT_RE.search(text):
        out["shift"] = int(m.group(1))
    out["confidence"] = round(min(0.98, 0.45 + 0.18 * len(out["intents"])), 2)
    return out
