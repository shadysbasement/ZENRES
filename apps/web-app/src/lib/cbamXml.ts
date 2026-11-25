/** Deterministic CBAM Registry XML / Excel builders (mirrors services/core-api/internal/cbam/xml.go). */
export interface GoodsLine {
	cn: string;
	name: string;
	tonnes: number;
	direct: number;
	indirect: number;
	see: number;
	quality: string;
}

const esc = (s: string) => s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');

export function buildRegistryXml(o: { quarter: string; installation: { name: string; operator: string; unlocode: string; country: string }; declarant: string; eori: string; goods: GoodsLine[]; consolidated?: boolean }): string {
	const lines = o.goods
		.map(
			(g, i) => `    <GoodsImported seq="${i + 1}">
      <CommodityCode>${g.cn}</CommodityCode>
      <Description>${esc(g.name)}</Description>
      <NetMassTonnes>${g.tonnes.toFixed(3)}</NetMassTonnes>
      <SpecificEmbeddedEmissions unit="tCO2e/t">
        <Direct>${g.direct.toFixed(4)}</Direct>
        <Indirect>${g.indirect.toFixed(4)}</Indirect>
        <Total>${g.see.toFixed(4)}</Total>
      </SpecificEmbeddedEmissions>
      <DeterminationMethod>${g.quality === 'estimated' ? 'DEFAULT_VALUES' : 'ACTUAL_DATA'}</DeterminationMethod>
      <DataQuality>${g.quality.toUpperCase()}</DataQuality>
    </GoodsImported>`
		)
		.join('\n');
	return `<?xml version="1.0" encoding="UTF-8"?>
<CBAMReport xmlns="urn:eu:taxud:cbam:v1" ruleSet="IR (EU) 2025/2621" period="${o.quarter}" type="${o.consolidated ? 'CONSOLIDATED' : 'INSTALLATION'}">
  <Declarant>
    <Name>${esc(o.declarant)}</Name>
    <EORI>${o.eori}</EORI>
  </Declarant>
  <Installation>
    <Name>${esc(o.installation.name)}</Name>
    <Operator>${esc(o.installation.operator)}</Operator>
    <UNLOCODE>${o.installation.unlocode}</UNLOCODE>
    <Country>${o.installation.country}</Country>
  </Installation>
  <Goods>
${lines}
  </Goods>
  <Disclaimer>Prepared with ZENRES reporting support. Not legal, tax or accounting advice.</Disclaimer>
</CBAMReport>
`;
}

export function buildExcelCsv(goods: GoodsLine[]): string {
	const head = 'CN code,Description,Net mass (t),Direct tCO2e/t,Indirect tCO2e/t,Total tCO2e/t,Data quality';
	return [head, ...goods.map((g) => [g.cn, `"${g.name.replace(/"/g, '""')}"`, g.tonnes.toFixed(3), g.direct, g.indirect, g.see, g.quality].join(','))].join('\n');
}
