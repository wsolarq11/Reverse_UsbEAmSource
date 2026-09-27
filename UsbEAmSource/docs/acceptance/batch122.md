# 批次 122 验收（启动权限归一化 + 拖拽文件入口构建）

日期：2026-09-23
子批次：launch-privilege + dropped-files-entry

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok (0.424s) |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1209 / MARKED=1209 / S=895 / S-inline=34 / S-sig=279 / P=1 / UNMARKED=0`。

相对批次 121（1207/893/34/279/1）：FUNCS +2，S +2。

## G3 逻辑等价

`backend/launchprivilege.go` + `backend/launcherprocess.go`：

- **normalizeAppLaunchPrivilegeMode**（0x140882680, 512B）修正 [P]→[S]：补全枚举。
  实测映射（len 跳转表 @0x140c37da0，len<5/>15 → ""）：
  admin/runas/administrator → "admin"；standard/normal/unelevated → "standard"；
  follow/launcher/followlauncher/follow-launcher → "followLauncher"；default → "default"；
  空/未知 → ""。修正了旧版「default/空 → default」的错误（空实为 ""）。
- **resolveEffectiveAppLaunchPrivilege**（0x140882880, 416B）[S]：normalize(configured) 为
  admin/standard/followLauncher → 返回；否则 normalize(fallback) 为三者之一 → 返回；
  否则默认 "standard"。
- **buildAppEntryWithDroppedFiles**（0x1408a6c20, 1024B）[S]：directory → 原样；
  normalizePathList(droppedFiles) 空 → 原样；否则 resolveLaunchableAppEntryForExplicitArgs →
  args = cleanStringList(append(resolved.args, paths...))（growslice + typedslicecopy 实证）。

## G4 复验

`TestNormalizeAppLaunchPrivilegeMode` 覆盖 13 枚举 + 空/未知；
`TestResolveEffectiveAppLaunchPrivilege` 覆盖 configured/fallback/默认三路径；
`TestBuildAppEntryWithDroppedFiles*` 覆盖 directory/空/合并三路径。全绿。

## 判定

四路 PASS，批次 122 闭环。normalizeAppLaunchPrivilegeMode 枚举补齐（[P]→[S]），
resolveEffectiveAppLaunchPrivilege 与 buildAppEntryWithDroppedFiles 100% asm 直译落 [S]。
