# 批次 270 · workspacemigration 域 12 函数落地：gate/排空/校验/暂存/提交/回滚（+12 [S]，-7 [S-sig]，+3 FUNCS）

## 目标

1. **落地 4 个未落地函数**：`validateWorkspaceMigrationRoots`（0x1409efac0）、
   `workspacePathContains`（0x1409f0160）、`waitWorkspaceMigrationDrain`（0x1409ef500）、
   `removeOwnedWorkspaceMigrationDirectory`（0x1409f34c0）。
2. **填充 3 个 [S-sig] 骨架**：`stageWorkspaceDataMigration`（0x1409f0240）、
   `stagedWorkspaceData.Commit`（0x1409f2460）、`stagedWorkspaceData.Rollback`（0x1409f3100）。
3. **订正 gate 签名与实现**：`workspaceDataMaintenanceGate.begin`（0x1409ef000，返回
   `(func(), error)` 退出闭包）/ `enter`（0x1409ef2a0，返回 `(chan struct{}, func(), error)`）；
   删除不存在的独立 `End` 方法（其体在 begin.func2.1 0x1409ef1e0）。
4. **订正被调方签名**：`beginWorkspaceDataOperation`（0x1409ef7c0，透明透传
   `(func(), error)`）、`beginWorkspaceMigrationMaintenance`（0x1409ef840，0 参 +
   WithTimeout(15s)→waitWorkspaceMigrationDrain）、`prepareWorkspaceDirectories`
   （0x1409f3760，6 目录 + 逆序清理闭包）。
5. **订正 7 处 defer 调用点**：`defer op.End()`/`defer gate.End()` → `defer op()`/`defer gate()`；
   `MigrateConfig` 两处调用点（begin 0 参、buildLayout 用 staged.targetPath）。

## 基线 / 收口

| 指标 | 基线（batch 269 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2868 | 2871 |
| MARKED | 2868 | 2871 |
| S | 1339 | 1349 |
| S-inline | 36 | 36 |
| S-sig | 1453 | 1446 |
| P | 40 | 40 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1375 | 1385 |
| USABLE | 1375 | 1385 |

`go build/vet/test -tags production ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；
`go1.25.12 test ./backend` `ok changeme/backend`。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1385  FUNCS=2871  MARKED=2871  P=40  S-eq=0  S-inline=36  S-sig=1446  S=1349  USABLE=1385
```

分项自洽：`S + S-inline + S-sig + P = 1349 + 36 + 1446 + 40 = 2871 = FUNCS`。
S 1339→1349（+10）、S-sig 1453→1446（-7）、FUNCS 2868→2871（+3）、P=40 持平、UNMARKED=0 保持。

账目：新增 `workspacemigration.go` 12 [S] + `prepareWorkspaceDirectories` [S-sig]→[S]；
删除 9 个骨架（bootstrapservice_migration.go 6 个 + screenshot_stubs.go 3 个，含 3 [S] + 6 [S-sig]）。

## G3 行为（asm 逐地址实证）

### 3.1 workspaceDataMaintenanceGate.begin [S 0x1409ef000]（gate_begin_end.asm.txt）

签名 `(func(), error)`。nil→`(workspaceMaintenanceNoop, nil)`；lock→maintenance 真→
`(nil, errWorkspaceMigrationInProgress)`；active==0→`g.idle=make(chan)`；active++→unlock；
`var once sync.Once` + 返回 `func(){ once.Do(func2.1) }`。func2.1（0x1409ef1e0）=
`lock; if active>0 {active--}; if active==0 && idle!=nil {close(idle); idle=nil}; unlock`。

### 3.2 workspaceDataMaintenanceGate.enter [S 0x1409ef2a0]（gateEnterExit.asm.txt）

签名 `(chan struct{}, func(), error)`。nil→`(closedChan, noop, nil)`；maintenance 真→
`(nil,nil,errWorkspaceMigrationInProgress)`；maintenance=true；`idleChan = active==0 ?
make+close(新chan) : g.idle`（**不回写 g.idle**，0x1409ef34f）；返回 `func(){ once.Do(func2.1) }`。
func2.1（0x1409ef480）= `lock; maintenance=false; unlock`。

### 3.3 waitWorkspaceMigrationDrain [S 0x1409ef500]

`idle==nil→nil`；`select { <-ctx.Done(): errors.Join(errWorkspaceDrainTimeout, ctx.Err());
<-idle: nil }`。errors.Join 内联实证：err[0].itab=0x1411d3100（extractError）、
data=0x141bc4050（"等待旧数据目录写入排空超时"，39B）。

### 3.4 beginWorkspaceDataOperation [S 0x1409ef7c0]

