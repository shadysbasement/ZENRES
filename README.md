# ZENRES

**Zero Emission Vision, Responsible Advancement.** ZENRES turns Indian MSME factory records into audit-ready EU CBAM evidence, verifiable dual-QR passports and revenue-generating return freight — and puts those factories in front of global B2B buyers.

## Live demo

- Public site: https://zenres.vercel.app
- Demo workspace: https://zenres-app.vercel.app/dashboard
- PGP admin auth API: https://zenres-auth.vercel.app/healthz

The workspace still runs on seeded pilot data. The PGP admin login is backed by the deployed Go auth API and Neon Postgres; the admin operations view still shows sample records and keeps edits in the browser. Ordinary account sign-in, the core API, and the worker need their remaining production integrations before live customer data is used.

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
├── db/migrations/         # 001 auth + vault · 002 core platform · 003 admin handoff · 004 legal consent · 005–006 support
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

A soft glass workspace adapted from the [Hospital Management System ERP reference](https://www.behance.net/gallery/250314135/Hospital-Management-System-ERP-Solution-UI-UX), using ZENRES content and original abstract backgrounds.

- **Typography** — self-hosted Urbanist, with tabular numerals for metrics and IBM Plex Mono for technical identifiers.
- **Palette & surfaces** — blue-gray canvas, translucent white cards, rounded module navigation, blue-to-lilac accents, and muted status colors.
- **Backgrounds** — Silk, Aurora, Contour, Rosewater, and Mist. The appearance control saves the selection and ambient-motion preference on the current device. No DNA imagery or third-party artwork is bundled.
- **Motion** — staggered card entrances, content-only page transitions, smooth module menus, animated charts, drawers, toasts, and optional slow ambient movement. Device reduced-motion preferences disable these effects. The workspace opens directly without an artificial loading delay.
- **Layout** — six overview cards, performance charts, inbox, production records, and an expandable factory twin. All modules share the same forms, tables, and navigation; the admin workspace retains its PGP authentication.

## Highlights

**Public site** — interactive CBAM exposure slider and FreightLink backhaul preview in the hero; four-pillar tabs; 8-step journey; Delhi ↔ Mumbai / Ludhiana ↔ Delhi corridor visualiser with hard-filter logic and the blueprint 180 L → 140 L avoided-fuel example; pricing with monthly/annual and USD/INR/EUR toggles plus the 14-row matrix; ZV-001 → A-101…A-103 passport schematic with role-based visibility; public QR resolver (valid / revoked / unknown).

The public `/features` page distinguishes interactive sample workflows, pilot backend code and capabilities that are not launched. `/status` checks website reachability and the configured auth API health endpoint on each request; it is a basic reachability view, not incident monitoring. The demo upload control keeps the chosen file in the browser session and does not claim secure evidence storage. The PGP admin console labels its sample service signals as demo data.

**Customer support** — `/support` has searchable English/Hindi guides, a reviewed-answer assistant, account-scoped tickets and conversations, callback requests, issue reports, feedback and service notices. The PGP-protected `/ops-vault/support` queue gives managers assignment, priority, response and resolution tracking, satisfaction and actual work-cost summaries; support agents can access only tickets assigned to them. Ticket changes are audited. The assistant uses approved guide text by default; optional model-generated wording requires `SUPPORT_AI_API_KEY`, `SUPPORT_AI_MODEL`, a signed-in user and their explicit opt-in. Only public guide text reaches the model. Set `SUPPORT_EMAIL` and `SUPPORT_WHATSAPP` only after approving real contact details; set `SUPPORT_EMAIL_NOTIFICATIONS=true` only with a configured, verified SMTP sender. Human live chat, response-time commitments and plan-specific entitlements need staffed operations and verified subscriptions; `SUPPORT_PLAN_ENTITLEMENTS_ENFORCED` defaults to false. The ticket conversation is available now, with polling for new replies. Apply migrations `005` and `006` to the auth database before deployment.

**Auth** — email + Argon2id (t=3, 64 MiB, p=4), Google & Apple OIDC with PKCE S256 and JWKS verification, Apple first-auth names and private relay support, safe account linking (never a silent merge), and the privacy-first **Proton flow**: domain/MX recognition → WKD key lookup → optionally PGP-encrypted single-use magic link → Proton Pass passkey enrolment. `ProtonOIDCStub` sits behind `ENABLE_PROTON_OIDC=false` — no fabricated Proton OAuth endpoints.

**Hidden PGP admin vault** — no link, sitemap, robots entry or client-bundle string points at it. The gate lives at `ADMIN_GATE_PATH` behind a catch-all route that 404s every other path identically; `Ctrl+Shift+Alt+Z` (or five Shift-clicks on the footer build ID) opens a passphrase prompt that only the server can resolve. The admin requests an RFC 9580 challenge encrypted to their pinned key, decrypts and clearsigns it offline (`gpg --decrypt | gpg --clearsign`), and the Go verifier checks signer fingerprint, verified bytes (constant-time hash compare), 120 s expiry, /24 subnet, 3-attempt cap and atomic single-use consumption. Unknown identifiers receive an indistinguishable decoy challenge. A one-time handoff code becomes a host-only `__Host-zv_ops` cookie on the app, bound to UA + subnet with 4 h / 30 min expiry.

**Web app** — procedural Threlte/Three.js factory twin (production lines with furnaces and a moving gantry, packaging conveyor with instanced crates sized from live WIP, glass mezzanine offices, racking whose fill heights track stock), camera presets, workers/SKUs/emissions overlays, Studio/CAD modes, 2D SVG fallback and a zone inspector whose edits broadcast to every open tab. Mail bell (Buyers & Agencies / Deadlines / Alerts) with RFQ conversion; ⌘K command palette; dedicated pages for every module in both blueprints; **silent tier gating** — locked modules never reach the DOM and 15 trigger-to-whisper rules surface calm, dismissible hints (14-day cooldown) only when your own actions call for them.

## Enrolling a real admin key

```bash
gpg --armor --export <FINGERPRINT> > founder.asc
make enroll-admin HANDLE=zv-founder-01 KEY=founder.asc
```

Set `ADMIN_GATE_PATH`, `ADMIN_GATE_SECRET` and `ADMIN_GATE_PASSPHRASE_SHA256` (sha256 of your trigger passphrase) in the deployment's secret store — never in git. For local testing without GnuPG, `services/auth-api/cmd/vault-dev-sign` generates a throwaway key and performs the decrypt + clearsign step.

The deployed auth service lives in the `services` Vercel project. Run migrations `001` and `003` against its Postgres database, then enroll a public key with `make enroll-admin`. The public and app projects both need `AUTH_API_URL`; the public site and auth API share `ADMIN_GATE_SECRET`, while the app and auth API share `INTERNAL_API_SECRET`. The one-time handoff is stored in Postgres so it works across Vercel instances.

## Legal release preparation

The six ZENRES legal pages are review drafts available in local development. They return 404 in production, have no production footer links and are excluded from the sitemap until the copy is approved. The legal text in `apps/web-public/src/lib/legal/policies.ts` still contains explicit review notes, so setting environment variables alone cannot publish it. An approved revision must replace those notes and change the `LEGAL_COPY_APPROVED` gate in `apps/web-public/src/lib/server/legal.ts` after legal review.

Before enabling account registration under the approved terms, confirm the actual operator or legal entity, contact address, support and privacy contacts, jurisdiction, retention schedule, policy versions and effective date in `.env.example`; review actual processing and vendor inventory, commercial/refund terms, and any needed DPA or partner agreements. ZENRES is currently an unincorporated project operated by Syed Mohammad Sufyan Ali, so public text must not claim a registered company exists. Apply migration `004_zenres_legal_consent.sql` to the auth database, then configure `ZENRES_LEGAL_APPROVED=true` with the same terms and privacy versions in the public and auth services. The app can set `PUBLIC_LEGAL_PUBLISHED=true` only when those public routes are live. New account creation then requires the current terms acceptance and privacy acknowledgement, stores immutable versioned events, and offers separate optional marketing consent with an account withdrawal control. New Google, Apple and Proton signups are held to the email flow until their own consent capture is implemented.

The retention and deletion/export procedures still require an approved schedule and an operator workflow. A disposable Postgres integration test now verifies cross-tenant isolation for plant reads and writes under `zevra_app`; extend that verification across every tenant table and authenticated role path before live customer records are admitted. Material future terms changes also need a renewal screen for existing accounts. These are release blockers, not capabilities claimed by the draft notice.

Live Tally sync, FreightLink, Connect and QR anchoring backend endpoints default off. Their individual `ZENRES_ENABLE_*` switches are for a later reviewed launch; the current web workspace remains a demo with sample data. Do not enable paid checkout until the reviewed subscription, refund and tax terms are in place.

## Tests

| Suite | Covers |
| --- | --- |
| `services` `go test ./...` | Argon2id policy; PGP vault with real ephemeral keypairs (valid, replay, tampered payload, forged block, wrong key, expired, subnet, attempt cap, decoy, revoked, private-key refusal); auth HTTP flows (cookie flags, single-use Proton link, PKCE, no silent OAuth merge, stealth 404s, CSRF); entitlement 403s; CBAM math and Registry XML; logistics vs embedded separation; corridor matcher (ranking, hard filters, no pay-to-rank, explicit booking); hash-chain tamper detection; idempotent Tally import; SSE fan-out |
| `npm test` | hint engine (all 15 rules, cooldown), manifests, carbon math, Tally CSV parser + idempotency, freight matcher, commercial terms |
| `pytest` | buyer-mail triage, deadlines, OCR confidence flags, Hindi/Devanagari floor-bot parsing, backhaul ranking |
| `python scripts/run-rls-integration.py` | starts a disposable local PostgreSQL database, applies migrations 001-002, and verifies plant read/write isolation under `zevra_app`; requires PostgreSQL command-line tools |
| `python scripts/run-support-integration.py` | starts a disposable PostgreSQL database, applies all migrations and verifies customer ticket isolation, support roles, admin session binding, private notes and work costs |
| Playwright (`apps/*/tests`) | responsive snapshots, pricing/corridor toggles, QR resolver states, stealth 404s and bundle scan, Starter gating, 3D canvas, bell increments, Tally re-import |

## Notes & honest gaps

- The web app uses seeded pilot data in the browser unless `PUBLIC_CORE_API_URL` is set; in live mode it loads the manifest, twin and notifications from the core API and streams updates over SSE. Not every page has a live write path yet.
- Go queries are hand-written with `pgx/v5` rather than sqlc-generated.
- The standard-user flows still use an in-process rate limiter and KV store. Move those to a shared store before using ordinary accounts at scale; the admin handoff already uses Postgres.
- Typst rendering needs the `typst` binary and the `tiaoma` package for QR codes.
- Google's "Continue with Google" button uses a monochrome G so the zero-yellow rule holds; check Google's branding guidelines before launch. The Proton mark is a geometric approximation — replace it with the official asset from Proton's brand kit.
- CBAM figures are illustrative estimates; ZENRES provides reporting support, not legal, tax or accounting advice.
