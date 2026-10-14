#!/bin/bash
# [P] 存根清单 —— 薄封装。
#
# 标记判定的唯一真相源是 tools/count_funcs.sh 的 awk 状态机（docs/STRATEGY.md §4）。
# 本脚本不再自带一份实现，只转发 python tools/index/q.py landed。
# 输出已在签名前带上符号 VA 与实测字节数（原 tools/p_funcs.py 的 CSV 视图已并入）。
#
# 用量：bash tools/list_p.sh [marker]     # 默认 P
set -euo pipefail
cd "$(cd "$(dirname "$0")/.." && pwd)"
exec python tools/index/q.py landed "${1:-P}"
