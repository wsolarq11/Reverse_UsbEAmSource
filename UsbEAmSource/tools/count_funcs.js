#!/usr/bin/env node
'use strict';
// 与 tools/count_funcs.sh 等价的口径统计（bash/awk 在本 Windows 环境不可用）。
// 输出键：FUNCS / MARKED / UNMARKED / S / S-inline / S-eq / S-sig / P / FAITHFUL / USABLE。
// 标记优先级：S-sig > S-eq > S-inline > S > P；仅在紧跟 func 的连续注释块首行取标记。
const fs = require('fs');
const path = require('path');

const backendDir = path.join(__dirname, '..', 'backend');

function marker(s) {
  if (/\[S-sig(\s|\])/.test(s)) return 'S-sig';
  if (/\[S-eq(\s|\])/.test(s)) return 'S-eq';
  if (/\[S-inline(\s|\])/.test(s)) return 'S-inline';
  if (/\[S(\s|\]|\t)/.test(s)) return 'S';
  if (/\[P(\s|\]|\t)/.test(s)) return 'P';
  return '';
}

const counts = {};
const add = (k, n = 1) => { counts[k] = (counts[k] || 0) + n; };

const files = fs.readdirSync(backendDir).filter(f => f.endsWith('.go') && !f.endsWith('_test.go'));
for (const f of files) {
  const lines = fs.readFileSync(path.join(backendDir, f), 'utf8').split('\n');
  let last = '';
  for (const line of lines) {
    if (line.startsWith('//')) {
      if (last === '') last = marker(line);
      continue;
    }
    if (line.startsWith('func ')) {
      if (last !== '') { add(last); add('MARKED'); } else { add('UNMARKED'); }
      add('FUNCS');
      last = '';
      continue;
    }
    last = '';
  }
}

const s = counts.S || 0;
const inline = counts['S-inline'] || 0;
const eq = counts['S-eq'] || 0;
const sig = counts['S-sig'] || 0;
const p = counts.P || 0;
counts.FAITHFUL = s + inline;
counts.USABLE = s + inline + eq;
counts.TRUE = s + inline + sig;

const order = ['FUNCS', 'MARKED', 'UNMARKED', 'S', 'S-inline', 'S-eq', 'S-sig', 'P', 'FAITHFUL', 'USABLE', 'TRUE'];
const out = [];
for (const k of order) out.push(`${k}=${counts[k] || 0}`);
console.log(out.join('  '));