bs==nil→`(nil, newExtractError("工作区服务不可用"))`（@0x140C65700，24B）；否则
`(&bs.workspaceDataMaintenance).begin()` 原样透传 `(func(), error)`。

### 3.5 beginWorkspaceMigrationMaintenance [S 0x1409ef840]（0 参）

bs==nil→`(nil, "工作区服务不可用")`；lock(+0x540)→app(+0x188)!=nil→
`(nil, errWorkspaceAppActive)`；`enter()`→err 透传；`context.WithTimeout(Background,
15*time.Second)`（0x1409ef933 `movabs rcx,0x37e11d600`=15e9ns）→defer cancel()→
`waitWorkspaceMigrationDrain(ctx, idleChan)` err 时 `exitFn()` 后返 err；成功返 exitFn。

### 3.6 validateWorkspaceMigrationRoots [S 0x1409efac0]

`(source,target string)(string,string,error)`。TrimSpace 空→"源数据目录不能为空"(27)/
"目标数据目录不能为空"(30)；Abs 双→Clean；`workspacePathsEqual`→"不能相同"(42)；
`workspacePathContains` 双向→"不能互为父子目录"(54)；inspect 源 err/!exists→
"源数据目录不存在"(24)；inspect 目标 err；最终路径三向重叠
（equal || contains(src,dst) || contains(dst,src)）→"最终路径重叠"(51)；
目标存在且 identity.valid/volume/file 全等→"指向同一文件对象"(54)；否则 (cleanSource,cleanTarget,nil)。

### 3.7 workspacePathContains [S 0x1409f0160]

`Rel(a,b)` err→false；`"."`→false；`".."`→false；`rel[:3]=="..\\"`→false；否则 `!IsAbs(rel)`。

### 3.8 stageWorkspaceDataMigration [S 0x1409f0240]

validate→Stat 源 err/!IsDir→"源数据路径不是目录"(27)；inspect 源/目标；目标 exists→Stat/
!IsDir→"目标数据路径已存在且不是目录"(42)、ReadDir err、len>0→"目标数据目录必须为空"(30)；
MkdirAll(Dir(target),0755)→re-inspect target→inspect parent→!exists→"目标数据目录父级不存在"(33)；
rand 8B→stagingName=`.{Base}.usbeam-staging-{hex}`→Mkdir 0700→inspect→卷不等→
"迁移暂存目录与目标目录不在同一卷"(48)；Walk 拷贝闭包（copyWorkspaceDataWithManifest.func1
0x1409f1400）；verify(staging,false)→verify(source,true) 失败→
`fmt.Errorf("迁移期间源数据发生变化: %w", err)`；re-inspect source→!same→
"迁移期间源数据目录身份发生变化"(45)；re-inspect parent→身份不等→
"迁移期间目标父目录身份发生变化"(45)；返回 staged。

拷贝闭包（0x1409f1400）关键：Rel err→err；walkErr→return（**Rel 在前**，0x1409f1482/0x1409f1487）；
`rel!="." && shouldSkip`→IsDir?SkipDir:nil（无 strict）；`Join(staging,rel)`→isReparse err/
"数据目录包含不支持的符号链接、目录联接或 reparse point: %s"；IsDir→MkdirAll(dst,Mode().Perm())；
`!IsRegular`→"数据目录包含不支持的文件类型: %s"；MkdirAll(Dir(dst),0755)；openSource→Stat→
`statErr!=nil || !os.SameFile(info,fi)`→{Close; statErr?statErr:"迁移源文件身份发生变化: %s"}；
openTarget(dst, Mode().Perm())（**perm 被 0x1409f4e93 覆盖为 GENERIC_WRITE 忽略**）→
sha256→io.Copy(MultiWriter(target,h),src)→Sync→Close×2；错误序 copyErr→syncErr→
targetCloseErr→srcCloseErr；`manifest[ToSlash(rel)]={n,hex}`。

### 3.9 verifyWorkspaceDataManifestWithPolicy [S 0x1409f1ba0] + func1 [S 0x1409f1da0]

`visited=make(map[string]struct{},len(manifest))`（type 0x140b48aa0）→Walk(path,func1)→err；
`len(visited)!=len(manifest)`→"迁移目标文件清单不完整"(33)。func1 捕获
{root.ptr,len,strict,manifest,visited}；顺序：**walkErr 在前**（0x1409f1df2 test rsi）→
Rel→`strict && rel!="." && shouldSkip`→IsDir?SkipDir:nil→isReparse err/
"迁移清单路径包含 reparse point: %s"→IsDir→nil→`Mode()&fs.ModeType!=0`→
"迁移目标包含非普通文件: %s"→relKey=`strings.ReplaceAll(rel,"\\","/")`（0x5c→0x2f）→
清单外→"迁移目标出现清单外文件: %s"→openSource→sha256→io.Copy→Close→size 不等→
"迁移文件校验失败: %s"→sha 不等→同串→visited[relKey]。

