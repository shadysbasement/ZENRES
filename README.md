# ZENRES

**Zero Emission Vision, Responsible Advancement.** ZENRES turns Indian MSME factory records into audit-ready EU CBAM evidence, verifiable dual-QR passports and revenue-generating return freight — and puts those factories in front of global B2B buyers.

This monorepo implements both Codex plans:

| Plan | What's here |
| --- | --- |
| **01 — Public site, auth & hidden PGP vault** | `apps/web-public` (SvelteKit 2 / Svelte 5) · `services/auth-api` (Go) |
| **02 — Web application** | `apps/web-app` (SvelteKit + Threlte 3D twin + ECharts) · `services/core-api` (Go, PostgreSQL RLS) · `services/intel-worker` (FastAPI) · `templates/typst` |

```
zevra/
├── apps/
│   ├── web-public/        # public site — landing, pricing, FreightLink, QR resolver, auth, onboarding, hidden vault
│   └── web-app/           # workspace — 3D digital twin, all four pillars, silent tier gating, ops vault
├── packages/design/       # shared tokens, self-hosted fonts, motion primitives, preloader, cursor
├── services/
│   ├── auth-api/          # Argon2id, Google/Apple OIDC+PKCE, Proton magic link + WKD + passkeys, PGP vault
│   ├── core-api/          # entitlements, RLS, SSE twin, Tally, CBAM/XML, QR, FreightLink, Connect, Importer Hub, seed
│   ├── intel-worker/      # OCR extraction, buyer-mail triage, vernacular floor-bot, backhaul ranking
│   └── internal/platform/ # shared Go helpers
├── db/migrations/         # 001 auth + vault · 002 core platform with row-level security
├── templates/typst/       # QR label sheets, EU Customs Audit Defence Pack, digital POD
└── scripts/check-palette.mjs   # zero-yellow / banned-font guardrail
```

---

## Quick start — the design, in two commands

The two web apps run standalone on seeded pilot data (Deccan Alloy Works, an aluminium MSME in Chakan, Pune), so you can explore every screen without any backend.

```bash
npm install
npm run dev:public   # http://localhost:5173
```

```bash
npm run dev:app      # http://localhost:5174
```

In the app, open the avatar menu → **Demo organisation** to switch between **MSME Starter**, **MSME Growth** and **EU Importer Hub** and watch the workspace re-shape itself.

## Full stack

```bash
cp .env.example .env
docker compose up --build          # Postgres 17, Redis 7, auth-api, core-api, intel-worker, both apps
docker compose run --rm seed       # three demo orgs; prints logins (password: ZENRES-Pilot-2026!)
```

Without Docker:

```bash
make dev-auth        # Go auth service with an in-memory store (no Postgres needed)
```

```bash
make test            # vitest + go vet/test + pytest + palette guardrail
```

---

## Design language

Swiss industrial minimalism with a portfolio-grade motion layer:

- **Typography** — self-hosted Switzer (UI), Newsreader italic (editorial accents), IBM Plex Mono with tabular numerals for every figure. No Inter / Roboto / Poppins / Space Grotesk.
- **Palette** — alabaster canvas, carbon obsidian, deep-forest emerald, steel cobalt; terracotta replaces any warning yellow. `node scripts/check-palette.mjs` fails the build on yellow/gold/amber or banned fonts.
- **Motion** — counting preloader with letter-by-letter wordmark and curtain lift, Lenis smooth scrolling, masked line-by-line headline reveals, scroll-scrubbed manifesto, pinned horizontal journey, magnetic buttons, custom cursor, view-transition page changes, animated SVG corridors. Everything respects `prefers-reduced-motion`.

## Highlights

**Public site** — interactive CBAM exposure slider and FreightLink backhaul preview in the hero; four-pillar tabs; 8-step journey; Delhi ↔ Mumbai / Ludhiana ↔ Delhi corridor visualiser with hard-filter logic and the blueprint 180 L → 140 L avoided-fuel example; pricing with monthly/annual and USD/INR/EUR toggles plus the 14-row matrix; ZV-001 → A-101…A-103 passport schematic with role-based visibility; public QR resolver (valid / revoked / unknown).

