#!/usr/bin/env node
'use strict';
// 拆分 HANDOFF.md：把静态知识、批次历史归档到独立文件，为精简导航版 HANDOFF 让位。
// 幂等：只读 HANDOFF.md 并写目标文件，可重复运行。归档文件由 git 保留原始 HANDOFF 历史。
const fs = require('fs');
const path = require('path');

const docsDir = path.join(__dirname, '..', 'docs');
const src = fs.readFileSync(path.join(docsDir, 'HANDOFF.md'), 'utf8');
const lines = src.split('\n');

// 1-based 定位：返回标题行号（首个匹配），无匹配返回 null
function findLine(re) {
  for (let i = 0; i < lines.length; i++) {
    if (re.test(lines[i])) return i + 1;
  }
  return null;
}

// 提取 [startLine, endLine]（1-based，含两端）
function extract(start, end) {
  const seg = lines.slice(start - 1, end);
  return seg.join('\n') + (seg.length ? '\n' : '');
}

function write(name, content) {
  const p = path.join(docsDir, name);
  fs.writeFileSync(p, content, 'utf8');
  console.log(`WROTE ${name} (${content.split('\n').length - 1} lines)`);
}

// ---- 锚点（标题行号）----
const h2 = findLine(/^## 2\. /);
const h3 = findLine(/^## 3\. /);
const h4 = findLine(/^## 4\. /);
const h5 = findLine(/^## 5\. /);
const h6 = findLine(/^## 6\. /);
const h7 = findLine(/^## 7\. /);
const h8 = findLine(/^## 8\. /);
const h9 = findLine(/^## 9\. 批次 27/);
const h9b = findLine(/^## 9b\. /);
const h10 = findLine(/^## 10\. /);
const h11 = findLine(/^## 11\. /);
const h12 = findLine(/^## 12\. /);
const hDisciplineInline = findLine(/^### 纪律约束/);

console.log({ h2, h3, h4, h5, h6, h7, h8, h9, h9b, h10, h11, h12, hDisciplineInline });

// ---- ENV.md：§2 目标物 + §3 环境 + §7 资产 ----
const envHead = '# 环境与资产（UsbEAm Launcher 1.0.3 逆向还原）\n\n> 从 `HANDOFF.md` 拆出。目标物、锁定构建参数、关键环境事实、反汇编资产清单的唯一真相源。\n\n';
write('ENV.md', envHead + extract(h2, h3 - 1) + extract(h3, h4 - 1) + extract(h7, h8 - 1));

// ---- DISCIPLINE.md：§4 纪律 + §6 内联纪律 + §9b 实证错误记录 ----
const discHead = '# 还原纪律与实证错误记录\n\n> 从 `HANDOFF.md` 拆出。逐寄存器还原纪律 + 累计 77 处实证错误纠正的唯一真相源。每批开工前必读。\n\n';
write('DISCIPLINE.md', discHead + extract(h4, h5 - 1) + extract(hDisciplineInline, findLine(/^### 开工命令模板/) - 1) + extract(h9b, h9b !== null ? (findLine(/^### 批次 33/) - 1) : h9b));

// ---- UNLANDED.md：§10 未落地清单 + §11 产物状态 ----
const unlHead = '# 未落地原始文件清单\n\n> 从 `HANDOFF.md` 拆出。蓝图 `^File:` 清单与 backend 非测试文件差集的唯一真相源。开工前请重跑 §0 命令。\n\n';
write('UNLANDED.md', unlHead + extract(h10, h12 - 1));

// ---- acceptance/batch19-39.md：批次 19-39 详细记录归档（缺 acceptance 的唯一 SSOT）----
const bHead = '# 批次 19–39 详细记录（历史归档）\n\n> 从 `HANDOFF.md` 拆出。批次 19–39 无独立 acceptance 文档，本文件是它们详细记录的唯一真相源。\n> 批次 40 起详见各自的 `docs/acceptance/batchNN.md`。\n\n';
const b19_39 =
  extract(h8, h9b - 1) +           // §8 批次 29-32 + 19-25（含 §9 批次 27，因 §9 嵌于 §8 区间）
  extract(findLine(/^### 批次 33/), h10 - 1); // 批次 33-39
write('acceptance/batch19-39.md', bHead + b19_39);

console.log('DONE');
