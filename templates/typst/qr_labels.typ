// Zevra printable QR label sheet (A4, 3 × 8 labels). Invoked by core-api:
//   typst compile --input data='{"labels":[{"code":"ZV-001","token":"…","url":"…"}]}' qr_labels.typ -
#import "@preview/tiaoma:0.3.0": qrcode
#let data = json.decode(sys.inputs.at("data", default: "{\"labels\":[]}"))
#set page(paper: "a4", margin: 10mm)
#set text(font: ("Switzer", "Helvetica"), size: 8pt, fill: rgb("#0E1412"))
#grid(columns: 3, gutter: 4mm,
  ..data.labels.map(l => box(width: 100%, height: 33mm, stroke: 0.4pt + rgb("#DCE3DF"), radius: 2pt, inset: 3mm)[
    #grid(columns: (22mm, 1fr), gutter: 3mm,
      qrcode(l.url, width: 22mm),
      [
        #text(size: 13pt, weight: 600)[#l.code] \
        #v(1mm)
        #text(font: "IBM Plex Mono", size: 6pt, fill: rgb("#6E7D77"))[#l.token] \
        #v(2mm)
        #text(size: 6pt, fill: rgb("#14533C"))[Scan to verify · zevra.io]
      ])
  ])
)