**Auth** — email + Argon2id (t=3, 64 MiB, p=4), Google & Apple OIDC with PKCE S256 and JWKS verification, Apple first-auth names and private relay support, safe account linking (never a silent merge), and the privacy-first **Proton flow**: domain/MX recognition → WKD key lookup → optionally PGP-encrypted single-use magic link → Proton Pass passkey enrolment. `ProtonOIDCStub` sits behind `ENABLE_PROTON_OIDC=false` — no fabricated Proton OAuth endpoints.

**Hidden PGP admin vault** — no link, sitemap, robots entry or client-bundle string points at it. The gate lives at `ADMIN_GATE_PATH` behind a catch-all route that 404s every other path identically; `Ctrl+Shift+Alt+Z` (or five Shift-clicks on the footer build ID) opens a passphrase prompt that only the server can resolve. The admin requests an RFC 9580 challenge encrypted to their pinned key, decrypts and clearsigns it offline (`gpg --decrypt | gpg --clearsign`), and the Go verifier checks signer fingerprint, verified bytes (constant-time hash compare), 120 s expiry, /24 subnet, 3-attempt cap and atomic single-use consumption. Unknown identifiers receive an indistinguishable decoy challenge. A one-time handoff code becomes a host-only `__Host-zv_ops` cookie on the app, bound to UA + subnet with 4 h / 30 min expiry.

**Web app** — procedural Threlte/Three.js factory twin (production lines with furnaces and a moving gantry, packaging conveyor with instanced crates sized from live WIP, glass mezzanine offices, racking whose fill heights track stock), camera presets, workers/SKUs/emissions overlays, Studio/CAD modes, 2D SVG fallback and a zone inspector whose edits broadcast to every open tab. Mail bell (Buyers & Agencies / Deadlines / Alerts) with RFQ conversion; ⌘K command palette; dedicated pages for every module in both blueprints; **silent tier gating** — locked modules never reach the DOM and 15 trigger-to-whisper rules surface calm, dismissible hints (14-day cooldown) only when your own actions call for them.

## Enrolling a real admin key

```bash
gpg --armor --export <FINGERPRINT> > founder.asc
make enroll-admin HANDLE=zv-founder-01 KEY=founder.asc
```

Set `ADMIN_GATE_PATH`, `ADMIN_GATE_SECRET` and `ADMIN_GATE_PASSPHRASE_SHA256` (sha256 of your trigger passphrase) in the deployment's secret store — never in git. For local testing without GnuPG, `services/auth-api/cmd/vault-dev-sign` generates a throwaway key and performs the decrypt + clearsign step.

## Tests

| Suite | Covers |
| --- | --- |
| `services` `go test ./...` | Argon2id policy; PGP vault with real ephemeral keypairs (valid, replay, tampered payload, forged block, wrong key, expired, subnet, attempt cap, decoy, revoked, private-key refusal); auth HTTP flows (cookie flags, single-use Proton link, PKCE, no silent OAuth merge, stealth 404s, CSRF); entitlement 403s; CBAM math and Registry XML; logistics vs embedded separation; corridor matcher (ranking, hard filters, no pay-to-rank, explicit booking); hash-chain tamper detection; idempotent Tally import; SSE fan-out |
| `npm test` | hint engine (all 15 rules, cooldown), manifests, carbon math, Tally CSV parser + idempotency, freight matcher, commercial terms |
| `pytest` | buyer-mail triage, deadlines, OCR confidence flags, Hindi/Devanagari floor-bot parsing, backhaul ranking |
| Playwright (`apps/*/tests`) | responsive snapshots, pricing/corridor toggles, QR resolver states, stealth 404s and bundle scan, Starter gating, 3D canvas, bell increments, Tally re-import |

## Notes & honest gaps

- The web app uses seeded pilot data in the browser unless `PUBLIC_CORE_API_URL` is set; in live mode it loads the manifest, twin and notifications from the core API and streams updates over SSE. Not every page has a live write path yet.
- Go queries are hand-written with `pgx/v5` rather than sqlc-generated.
- The in-process rate limiter and KV store are single-replica; swap in Redis for multi-replica deployments.
- Typst rendering needs the `typst` binary and the `tiaoma` package for QR codes.
- Google's "Continue with Google" button uses a monochrome G so the zero-yellow rule holds; check Google's branding guidelines before launch. The Proton mark is a geometric approximation — replace it with the official asset from Proton's brand kit.
- CBAM figures are illustrative estimates; ZENRES provides reporting support, not legal, tax or accounting advice.
