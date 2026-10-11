const fs = require('fs');
const path = require('path');
const basedir = path.join(__dirname);

// list all P-marked functions from backend sources
const backendDir = path.join(basedir, 'backend');
const files = fs.readdirSync(backendDir).filter(f => f.endsWith('.go') && !f.endsWith('_test.go'));

const pFuncs = [];
for (const f of files) {
  const t = fs.readFileSync(path.join(backendDir, f), 'utf8');
  const lines = t.split(/\r?\n/);
  for (let i = 0; i < lines.length; i++) {
    if (!/^\s*\/\/\s*\[P\b/.test(lines[i])) continue;
    for (let j = i + 1; j < Math.min(i + 6, lines.length); j++) {
      const fm = lines[j].match(/^func\s+(\([^)]*\)\s*)?(\w+)/);
      if (!fm) continue;
      let name = fm[2];
      if (fm[1]) {
        const r = fm[1].trim().replace(/^\(|\)$/g, '').trim();
        const parts = r.split(/\s+/);
        const typeName = parts[parts.length - 1].replace(/^\*/, '');
        name = typeName + '.' + name;
      }
      // verify this func uses the [P] marker: check count_funcs would pick it up
      // find the last "//" comment line preceding the func
      let marker = '';
      for (let k = j - 1; k >= 0 && lines[k].startsWith('//'); k--) {
        const m = lines[k].match(/\[(S-sig|S-eq|S-inline|S|P)\b/);
        if (m) { marker = m[1]; break; }
      }
      if (marker === 'P') {
        pFuncs.push({ file: f, line: i + 1, name });
      }
      break;
    }
  }
}

// list asm files in pipeline ws*
const pipeDir = path.join(basedir, 'docs', 'goresym', 'pipeline', 'tmp');
const pipeFiles = fs.readdirSync(pipeDir).filter(f => f.endsWith('.asm.txt'));
const pipeMap = new Map();
for (const f of pipeFiles) {
  if (f.startsWith('ws')) {
    const stripped = f.substring(2, f.length - 8);
    pipeMap.set(stripped, f);
  }
}

console.log('P functions:', pFuncs.length);
console.log('Pipeline ws asm:', pipeMap.size);
console.log('');

let matched = 0;
for (const pf of pFuncs) {
  if (pipeMap.has(pf.name)) {
    console.log('MATCH:', pf.name, '->', pf.file + ':' + pf.line);
    matched++;
  }
}
console.log('Matched:', matched);