"""Tier-3 predictive backhaul ranking (EU Importer Hub). Suggestions only — bookings stay explicit.

A transparent logistic score over detour, timing slack, payload fit, historical corridor demand and
carrier reliability, returned with a model-confidence value so the UI can flag weak recommendations.
"""
from __future__ import annotations

import math

WEIGHTS = {"detour": -0.035, "slack_h": -0.02, "fit": 2.4, "demand": 1.6, "reliability": 0.9}
BIAS = -1.2


def score(c: dict) -> dict:
    fit = 1 - abs(c["payload_t"] - c["load_t"]) / max(c["payload_t"], 1)
    z = (BIAS + WEIGHTS["detour"] * c["detour_km"] + WEIGHTS["slack_h"] * c["slack_h"] + WEIGHTS["fit"] * fit
         + WEIGHTS["demand"] * c.get("corridor_demand", 0.5) + WEIGHTS["reliability"] * (c.get("reliability", 4) - 3))
    p = 1 / (1 + math.exp(-z))
    evidence = min(1.0, c.get("history_trips", 0) / 20)
    return {"id": c["id"], "probability": round(p, 3), "confidence": round(0.5 + 0.5 * evidence * (1 - abs(0.5 - p)), 3)}


def rank(candidates: list[dict]) -> list[dict]:
    return sorted((score(c) for c in candidates), key=lambda r: r["probability"], reverse=True)
