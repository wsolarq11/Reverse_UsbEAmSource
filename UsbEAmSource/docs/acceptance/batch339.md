# 批次 339 · 输入消除保护/原生显示判定/观众关闭/预览边界 +4（FUNCS 3111）

## 目标

落地 4 个函数：`oledBlackoutService.inputDismissGuardActiveLocked`、
`screenshotPreviewWindowService.shouldDisplayNative`、`pluginWindowService.CloseAudience`、
`screenshotPreviewBounds`。

## 基线 / 收口

| 指标 | 基线（batch 338 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3107 | 3111 |
| MARKED | 3107 | 3111 |
| S | 1500 | 1500 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1528 | 1532 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1537 | 1537 |
| USABLE | 1538 | 1538 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1537  FUNCS=3111  MARKED=3111  P=41  S-eq=1  S-inline=37  S-sig=1532  S=1500  USABLE=1538
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1500 + 37 + 1 + 1532 + 41 = 3111 = FUNCS`。
S-sig 1528→1532（+4）、FUNCS 3107→3111（+4）、FAITHFUL/USABLE 保持、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 inputDismissGuardActiveLocked [S-sig 0x14090d200, 288B]

deadline(+0x180) 零→false；否则 now.Before(deadline)。

### 3.2 shouldDisplayNative [S-sig 0x1409998e0, 288B]

lock → 读字段(+0xa8/+0xb0/+0x28/+0xb1) 判定 → unlock。

### 3.3 CloseAudience [S-sig 0x140930080, 288B]

nil 早退；TrimSpace(id)；lock → windows(+0x10)[id] → unlock；win 非空且 audience(+0x88)
非空则调 close(+0x30)。

### 3.4 screenshotPreviewBounds [S-sig 0x14099b660, 288B]

ScreenManager.GetAll → 宽高减 x/y 再减 0x12 → clamp。

## G4 独立复核

- `backend/oledblackout.go`：+inputDismissGuardActiveLocked [S-sig]。
- `backend/screenshot_preview_windows.go`：+shouldDisplayNative [S-sig] +screenshotPreviewBounds [S-sig]。
- `backend/bootstrapservice_state_deps.go`：+CloseAudience [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（4 [S-sig]）。FUNCS 3111/4754 = 65.44%。下一批：remoteicons 域
validateRemoteIconCachePath / isRemoteIconSVG / readRemoteIconCache。
