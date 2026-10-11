const fs = require('fs');

// Read count_funcs output
const glob = fs.readdirSync('backend').filter(f => f.endsWith('.go') && !f.endsWith('_test.go'));

const results = [];
for (const f of glob) {
  const t = fs.readFileSync('backend/' + f, 'utf8');
  // Find all // [P ...] comments followed by func stubs
  const lines = t.split(/\r?\n/);
  for (let i = 0; i < lines.length; i++) {
    const l = lines[i];
    const m = l.match(/^\s*\/\/\s*\[P\s*(?:0x([0-9a-fA-F]+))?.*\]/);
    if (!m) continue;
    const addr = m[1] ? '0x' + m[1] : '';
    // find next func declaration
    let funcLine = '';
    for (let j = i + 1; j < lines.length && j < i + 5; j++) {
      const fm = lines[j].match(/^func\s+(\([^)]*\)\s*)?\w+/);
      if (fm) { funcLine = lines[j]; break; }
    }
    results.push({ file: f, line: i + 1, addr, funcLine: funcLine.trim() });
  }
}

console.log('P count:', results.length);
for (const r of results) {
  console.log(r.file + ':' + r.line + '  ' + r.addr + '  ' + r.funcLine);
}