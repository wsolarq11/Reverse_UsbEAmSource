const fs = require('fs');
const sym = fs.readFileSync('docs/goresym/symbols.txt', 'utf8').split(/\r?\n/);
const addrs = {};
for (const l of sym) {
  const m = l.match(/^(0x[0-9a-fA-F]+)\s+(.+)$/);
  if (m) addrs[m[2]] = parseInt(m[1], 16);
}
const names = Object.keys(addrs);
const sorted = names.map(n => ({ n, a: addrs[n] })).sort((x, y) => x.a - y.a);
const sz = {};
for (let i = 0; i < sorted.length; i++) {
  const next = i + 1 < sorted.length ? sorted[i + 1].a : addrs[sorted[i].n] + 0x200;
  sz[sorted[i].n] = next - sorted[i].a;
}
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
const missing = [];
for (const n of names) {
  if (!n.startsWith('main.')) continue;
  const plain = normSym(n);
  if (isGen(plain)) continue;
  const base = plain.split('.');
  const fn = base[base.length - 1];
  const rec = base.length > 1 ? base.slice(0, -1).join('.') : '';
  const key = rec ? rec + '.' + fn : fn;
  if (funcs.has(key) || funcs.has(fn)) continue;
  missing.push({ plain, sz: sz[n], a: addrs[n], key });
}
missing.sort((x, y) => x.sz - y.sz);
console.log('未落地普通函数总数:', missing.length);
console.log('=== 最小 40 个 ===');
missing.slice(0, 40).forEach(m => console.log('  ' + String(m.sz).padStart(5) + 'B  ' + '0x' + m.a.toString(16).padStart(10) + '  ' + m.plain));