### 3.10 stagedWorkspaceData.Commit [S 0x1409f2460]

nil→"迁移暂存状态不可用"(27)；verify(staging,false)→verify(source,true) 失败→
`fmt.Errorf("提交前源数据发生变化: %w", err)`；re-inspect source→!same→
"提交前源数据目录身份发生变化"(42)；inspect staging→!exists/valid/volume/file 不等→
"迁移暂存目录身份发生变化"(36)；inspect parent→!exists/身份不等→"目标父目录身份发生变化"(33)；
inspect target→!same→"目标数据目录身份发生变化"(36)；Stat(target)：err&&!Is(NotExist)→err，
err==nil&&!IsDir→"目标数据路径已存在且不是目录"(42)，ReadDir err→err，len>0→
"目标数据目录必须为空"(30)，Remove err→err，targetWasEmpty=true；Rename(staging,target) err→
{targetWasEmpty 时 MkdirAll(target,0755); return err}；committed=true；committedIdentity=stagingIdentity；
re-inspect target err→err，!exists/身份不等→"迁移提交后的目标目录身份不一致"(45)；
committedIdentity=新 identity；nil。

### 3.11 stagedWorkspaceData.Rollback [S 0x1409f3100]

nil→nil；!committed→removeOwned(stagingPath, stagingIdentity)；否则 inspect(target)：身份全等→
verify(target,false) err→`fmt.Errorf("迁移目标已被外部修改，拒绝删除: %w", err)`→RemoveAll err→err→
targetWasEmpty 时 MkdirAll(target,0755)→nil；inspectErr Is(NotExist)→nil；inspectErr!=nil→err；
!exists→nil；否则 "迁移目标身份已变化，拒绝删除"(42)。

### 3.12 removeOwnedWorkspaceMigrationDirectory [S 0x1409f34c0]

inspect→ErrNotExist→nil；!exists→nil；`!identity.valid || !inspection.identity.valid ||
volume!= || file!=`→"迁移暂存目录身份已变化，拒绝删除"(48)；否则 RemoveAll。

### 3.13 prepareWorkspaceDirectories [S 0x1409f3760]

6 目录 `{Root,IconDir,IndexDir,ScreenshotDir,WebView2Dir,BackgroundDir}`（栈偏移
0x00/0x50/0x60/0x70/0x80/0x90）；`created=make([]string,0,6)`；cleanup(func1 0x1409f3b80)=
`for i:=len(created)-1;i>=0;i--{os.Remove(created[i])}`；每目录 TrimSpace 空→continue；
os.Stat 非 ErrNotExist err→cleanup()+return(nil,err)；ErrNotExist→append；MkdirAll(trimmed,0755)
err→cleanup()+return(nil,err)；return(cleanup,nil)。

## 错误类型与全局身份实证

包级 error 接口（asm 从 .data 全局加载 itab+data）：`errWorkspaceMigrationInProgress`
（0x141BC3E00，"数据目录迁移正在进行" 30B）、`errWorkspaceAppActive`（0x141BC3E10，
"WebView2 数据目录正在使用中，请退出应用后执行离线迁移" 75B）、`errWorkspaceDrainTimeout`
（0x141BC4050，"等待旧数据目录写入排空超时" 39B）。"工作区服务不可用" 为每次
`newExtractError`（newobject，24B，@0x140C65700），非全局。

## G4 独立复核

新增 `backend/workspacemigration.go`（12 [S]）；`bootstrapservice_migration.go` 删 6 骨架、
保留 buildWorkspaceLayoutWithConfig/applyWorkspaceLayout/importLauncherBackgroundImage；
`screenshot_stubs.go` 删 beginWorkspaceDataOperation/begin/End，保留截图辅助三函数；
`workspacemigration_identity_windows.go` 订正 `openWorkspaceMigrationTargetFileNoFollow`
签名补 perm 参数（asm 0x1409f4e93 忽略 perm）；`bootstrapservice_commitconfig.go`
prepareWorkspaceDirectories 升 [S]；`bootstrapservice.go` 订正 MigrateConfig 两调用点 +
7 处 defer。全量 build/vet/test 回归 PASS，无新增测试。

## 移交（本轮收尾）

workspace migration 域主体（gate/排空/校验/暂存/提交/回滚/根校验/身份删除）**全部落地 [S]**。
本域剩余骨架：`buildWorkspaceLayoutWithConfig`（0x1407a1ac0）、`importLauncherBackgroundImage`
（0x140798e40）仍 [S-sig]（非本批范围）。P=40 持平。FUNCS 2871/4754 = 60.39%。
