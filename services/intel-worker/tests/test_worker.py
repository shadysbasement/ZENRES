from datetime import datetime, timezone

from fastapi.testclient import TestClient

from app import backhaul, floorbot, mail, ocr
from app.main import app

NOW = datetime(2026, 11, 26, 9, 0, tzinfo=timezone.utc)
client = TestClient(app)


def test_eu_buyer_rfq_is_classified_with_deadline_and_cn():
    r = mail.classify("l.hoffmann@rhein-alu-import.de", "RFQ — 40 t 6063-T6 bars + CBAM XML",
                      "Please quote 40 t for Q1. We need your CBAM XML by 15 Dec 2026.", NOW)
    assert r["category"] == "buyer_agency"
    assert r["tag"] == "CN 76042910"
    assert r["deadline"].startswith("2026-12-15")
    assert r["severity"] == "elevated"


def test_compliance_deadline_and_overdue():
    r = mail.classify("calendar@zevra.io", "Q4 embedded emissions due", "Supplier data is due by 20 Nov 2026.", NOW)
    assert r["category"] == "deadline" and r["severity"] == "critical"


def test_operational_alert():
    r = mail.classify("system@zevra.io", "Tally sync mismatch", "2 vouchers need review", NOW)
    assert r["category"] == "operational_alert" and r["deadline"] is None


def test_ocr_flags_low_confidence_fields():
    text = "MSEDCL HT Consumer\nConsumer No: 170019284551\nUnits Consumed (kWh): 421,560\nMax Demand (kVA): 1,180\nAmount Payable: Rs 38,41,200"
    out = ocr.extract(text)
    fields = {f["field"]: f for f in out["fields"]}
    assert out["record_type"] == "electricity"
    assert fields["kwh"]["value"] == "421,560"
    assert fields["max_demand_kva"]["needs_review"] is True
    assert fields["amount_inr"]["needs_review"] is False


def test_hindi_floorbot_message():
    out = floorbot.parse("Line A 4.2 ton, batch A-104. 11 log present. DG 96 लीटर")
    assert out["zone_code"] == "PROD-A" and out["batch_code"] == "A-104"
    assert out["produced_t"] == 4.2 and out["active_workers"] == 11
    assert out["scope1_tco2e"] == 0.257


def test_devanagari_digits_and_number_words():
    out = floorbot.parse("लाइन B में ५ टन, das log")
    assert out["zone_code"] == "PROD-B" and out["produced_t"] == 5 and out["active_workers"] == 10


def test_backhaul_ranking_prefers_low_detour_and_reports_confidence():
    ranked = backhaul.rank([
        {"id": "far", "detour_km": 90, "slack_h": 20, "payload_t": 21, "load_t": 14, "history_trips": 2},
        {"id": "near", "detour_km": 10, "slack_h": 2, "payload_t": 16, "load_t": 14, "history_trips": 25},
    ])
    assert ranked[0]["id"] == "near" and 0 < ranked[0]["confidence"] <= 1


def test_http_endpoints():
    assert client.get("/healthz").json() == {"status": "ok"}
    r = client.post("/worker/mail/classify", json={"sender": "arjun@euro-sourcing-agents.com", "subject": "Sample order request", "body": "500 kg of 6061 profiles"})
    assert r.status_code == 200 and r.json()["category"] == "buyer_agency"
    r = client.post("/worker/ocr/extract", files={"file": ("bill.txt", b"Units Consumed (kWh): 1,000\nAmount Payable: Rs 9,400", "text/plain")})
    assert r.json()["record_type"] == "electricity"
