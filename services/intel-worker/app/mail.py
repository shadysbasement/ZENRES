"""Buyer / agency mail triage: category, CN-code tag, deadline extraction and severity."""
from __future__ import annotations

import re
from datetime import datetime, timedelta, timezone

BUYER_KEYWORDS = ("rfq", "cbam xml", "sample order", "container specs", "price quote", "quotation", "request for quote", "purchase order")
DEADLINE_KEYWORDS = ("deadline", "due", "by ", "before ", "no later than", "submission")
ALERT_KEYWORDS = ("sync", "low stock", "mismatch", "dispute", "pickup", "delivered", "counter-offer")
EU_TLDS = (".de", ".nl", ".it", ".fr", ".be", ".es", ".pl", ".dk", ".se", ".at", ".ie", ".pt", ".fi")
AGENCY_HINTS = ("sourcing", "buying", "agent", "agency", "trading", "import")

MONTHS = {m: i for i, m in enumerate(["jan", "feb", "mar", "apr", "may", "jun", "jul", "aug", "sep", "oct", "nov", "dec"], start=1)}
CN_RE = re.compile(r"\b(?:CN\s*)?((?:72|73|76|25|28|31)\d{6})\b")
SKU_CN = {"6063": "76042910", "6061": "76042100", "1350": "76051100", "5052": "76061220"}
DATE_RES = [
    re.compile(r"\b(\d{1,2})\s+(jan|feb|mar|apr|may|jun|jul|aug|sep|oct|nov|dec)[a-z]*\.?\s*(\d{4})?", re.I),
    re.compile(r"\b(jan|feb|mar|apr|may|jun|jul|aug|sep|oct|nov|dec)[a-z]*\.?\s+(\d{1,2})(?:st|nd|rd|th)?,?\s*(\d{4})?", re.I),
    re.compile(r"\b(\d{4})-(\d{2})-(\d{2})\b"),
]


def extract_deadline(text: str, now: datetime | None = None) -> datetime | None:
    now = now or datetime.now(timezone.utc)
    lower = text.lower()
    if not any(k in lower for k in DEADLINE_KEYWORDS):
        return None
    for i, rx in enumerate(DATE_RES):
        m = rx.search(text)
        if not m:
            continue
        try:
            if i == 0:
                day, mon, year = int(m.group(1)), MONTHS[m.group(2)[:3].lower()], int(m.group(3) or now.year)
            elif i == 1:
                mon, day, year = MONTHS[m.group(1)[:3].lower()], int(m.group(2)), int(m.group(3) or now.year)
            else:
                year, mon, day = int(m.group(1)), int(m.group(2)), int(m.group(3))
            d = datetime(year, mon, day, 17, 0, tzinfo=timezone.utc)
            if d < now - timedelta(days=1) and not (m.lastindex and m.group(m.lastindex) and len(m.group(m.lastindex)) == 4):
                d = d.replace(year=d.year + 1)
            return d
        except (ValueError, KeyError):
            continue
    rel = re.search(r"within\s+(\d{1,2})\s+days", lower)
    if rel:
        return now + timedelta(days=int(rel.group(1)))
    return None


def detect_cn(text: str) -> str | None:
    m = CN_RE.search(text)
    if m:
        return f"CN {m.group(1)}"
    for alloy, cn in SKU_CN.items():
        if alloy in text:
            return f"CN {cn}"
    return None


def classify(sender: str, subject: str, body: str, now: datetime | None = None) -> dict:
    text = f"{subject}\n{body}"
    lower = text.lower()
    domain = sender.split("@")[-1].lower() if "@" in sender else ""
    deadline = extract_deadline(text, now)
    is_buyer = any(k in lower for k in BUYER_KEYWORDS) or (domain.endswith(EU_TLDS) and ("quote" in lower or "order" in lower))
    if is_buyer:
        category = "buyer_agency"
    elif deadline or "cbam" in lower and "due" in lower:
        category = "deadline"
    elif any(k in lower for k in ALERT_KEYWORDS):
        category = "operational_alert"
    else:
        category = "buyer_agency" if domain.endswith(EU_TLDS) else "operational_alert"

    severity = "normal"
    if deadline:
        days = (deadline - (now or datetime.now(timezone.utc))).days
        severity = "critical" if days < 0 else "elevated" if days <= 21 else "normal"
    if "urgent" in lower:
        severity = "elevated" if severity == "normal" else severity

    org = domain.split(".")[0].replace("-", " ").title() if domain else ""
    role = "Buying house" if any(h in domain or h in lower[:200] for h in AGENCY_HINTS if h != "import") else "EU Importer" if domain.endswith(EU_TLDS) else "Buyer"
    return {
        "category": category,
        "tag": detect_cn(text),
        "deadline": deadline.isoformat() if deadline else None,
        "severity": severity,
        "org_badge": f"{org} · {role}" if org and category == "buyer_agency" else "",
    }
