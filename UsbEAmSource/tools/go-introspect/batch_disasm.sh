#!/usr/bin/env bash
# batch_disasm.sh — 成批 capstone 标注反汇编已 dump 的函数 bin，归档到 outdir。
# 研究用途。依赖同目录 va_disasm.py（capstone + symbols.txt 标注）。
#
# 用法: batch_disasm.sh <dumpdir> <outdir> <VA-hex|bins> [更多 ...]
#   每个参数形如 "VA,basename"（VA 为 0x 十六进制，basename 无需扩展名，bin 位于 dumpdir/<basename>.bin）
set -euo pipefail

DUMP="${1:?dumpdir}"; OUT="${2:?outdir}"; shift 2
TD="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
mkdir -p "$OUT"
ok=0; miss=0
for item in "$@"; do
  VA="${item%%,*}"; NAME="${item#*,}"
  BIN="$DUMP/${NAME}.bin"
  if [ ! -f "$BIN" ]; then
    echo "MISS $NAME"; miss=$((miss+1)); continue
  fi
  if python "$TD/va_disasm.py" "$BIN" "$VA" 99999 "$OUT/${NAME}.asm.txt" 2>/dev/null \
     && [ -s "$OUT/${NAME}.asm.txt" ]; then
    echo "OK $NAME"
    ok=$((ok+1))
  else
    echo "DISFAIL $NAME"; miss=$((miss+1))
  fi
done
echo "batch: $ok ok, $miss fail -> $OUT"