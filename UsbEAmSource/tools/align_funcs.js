// 函数名对齐 v4：symbols.txt vs backend/*.go，准确过滤生成函数
const fs = require('fs');
const path = require('path');

const symLines = fs.readFileSync('docs/goresym/symbols.txt', 'utf8').split(/\r?\n/);
const symFuncs = [];
for (const l of symLines) {
  const m = l.match(/^(0x[0-9a-fA-F]+)\s+main\.(.+)$/);
  if (!m) continue;
  symFuncs.push({ name: m[2], addr: parseInt(m[1], 16) });
}

// 生成函数识别（编译器产物，无法显式落地）
function isGenerated(name) {
  // funcN / funcN.M / gowrapN / deferwrapN / PrintffuncN / 末尾 .N 嵌套闭包
  return /\.?(func\d+|deferwrap\d+|gowrap\d+|Printffunc\d+)(\.\d+)*$/.test(name) ||
         /\.\d+$/.test(name);
}

function keyOf(name) {
  const m = name.match(/^\(\*([^)]+)\)\.?(.+)$/);
  if (m) return m[1] + '.' + m[2];
  const m2 = name.match(/^([A-Za-z_][A-Za-z0-9_]*)\.(.+)$/);
  if (m2) return m2[1] + '.' + m2[2];
  return name;
}

function backendKeys() {
  const keys = new Map();
  for (const f of fs.readdirSync('backend')) {
    if (!f.endsWith('.go') || f.endsWith('_test.go')) continue;
    for (const l of fs.readFileSync(path.join('backend', f), 'utf8').split(/\r?\n/)) {
      const m = l.match(/^func\s+(\([^)]*\)\s*)?([A-Za-z_][A-Za-z0-9_]*)\s*\(/);
      if (!m) continue;
      const recvRaw = (m[1] || '').trim();
      let key = m[2];
      if (recvRaw) {
        const rm = recvRaw.match(/\(([^)]+)\)/);
        if (rm) {
          const parts = rm[1].trim().split(/\s+/);
          key = parts[parts.length - 1].replace(/^\*/, '') + '.' + m[2];
        }
      }
      keys.set(key, f);
    }
  }
  return keys;
}

const bk = backendKeys();
const missing = [];
let matched = 0, gen = 0;
for (const s of symFuncs) {
  if (isGenerated(s.name)) { gen++; continue; }
  const key = keyOf(s.name);
  if (bk.has(key)) { matched++; continue; }
  missing.push(s);
}

function size(s) {
  let next = Infinity;
  for (const t of symFuncs) if (t.addr > s.addr && t.addr < next) next = t.addr;
  return next - s.addr;
}

console.log('symbols main funcs =', symFuncs.length);
console.log('generated          =', gen);
console.log('backend funcs      =', bk.size);
console.log('matched (plain)    =', matched);
console.log('missing (plain)    =', missing.length);

const sorted = missing.slice().sort((a, b) => size(a) - size(b));
console.log('');
console.log('=== 最小 50 个未落地普通函数 ===');
for (const s of sorted.slice(0, 50)) {
  console.log(s.name.padEnd(60) + size(s) + 'B  @0x' + s.addr.toString(16));
}
