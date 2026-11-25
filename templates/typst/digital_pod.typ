// FreightLink digital proof of delivery.
#let d = json.decode(sys.inputs.at("data", default: "{}"))
#set page(paper: "a5", margin: 14mm)
#set text(font: ("Switzer", "Helvetica"), size: 9pt)
#text(size: 18pt, weight: 600)[Proof of delivery] #h(1fr) #text(font: "IBM Plex Mono")[#d.at("ref", default: "")]
#v(4mm)
#table(columns: (auto, 1fr), stroke: none, inset: 4pt,
  [Route], [#d.at("route", default: "")], [Vehicle], [#d.at("vehicle", default: "")],
  [Pickup QR], [#d.at("pickup", default: "")], [Delivery QR], [#d.at("delivery", default: "")],
  [Received by], [#d.at("receiver", default: "")], [Delivered at], [#d.at("at", default: "")])
#v(6mm)
#text(size: 7pt, fill: rgb("#6E7D77"))[Avoided-fuel figures are logistics estimates reported separately from CBAM embedded emissions.]
