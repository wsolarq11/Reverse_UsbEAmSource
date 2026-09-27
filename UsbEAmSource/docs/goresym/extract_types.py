txt = open('all_types.txt', encoding='utf-8', errors='replace').read()
lines = txt.split('\n')
blocks=[]; i=0
while i < len(lines):
    if lines[i].startswith('type main.'):
        start=i
        j=i+1; depth=lines[i].count('{')-lines[i].count('}')
        while j < len(lines) and depth>0:
            depth += lines[j].count('{')-lines[j].count('}')
            j+=1
        blocks.append('\n'.join(lines[start:min(j,len(lines))]))
        i=j
    else:
        i+=1
open('main_types_reconstructed.go','w',encoding='utf-8').write(
  'package main\n\n// AUTO-RECONSTRUCTED from target binary via redress (go1.25.12 pclntab)\n// 仅研究用途；字段与 json tag 取自目标 exe。\n\n'+'\n\n'.join(blocks)+'\n')
print("main type blocks:", len(blocks))
print("bytes:", len(open('main_types_reconstructed.go',encoding='utf-8').read()))
