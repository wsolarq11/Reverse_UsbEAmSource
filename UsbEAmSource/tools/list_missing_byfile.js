const fs = require('fs');
// 读取 symbols
const sym = fs.readFileSync('docs/goresym/symbols.txt', 'utf8').split(/\r?\n/);
const addrs = {};
const symToSrc = {}; // symbol -> source file
for (const l of sym) {
  const m = l.match(/^(0x[0-9a-fA-F]+)\s+(.+)$/);
  if (m) addrs[m[2]] = parseInt(m[1], 16);
}
// 读取 source_funcs.txt 建立 funcName -> source file
const sf = fs.readFileSync('docs/goresym/source_funcs.txt', 'utf8').split(/\r?\n/);
let cur = '';
const srcFileByFunc = {};
for (const l of sf) {
  if (l.startsWith('File: ')) {
    cur = l.replace(/^File: /, '').replace(/\t.*$/, '').trim();
  } else if (l.startsWith('\t') && cur) {
    const m = l.match(/^\t(.+?) Lines:/);
    if (m) srcFileByFunc[m[1].trim()] = cur;
  }
}
// backend 顶层函数
const glob = fs.readdirSync('backend').filter(f => f.endsWith('.go') && !f.endsWith('_test.go'));
const funcs = new Set();
for (const f of glob) {
  const t = fs.readFileSync('backend/' + f, 'utf8');
  for (const m of t.matchAll(/^func\s+(\([^)]*\)\s*)?(\w+)/gm)) {
    let n = m[2];
    if (m[1]) {
      const r = m[1].trim().replace(/^\(|\)$/g, '').trim();
      const parts = r.split(/\s+/);
      const typeName = parts[parts.length - 1].replace(/^\*/, '');
      n = typeName + '.' + n;
    }
    funcs.add(n);
  }
}
function isGen(n) {
  return /(func\d+|deferwrap\d+|gowrap\d+|Printffunc\d+)(\.\d+)*$/.test(n) || /\.\d+$/.test(n) || /-fm$/.test(n);
}
function normSym(n) {
  let x = n.replace(/^main\./, '');
  x = x.replace(/^\(\*([^)]+)\)\./, '$1.');
  x = x.replace(/^\(\*([^)]+)\)/, '$1');
  return x;
}
// 统计每个 source 文件的未落地函数
const fileMissing = {}; // source file -> count
for (const n of Object.keys(addrs)) {
  if (!n.startsWith('main.')) continue;
  const plain = normSym(n);
  if (isGen(plain)) continue;
  const base = plain.split('.');
  const fn = base[base.length - 1];
  const rec = base.length > 1 ? base.slice(0, -1).join('.') : '';
  const key = rec ? rec + '.' + fn : fn;
  if (funcs.has(key) || funcs.has(fn)) continue;
  // 找 source 文件
  let sf2 = srcFileByFunc[plain];
  if (!sf2) {
    // 尝试去掉 receiver
    sf2 = srcFileByFunc[fn];
  }
  if (!sf2) sf2 = '(unknown)';
  fileMissing[sf2] = (fileMissing[sf2] || 0) + 1;
}
const sorted = Object.entries(fileMissing).sort((a, b) => b[1] - a[1]);
console.log('=== 未落地函数按 source 文件分布（top 40）===');
sorted.slice(0, 40).forEach(([f, c]) => console.log('  ' + String(c).padStart(4) + '  ' + f));
console.log('total files with missing:', sorted.length);
