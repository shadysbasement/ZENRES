"""Zevra intelligence worker (FastAPI). Called only by the Go core API on the private network."""
from __future__ import annotations

import os

from fastapi import Depends, FastAPI, File, Header, HTTPException, UploadFile
from pydantic import BaseModel, Field

from . import backhaul, floorbot, mail, ocr

app = FastAPI(title="zevra-intel-worker", version="0.1.0", docs_url=None, redoc_url=None)
SECRET = os.getenv("WORKER_SHARED_SECRET", "")


def internal(x_worker_secret: str | None = Header(default=None)) -> None:
    if SECRET and x_worker_secret != SECRET:
        raise HTTPException(status_code=404)


class MailIn(BaseModel):
    sender: str = ""
    subject: str = Field(max_length=500)
    body: str = Field(default="", max_length=50_000)


class FloorIn(BaseModel):
    message: str = Field(max_length=2_000)


class Candidate(BaseModel):
    id: str
    detour_km: float
    slack_h: float
    payload_t: float
    load_t: float
    corridor_demand: float = 0.5
    reliability: float = 4.0
    history_trips: int = 0


@app.get("/healthz")
def health() -> dict:
    return {"status": "ok"}


@app.post("/worker/mail/classify", dependencies=[Depends(internal)])
def classify_mail(m: MailIn) -> dict:
    return mail.classify(m.sender, m.subject, m.body)


@app.post("/worker/ocr/extract", dependencies=[Depends(internal)])
async def ocr_extract(file: UploadFile = File(...)) -> dict:
    data = await file.read()
    if len(data) > 20 * 1024 * 1024:
        raise HTTPException(status_code=413, detail="file_too_large")
    text = ocr.text_from_bytes(data, file.filename or "upload")
    return {"filename": file.filename, **ocr.extract(text)}


@app.post("/worker/floorbot/parse", dependencies=[Depends(internal)])
def floorbot_parse(m: FloorIn) -> dict:
    return floorbot.parse(m.message)


@app.post("/worker/backhaul/rank", dependencies=[Depends(internal)])
def backhaul_rank(cands: list[Candidate]) -> list[dict]:
    return backhaul.rank([c.model_dump() for c in cands])
