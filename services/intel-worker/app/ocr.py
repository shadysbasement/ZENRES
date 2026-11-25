"""Utility-bill / invoice field extraction with per-field confidence for human-in-the-loop review.

Text is taken from a text layer (pdfplumber) or OCR (pytesseract) when the optional `ocr` extra is
installed; plain-text uploads are parsed directly. Low-confidence fields are flagged, never auto-posted.
"""
from __future__ import annotations

import io
import re

FIELDS = {
    "supplier": (re.compile(r"(MSEDCL|MSETCL|TATA POWER|ADANI ELECTRICITY|BSES|TORRENT POWER|MNGL|IGL|HINDALCO|NALCO)[^\n]*", re.I), 0.97),
    "consumer_no": (re.compile(r"(?:consumer|ca|account)\s*(?:no\.?|number)\s*[:\-]?\s*([0-9]{8,14})", re.I), 0.95),
    "billing_period": (re.compile(r"(?:billing\s*period|bill\s*period|period)\s*[:\-]?\s*([0-9]{1,2}[-/ ][A-Za-z0-9]{2,3}[-/ ][0-9]{2,4}\s*(?:to|-|→)\s*[0-9]{1,2}[-/ ][A-Za-z0-9]{2,3}[-/ ][0-9]{2,4})", re.I), 0.9),
    "kwh": (re.compile(r"(?:units\s*consumed|total\s*units|kwh)\s*(?:\(kwh\))?\s*[:\-]?\s*([0-9][0-9,]*\.?[0-9]*)", re.I), 0.93),
    "max_demand_kva": (re.compile(r"(?:max(?:imum)?\s*demand|md)\s*(?:\(kva\))?\s*[:\-]?\s*([0-9][0-9,]*\.?[0-9]*)", re.I), 0.72),
    "diesel_litres": (re.compile(r"(?:diesel|hsd)\s*(?:qty|quantity|litres|ltrs|l)?\s*[:\-]?\s*([0-9][0-9,]*\.?[0-9]*)", re.I), 0.88),
    "amount_inr": (re.compile(r"(?:amount\s*payable|net\s*payable|total\s*amount|grand\s*total)\s*[:\-]?\s*(?:rs\.?|₹|inr)?\s*([0-9][0-9,]*\.?[0-9]*)", re.I), 0.96),
}

REVIEW_THRESHOLD = 0.8


def text_from_bytes(data: bytes, filename: str) -> str:
    name = filename.lower()
    if name.endswith((".txt", ".csv")):
        return data.decode("utf-8", errors="ignore")
    if name.endswith(".pdf"):
        try:
            import pdfplumber  # type: ignore

            with pdfplumber.open(io.BytesIO(data)) as pdf:
                return "\n".join(p.extract_text() or "" for p in pdf.pages)
        except ImportError:
            pass
    try:
        import pytesseract  # type: ignore
        from PIL import Image  # type: ignore

        return pytesseract.image_to_string(Image.open(io.BytesIO(data)))
    except Exception:  # noqa: BLE001 — OCR extra not installed or unreadable image
        return data.decode("utf-8", errors="ignore")


def extract(text: str) -> dict:
    fields = []
    for key, (rx, base_conf) in FIELDS.items():
        m = rx.search(text)
        if not m:
            continue
        value = (m.group(1) if m.groups() else m.group(0)).strip()
        conf = base_conf
        if key in ("kwh", "amount_inr", "diesel_litres", "max_demand_kva"):
            try:
                float(value.replace(",", ""))
            except ValueError:
                conf -= 0.3
        fields.append({"field": key, "value": value, "confidence": round(conf, 2), "needs_review": conf < REVIEW_THRESHOLD})
    record_type = "electricity" if any(f["field"] == "kwh" for f in fields) else "fuel" if any(f["field"] == "diesel_litres" for f in fields) else None
    return {"record_type": record_type, "fields": fields, "needs_review": any(f["needs_review"] for f in fields) or not fields}
