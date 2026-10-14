#!/usr/bin/env node
// 未落地清单 —— 薄封装。
//
// 判定实现只有一处：tools/index/common.py（func_key / is_generated / is_never_land /
// classify_missing）。本文件仅转发到 `python tools/index/q.py missing`，不再自带逻辑。
// 历史上这里有一份独立的、与索引侧不一致的实现，对同一份代码给出不同的缺失数
// （短名回退把别的类型的同名方法算成本函数的实现）。保留本入口只为兼容
// docs/HANDOFF.md §0 里 `node tools/list_missing.js` 的引用。
const { spawnSync } = require('child_process');
const path = require('path');

const root = path.resolve(__dirname, '..');
const r = spawnSync('python', ['tools/index/q.py', 'missing', ...process.argv.slice(2)], {
  cwd: root,
  stdio: 'inherit',
});
process.exit(r.status === null ? 1 : r.status);
