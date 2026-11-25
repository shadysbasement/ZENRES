// EU Customs Audit Defence Pack (EU Importer Hub). Deterministic, sub-second render.
#let d = json.decode(sys.inputs.at("data", default: "{}"))
#set page(paper: "a4", margin: (x: 18mm, y: 20mm),
  header: text(size: 7pt, fill: rgb("#6E7D77"))[ZEVRA · EU CUSTOMS AUDIT DEFENCE PACK · #d.at("period", default: "")],
  footer: context text(size: 7pt, fill: rgb("#6E7D77"))[Reporting support only — not legal, tax or accounting advice. #h(1fr) #counter(page).display()])
#set text(font: ("Switzer", "Helvetica"), size: 9.5pt, fill: rgb("#0E1412"))
#text(size: 22pt, weight: 600)[Audit defence pack] \
#text(fill: rgb("#14533C"), style: "italic")[#d.at("declarant", default: "Declarant")] · EORI #d.at("eori", default: "—")
#v(6mm)
#table(columns: (auto, 1fr, auto, auto, auto), stroke: 0.4pt + rgb("#DCE3DF"), inset: 6pt,
  [*CN*], [*Supplier · goods*], [*t*], [*tCO₂e/t*], [*Quality*],
  ..d.at("lines", default: ()).map(l => (l.cn, l.supplier + " · " + l.name, str(l.tonnes), str(l.see), l.quality)).flatten())
#v(4mm)
== Evidence trail
Every figure above links to evidence in the Zevra Audit Locker. Hash-chain head: #raw(d.at("chain_head", default: "—"))
