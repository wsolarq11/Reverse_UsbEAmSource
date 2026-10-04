#!/usr/bin/env node
'use strict';
// 从 acceptance 目录生成批次索引表，写入 docs/BATCH_INDEX.md。
// 幂等：只读 acceptance 目录，重写 BATCH_INDEX.md（重写结果一致）。
const fs = require('fs');
const path = require('path');

const docsDir = path.join(__dirname, '..', 'docs');
const accDir = path.join(docsDir, 'acceptance');

function numOf(name) {
  const m = name.match(/^(?:batch)?(\d+)\.md$/i);
  return m ? parseInt(m[1], 10) : null;
}

function titleOf(file) {
  const txt = fs.readFileSync(path.join(accDir, file), 'utf8');
  for (const line of txt.split('\n')) {
    const m = line.match(/^#\s+(.*)$/);
    if (m) return m[1].trim();
  }
  return '';
}

const files = fs.readdirSync(accDir).filter(f => /\.md$/.test(f));
const rows = [];
for (const f of files) {
  const n = numOf(f);
  if (n === null) continue;
  rows.push({ n, f, t: titleOf(f) });
}
rows.sort((a, b) => a.n - b.n);

let out = '# 批次索引（UsbEAm Launcher 1.0.3 逆向还原）\n\n';
out += '> 由 `tools/gen_batch_index.js` 生成，勿手改。每批验收文档的唯一入口索引。\n\n';
out += '| 批次 | 验收文档 | 摘要 |\n|---|---|---|\n';
out += '| 19–39 | [batch19-39.md](acceptance/batch19-39.md) | 早期批次归档（memoryrelease/SQLite/mouseGesture/OLEDBlackout/Screenshot/launcherasset/icon/windowManagement） |\n';
for (const r of rows) {
  out += `| ${r.n} | [${r.f}](${`acceptance/${r.f}`}) | ${r.t} |\n`;
}

fs.writeFileSync(path.join(docsDir, 'BATCH_INDEX.md'), out, 'utf8');
console.log(`WROTE BATCH_INDEX.md (${out.split('\n').length - 1} lines)`);
