# 批次 246 — normalizeLauncherUpdatePackageRoot 函数体订正（错误 [S] 还原修正）

## 基线 / 收口

| 指标 | 基线（批次 245 收口） | 收口（批次 246） |
|---|---|---|
| FUNCS | 2788 | 2788 |
| S | 1256 | 1256 |
| S-inline | 36 | 36 |
| S-sig | 1455 | 1455 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |

计数不变（该函数本已标记 [S] 计入），但**函数体从错误还原订正为 asm 直译**，修正功能等价性。

SHA256（`bash build.sh` 重建 `artifacts/UsbEAm_Launcher_rebuilt.exe`，21,321,728 B）
`d03e41f104dc2a5d2a6d67298fbe7c697584754ebd0c19a964a781dd3220839e`。

## 本批内容

`backend/launcherupdate_plan.go:332` `normalizeLauncherUpdatePackageRoot`：

- 旧（错误）：`TrimSpace → filepath.Clean → filepath.Abs`（标为 [S] 但逻辑与 asm 不符）。
- 新（asm 直译）：`strings.Replace(root,"\\","-",-1) → TrimSpace → 空/含"-"/"."/".." 返回空`。

## asm 证据（0x1408b6b60，完整翻译）

- morestack 序言存 2 槽（rax/rbx）＝`(root string) string`。
- `0x1408b6b94 call strings.Replace`：old=0x1411cac40(字节 0x5c `"\\"`, len=1)、
  new=0x140c3362f(字节 0x2d `"-"`, len=1)、n=-1。
- `0x1408b6b99 call strings.TrimSpace`。
- `0x1408b6ba0 test rbx`：len==0 → 返空；`0x1408b6bc0 call stringslite.Index(result,"-")`：含 `-` → 返空；
  len==1 且首字节 0x2e(`"."`) → 返空；len==2 且 word 0x2e2e(`".."`) → 返空；否则返 result。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build/vet ./backend` EXIT=0；`test ./backend` `ok changeme/backend`。
- **G2 count_funcs**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1455 / P=41 / UNMARKED=0`。
- **G3 行为**：纯函数体订正，无调用方变更，既有测试全量 PASS。
- **G4 review**：`launcherupdate_plan.go`（samePathFold/hashLauncherUpdateFile 等相邻 [S] 未受影响）。

## 遗留（下一批）

- P=41（oledblackout_windows 25 / screenshot_windows 7 / screenshot_uia_windows 6 / mousegestures 1 /
  nativedrag_windows 1 / oledblackout 1 ghost）。
- §10 差集 57 文件；FUNCS 缺口 697 个未落地顶层函数 + 闭包/方法。
