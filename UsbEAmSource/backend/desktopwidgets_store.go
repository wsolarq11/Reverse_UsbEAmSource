package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
)

// desktopWidgetStoreError 构造桌面组件存储错误。
// [S 汇编 0x1407c4f20]：TrimSpace(op)、TrimSpace(path) → fmt.Errorf("%s: %s: %s"@0x140c440c7,
// "DESKTOP_WIDGET_STORE_INVALID"@0x140c6b920,28B, op, path)。
func desktopWidgetStoreError(op, path string) error {
	return fmt.Errorf("%s: %s: %s", "DESKTOP_WIDGET_STORE_INVALID", strings.TrimSpace(op), strings.TrimSpace(path))
}

// normalizeDesktopWidgetDocument 归一化整份桌面组件文档：版本缺省置 1、九个 map 字段缺省置空 map，
// 再遍历 Widgets 归一化每个 widget 的 Config、遍历 Reminders 归一化每个 reminder 的 ScheduleKind。
// [S 汇编 0x1407bfec0, 1376B] 实证：
//
//	Version==0→1（[rsp+0x2c8]）；九个 map 字段 nil→makemap_small（Widgets@0x2e8 … ProtectedSecrets@0x328）；
//	mapIterStart/mapIterNext 遍历 Widgets（map[string]DesktopWidget）→ widget.Config=
//	normalizeDesktopWidgetConfig(widget.Type, widget.Config) 后 mapassign 写回；
//	再遍历 Reminders（map[string]DesktopReminder）→ reminder.ScheduleKind=
//	normalizeDesktopReminderScheduleKind(reminder.ScheduleKind) 后 mapassign 写回；其余 map 仅初始化。
func normalizeDesktopWidgetDocument(doc DesktopWidgetDocument) DesktopWidgetDocument {
	if doc.Version == 0 {
		doc.Version = 1
	}
	if doc.Widgets == nil {
		doc.Widgets = make(map[string]DesktopWidget)
	}
	if doc.Notes == nil {
		doc.Notes = make(map[string]DesktopNote)
	}
	if doc.Reminders == nil {
		doc.Reminders = make(map[string]DesktopReminder)
	}
	if doc.Timers == nil {
		doc.Timers = make(map[string]DesktopTimer)
	}
	if doc.Stopwatches == nil {
		doc.Stopwatches = make(map[string]DesktopStopwatch)
	}
	if doc.StopwatchLaps == nil {
		doc.StopwatchLaps = make(map[string][]DesktopStopwatchLap)
	}
	if doc.WeatherCache == nil {
		doc.WeatherCache = make(map[string]DesktopWeatherSnapshot)
	}
	if doc.NotificationDeliveries == nil {
		doc.NotificationDeliveries = make(map[string]DesktopNotificationRecord)
	}
	if doc.ProtectedSecrets == nil {
		doc.ProtectedSecrets = make(map[string]DesktopProtectedSecret)
	}
	for key, widget := range doc.Widgets {
		widget.Config = normalizeDesktopWidgetConfig(widget.Type, widget.Config)
		doc.Widgets[key] = widget
	}
	for key, reminder := range doc.Reminders {
		reminder.ScheduleKind = normalizeDesktopReminderScheduleKind(reminder.ScheduleKind)
		doc.Reminders[key] = reminder
	}
	return doc
}

// readDesktopWidgetStoreBytes 读取组件存储文件字节，上限 32MB（0x2000000）。
// [S 汇编 0x1407c4280, 1200B] 实证：
//
//	TrimSpace(path) 空→errors.New("首页组件存储路径不能为空"@0x140c782eb,36B)；
//	os.OpenFile(path, O_RDONLY=0, 0) 失败→err；defer f.Close()；
//	f.Stat() 失败→err；info.Mode()&os.ModeType(test eax,0x8f280000)!=0→
//	desktopWidgetStoreError("路径不是普通文件"@0x140c64e90,24B, path)；
//	info.Size()>0x2000000→desktopWidgetStoreError(fmt.Sprintf("文件超过 %d 字节"@0x140c611fa,22B,
//	0x2000000@0x1411cd548), path)；io.ReadAll(io.LimitReader(f, 0x2000001)) 失败→err；
//	len(data)>0x2000000→同 size 超限错误；否则返回 (data, nil)。
func readDesktopWidgetStoreBytes(path string) ([]byte, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, errors.New("首页组件存储路径不能为空")
	}
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeType != 0 {
		return nil, desktopWidgetStoreError("路径不是普通文件", path)
	}
	if info.Size() > 0x2000000 {
		return nil, desktopWidgetStoreError(fmt.Sprintf("文件超过 %d 字节", 0x2000000), path)
	}
	data, err := io.ReadAll(io.LimitReader(f, 0x2000001))
	if err != nil {
		return nil, err
	}
	if len(data) > 0x2000000 {
		return nil, desktopWidgetStoreError(fmt.Sprintf("文件超过 %d 字节", 0x2000000), path)
	}
	return data, nil
}

// ensureDesktopWidgetJSONEOF 确保 JSON 解码器已读到文档末尾：再解码一次应为 io.EOF。
// [S 汇编 0x1407c4e20, 256B] 实证：
//
//	dec.Decode(&interface{} /*type size=0x10 @0x140b1a580*/)；
//	errors.Is(err, io.EOF)→nil；err==nil→errors.New("JSON 包含多个顶层值"@0x140c6829f,26B)；
//	否则 fmt.Errorf("JSON 尾部数据无效: %w"@0x140c69cf0,27B, err)。
func ensureDesktopWidgetJSONEOF(dec *json.Decoder) error {
	var v interface{}
	err := dec.Decode(&v)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err == nil {
		return errors.New("JSON 包含多个顶层值")
	}
	return fmt.Errorf("JSON 尾部数据无效: %w", err)
}

// cloneDesktopWidgetDocument 深拷贝一份组件文档：先归一化，再经 JSON 序列化/反序列化，失败回退零值文档。
// [S 汇编 0x1407c34e0, 1120B] 实证：
//
//	doc=normalizeDesktopWidgetDocument(doc)；json.Marshal(doc)（convT 转 interface{}）失败→
//	normalizeDesktopWidgetDocument(零值)；json.Unmarshal(data,&out) 失败→同零值回退；
//	成功→normalizeDesktopWidgetDocument(out)。
func cloneDesktopWidgetDocument(doc DesktopWidgetDocument) DesktopWidgetDocument {
	doc = normalizeDesktopWidgetDocument(doc)
	data, err := json.Marshal(doc)
	if err != nil {
		return normalizeDesktopWidgetDocument(DesktopWidgetDocument{})
	}
	var out DesktopWidgetDocument
	if err := json.Unmarshal(data, &out); err != nil {
		return normalizeDesktopWidgetDocument(DesktopWidgetDocument{})
	}
	return normalizeDesktopWidgetDocument(out)
}

// validateDesktopWidgetDocument 校验组件文档（版本/修订号/组件/便签/计次/整体序列化大小）。
// [S 汇编 0x1407c3940, 2368B] 实证：
//   Version!=1→("version","不支持的版本 %d"@0x140c5f674,21B,Version)；
//   Revision<0→("revision","不能为负数"@0x140c52cce,15B)；
//   len(Widgets)>200→("widgets","组件数量超过 %d"@0x140c5f689,21B,200)；
//   Widgets 遍历：key!=w.ID→("widgets","组件键与 id 不一致"@0x140c66a8d,25B)；
//   TrimSpace(w.Type) 白名单 {note,clock,timer,weather,calendar,reminder,stopwatch,worldClock} 否则
//   ("widgets["+key+"]","组件类型无效"@0x140c59ce4,18B)；
//   utf8.RuneCountInString(w.Title)>160(0xa0) 或 w.Revision<=0→("widgets["+key+"]",
//   "标题或 revision 无效"@0x140c66aa6,25B)；json.Marshal(w.Config) 失败或 len>0x10000→
//   ("widgets["+key+"]","组件配置超过限制"@0x140c64e60,24B)；
//   Notes 遍历：key!=note.WidgetID 或 len(note.Body)>0x40000 或 RuneCountInString(note.Body)>0x10000→
//   ("notes["+key+"]","便签身份或正文长度无效"@0x140c739b1,33B)；累计 len(Body)>0x800000→
//   ("notes","便签正文总量超过限制"@0x140c6ee83,30B)；
//   StopwatchLaps 遍历：len(laps)>0x3e8→("stopwatchLaps["+key+"]","计次数量超过限制"@0x140c64e78,24B)；
//   累计 len(laps)>0x2710→("stopwatchLaps","全局计次数量超过限制"@0x140c6eea1,30B)；
//   终 json.Marshal(doc) 失败→返回原始 err；len(data)+1>0x2000000→("document",
//   "序列化后超过 %d 字节"@0x140c6ba00,28B,0x2000000)；否则 validateLauncherConfigJSONStructure(data)。
func validateDesktopWidgetDocument(doc DesktopWidgetDocument) error {
	if doc.Version != 1 {
		return desktopWidgetStoreError("version", fmt.Sprintf("不支持的版本 %d", doc.Version))
	}
	if doc.Revision < 0 {
		return desktopWidgetStoreError("revision", "不能为负数")
	}
	if len(doc.Widgets) > 200 {
		return desktopWidgetStoreError("widgets", fmt.Sprintf("组件数量超过 %d", 200))
	}
	for key, w := range doc.Widgets {
		if key != w.ID {
			return desktopWidgetStoreError("widgets", "组件键与 id 不一致")
		}
		switch strings.TrimSpace(w.Type) {
		case "note", "clock", "timer", "weather", "calendar", "reminder", "stopwatch", "worldClock":
		default:
			return desktopWidgetStoreError("widgets["+key+"]", "组件类型无效")
		}
		if utf8.RuneCountInString(w.Title) > 160 || w.Revision <= 0 {
			return desktopWidgetStoreError("widgets["+key+"]", "标题或 revision 无效")
		}
		configData, err := json.Marshal(w.Config)
		if err != nil || len(configData) > 0x10000 {
			return desktopWidgetStoreError("widgets["+key+"]", "组件配置超过限制")
		}
	}
	var cumNotes int
	for key, note := range doc.Notes {
		if key != note.WidgetID || len(note.Body) > 0x40000 || utf8.RuneCountInString(note.Body) > 0x10000 {
			return desktopWidgetStoreError("notes["+key+"]", "便签身份或正文长度无效")
		}
		cumNotes += len(note.Body)
	}
	if cumNotes > 0x800000 {
		return desktopWidgetStoreError("notes", "便签正文总量超过限制")
	}
	var cumLaps int
	for key, laps := range doc.StopwatchLaps {
		if len(laps) > 0x3e8 {
			return desktopWidgetStoreError("stopwatchLaps["+key+"]", "计次数量超过限制")
		}
		cumLaps += len(laps)
	}
	if cumLaps > 0x2710 {
		return desktopWidgetStoreError("stopwatchLaps", "全局计次数量超过限制")
	}
	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	if len(data)+1 > 0x2000000 {
		return desktopWidgetStoreError("document", fmt.Sprintf("序列化后超过 %d 字节", 0x2000000))
	}
	return validateLauncherConfigJSONStructure(data)
}

// loadUnlocked 在持锁状态下加载组件文档：读文件→结构校验→JSON 解析→末尾校验→归一化→文档校验→缓存。
// [S 汇编 0x1407c2440, 3586B] 实证（返回 (bool,DesktopWidgetDocument,error)，错误分支返回 defaultDoc）：
//
//	defaultDoc=normalizeDesktopWidgetDocument(零值)；TrimSpace(path) 空→(false,defaultDoc,
//	errors.New("首页组件存储路径不能为空"@0x140c782eb,36B))；s.loaded→(s.cachedExists,clone(s.cached),nil)；
//	s.cachedLoadErr!=nil→(false,defaultDoc,s.cachedLoadErr)；readDesktopWidgetStoreBytes(path)；
//	errors.Is(err,os.ErrNotExist)→缓存 defaultDoc、loaded=true、cachedExists=false、
//	(false,clone(defaultDoc),nil)；err!=nil→cachedLoadErr=err、(false,defaultDoc,err)；
//	validateLauncherConfigJSONStructure(data) 失败→cachedLoadErr=desktopWidgetStoreError(s.path,err.Error())；
//	json.NewDecoder(bytes.NewReader(data))→Decode(&doc) 失败→cachedLoadErr=
//	desktopWidgetStoreError(s.path,fmt.Sprintf("解析 JSON 失败: %v"@0x140c611e4,22B,err))；
//	ensureDesktopWidgetJSONEOF(dec) 失败→cachedLoadErr=desktopWidgetStoreError(s.path,err.Error())；
//	doc=normalizeDesktopWidgetDocument(doc)；validateDesktopWidgetDocument(doc) 失败→cachedLoadErr=err；
//	成功→s.cached=clone(doc)、cachedExists=true、cachedLoadErr=nil、loaded=true、(true,clone(doc),nil)。
func (s *launcherWidgetStore) loadUnlocked() (bool, DesktopWidgetDocument, error) {
	defaultDoc := normalizeDesktopWidgetDocument(DesktopWidgetDocument{})
	if strings.TrimSpace(s.path) == "" {
		return false, defaultDoc, errors.New("首页组件存储路径不能为空")
	}
	if s.loaded {
		return s.cachedExists, cloneDesktopWidgetDocument(s.cached), nil
	}
	if s.cachedLoadErr != nil {
		return false, defaultDoc, s.cachedLoadErr
	}
	data, err := readDesktopWidgetStoreBytes(s.path)
	if errors.Is(err, os.ErrNotExist) {
		s.cached = defaultDoc
		s.loaded = true
		s.cachedExists = false
		return false, cloneDesktopWidgetDocument(defaultDoc), nil
	}
	if err != nil {
		s.cachedLoadErr = err
		return false, defaultDoc, err
	}
	if err := validateLauncherConfigJSONStructure(data); err != nil {
		s.cachedLoadErr = desktopWidgetStoreError(s.path, err.Error())
		return false, defaultDoc, s.cachedLoadErr
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	var doc DesktopWidgetDocument
	if err := dec.Decode(&doc); err != nil {
		s.cachedLoadErr = desktopWidgetStoreError(s.path, fmt.Sprintf("解析 JSON 失败: %v", err))
		return false, defaultDoc, s.cachedLoadErr
	}
	if err := ensureDesktopWidgetJSONEOF(dec); err != nil {
		s.cachedLoadErr = desktopWidgetStoreError(s.path, err.Error())
		return false, defaultDoc, s.cachedLoadErr
	}
	doc = normalizeDesktopWidgetDocument(doc)
	if err := validateDesktopWidgetDocument(doc); err != nil {
		s.cachedLoadErr = err
		return false, defaultDoc, s.cachedLoadErr
	}
	s.cached = cloneDesktopWidgetDocument(doc)
	s.cachedExists = true
	s.cachedLoadErr = nil
	s.loaded = true
	return true, cloneDesktopWidgetDocument(doc), nil
}

// Read 读取组件文档：加锁后交给 loadUnlocked，丢弃中间的文档副本只返回 (exists, err)。
// [S 汇编 0x1407c0420, 608B] 实证：s==nil→(false,errors.New("首页组件存储不可用"@0x140c69cd5,27B))；
// s.mu.Lock()（cmpxchg [s+0x10]）+ defer Unlock；exists,_,err=loadUnlocked() 后返回 (exists,err)。
func (s *launcherWidgetStore) Read() (bool, error) {
	if s == nil {
		return false, errors.New("首页组件存储不可用")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	exists, _, err := s.loadUnlocked()
	return exists, err
}

// writeDesktopWidgetDocumentFile 原子写组件文档到磁盘（临时文件 + rename + 重试）。
// [S 汇编 0x1407c47c0, 1504B] 实证：
//   path=TrimSpace(path)，空→errors.New("首页组件存储路径不能为空"@0x140c782eb,36B)；
//   doc=normalizeDesktopWidgetDocument(doc)；validateDesktopWidgetDocument(doc) 失败→返回该 err；
//   dir=filepath.Dir(path)；os.MkdirAll(dir,0x1ed=0o755) 失败→返回；
//   data=json.MarshalIndent(doc,"","  ")+'\n'（indent 2 空格@0x140c3366b，追加 0x0a 换行）；
//   tmp=os.CreateTemp(dir,".launcher-widgets-*.tmp"@0x140c62ea1,23B) 失败→返回；defer tmp.Close()；
//   tmp.Write(data) 失败→返回；tmp.Sync() 失败→返回；tmp.Close() 失败→返回；
//   rename 重试 ≤5 次：os.Rename(tmp.Name(),path) 成功→nil；失败且 errors.Is(err,os.ErrPermission)||
//   errors.Is(err,os.ErrExist) 则 time.Sleep(50ms=0x2faf080) 后 attempt++ 重试，否则返回 err；
//   5 次后仍失败返回最后的 rename err。
//   注：重试条件两 error 变量为 .data 段 @0x141c10970/0x141c10980，紧邻 os.ErrNotExist@0x141c10990
//   （func1 内 errors.Is(os.ErrNotExist) 实证），按 io/fs 包 error 声明顺序
//   （ErrPermission→ErrExist→ErrNotExist）确定为 os.ErrPermission 与 os.ErrExist。
func writeDesktopWidgetDocumentFile(path string, doc DesktopWidgetDocument) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return errors.New("首页组件存储路径不能为空")
	}
	doc = normalizeDesktopWidgetDocument(doc)
	if err := validateDesktopWidgetDocument(doc); err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp, err := os.CreateTemp(filepath.Dir(path), ".launcher-widgets-*.tmp")
	if err != nil {
		return err
	}
	defer tmp.Close()
	if _, err := tmp.Write(data); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	var renameErr error
	for attempt := 0; attempt < 5; attempt++ {
		renameErr = os.Rename(tmp.Name(), path)
		if renameErr == nil {
			return nil
		}
		if !errors.Is(renameErr, os.ErrPermission) && !errors.Is(renameErr, os.ErrExist) {
			return renameErr
		}
		time.Sleep(50 * time.Millisecond)
	}
	return renameErr
}

// writeUnlocked 在持锁状态下写组件文档：归一化→校验→通过 writeDocument（默认 writeDesktopWidgetDocumentFile）
// 落盘→缓存。用于 Ensure/Update/ReplaceWithRollback 等写入路径。
// [S 汇编 0x1407c3260, 640B] 实证：
//   doc=normalizeDesktopWidgetDocument(doc)；validateDesktopWidgetDocument(doc) 失败→返回该 err；
//   w:=s.writeDocument（字段 @0x98），w==nil 时取默认 writeDesktopWidgetDocumentFile
//   （全局函数指针 @0x141096dd0=0x1407c47c0）；w(s.path,doc) 失败→返回该 err；
//   成功→s.cached=cloneDesktopWidgetDocument(doc)、cachedExists=true、cachedLoadErr=nil、
//   loaded=true、return nil。
func (s *launcherWidgetStore) writeUnlocked(doc DesktopWidgetDocument) error {
	doc = normalizeDesktopWidgetDocument(doc)
	if err := validateDesktopWidgetDocument(doc); err != nil {
		return err
	}
	w := s.writeDocument
	if w == nil {
		w = writeDesktopWidgetDocumentFile
	}
	if err := w(s.path, doc); err != nil {
		return err
	}
	s.cached = cloneDesktopWidgetDocument(doc)
	s.cachedExists = true
	s.cachedLoadErr = nil
	s.loaded = true
	return nil
}

// Ensure 确保组件文档存在：不存在则以零值文档为底、盖上当前 UTC 时间戳后写盘，返回文档副本。
// [S 汇编 0x1407c06e0, 1504B] 实证（返回 (DesktopWidgetDocument,error)）：
//   s==nil→(零值,errors.New("首页组件存储不可用"@0x140c69cd5,27B))；s.mu.Lock()+defer Unlock；
//   exists,doc,err:=loadUnlocked()；err!=nil→(零值,err)；exists→(doc,nil)；
//   否则 doc=normalizeDesktopWidgetDocument(零值)；doc.UpdatedAt=time.Now().UTC().
//   Format(time.RFC3339Nano="2006-01-02T15:04:05.999999999Z07:00"@0x140c7708d,35B)；
//   writeUnlocked(doc) 失败→(零值,err)；成功→(cloneDesktopWidgetDocument(doc),nil)。
func (s *launcherWidgetStore) Ensure() (DesktopWidgetDocument, error) {
	if s == nil {
		return DesktopWidgetDocument{}, errors.New("首页组件存储不可用")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	exists, doc, err := s.loadUnlocked()
	if err != nil {
		return DesktopWidgetDocument{}, err
	}
	if exists {
		return doc, nil
	}
	doc = normalizeDesktopWidgetDocument(DesktopWidgetDocument{})
	doc.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	if err := s.writeUnlocked(doc); err != nil {
		return DesktopWidgetDocument{}, err
	}
	return cloneDesktopWidgetDocument(doc), nil
}

// Update 在持锁状态下更新组件文档：深拷贝现有文档→可选的 mutate 回调改副本→Revision+1 并盖 UTC
// 时间戳→写盘→返回副本。mutate 为 nil 时跳过（仅提升版本号与时间戳）。
// [S 汇编 0x1407c1100, 1600B] 实证（返回 (DesktopWidgetDocument,error)）：
//   s==nil→(零值,errors.New("首页组件存储不可用"@0x140c69cd5,27B))；s.mu.Lock()+defer Unlock；
//   _,doc,err:=loadUnlocked()（exists 忽略）；err!=nil→(零值,err)；clone:=cloneDesktopWidgetDocument(doc)；
//   mutate!=nil 则 err=mutate(&clone)（回调指针 @[rsp+0x278]，接收堆上 clone 指针），err!=nil→(零值,err)；
//   clone.Revision=doc.Revision+1（inc）；clone.UpdatedAt=time.Now().UTC().Format(time.RFC3339Nano)；
//   writeUnlocked(clone) 失败→(零值,err)；成功→(cloneDesktopWidgetDocument(clone),nil)。
func (s *launcherWidgetStore) Update(mutate func(*DesktopWidgetDocument) error) (DesktopWidgetDocument, error) {
	if s == nil {
		return DesktopWidgetDocument{}, errors.New("首页组件存储不可用")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	_, doc, err := s.loadUnlocked()
	if err != nil {
		return DesktopWidgetDocument{}, err
	}
	clone := cloneDesktopWidgetDocument(doc)
	if mutate != nil {
		if err := mutate(&clone); err != nil {
			return DesktopWidgetDocument{}, err
		}
	}
	clone.Revision = doc.Revision + 1
	clone.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	if err := s.writeUnlocked(clone); err != nil {
		return DesktopWidgetDocument{}, err
	}
	return cloneDesktopWidgetDocument(clone), nil
}

// ReplaceWithRollback 用新文档整体替换（带回滚）：归一化+时间戳+校验新文档→直接 writeDocument 落盘→
// 缓存新文档→返回一个回滚闭包（调用时恢复旧文档或删除文件）。返回 (func() error, error)。
// [S 汇编 0x1407c17a0, 3136B + func1@0x1407c1e60] 实证：
//   s==nil→(nil,errors.New("首页组件存储不可用"@0x140c69cd5,27B))；s.mu.Lock()+defer Unlock；
//   exists,oldDoc,err:=loadUnlocked()；err!=nil 时 oldDoc 清零且 exists=false（不返回错误）；
//   newDoc=normalizeDesktopWidgetDocument(newDoc)；newDoc.UpdatedAt=time.Now().UTC().Format(RFC3339Nano)；
//   validateDesktopWidgetDocument(newDoc) 失败→(nil,err)；w:=s.writeDocument（nil 取默认）→
//   w(s.path,newDoc) 失败→(nil,err)；成功→s.cached=clone(newDoc)、cachedExists=true、cachedLoadErr=nil、
//   loaded=true；返回闭包（捕获 s/exists/w/oldDoc）：
//     闭包内加锁 defer 解锁；exists 则 w(s.path,oldDoc) 恢复（失败返回 err）→缓存 oldDoc、
//     cachedExists=true；否则 os.Remove(s.path)，err 非 nil 且 !errors.Is(err,os.ErrNotExist) 返回 err→
//     缓存 normalizeDesktopWidgetDocument(零值)、cachedExists=false；两条路径均 loaded=true、return nil。
func (s *launcherWidgetStore) ReplaceWithRollback(newDoc DesktopWidgetDocument) (func() error, error) {
	if s == nil {
		return nil, errors.New("首页组件存储不可用")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	exists, oldDoc, err := s.loadUnlocked()
	if err != nil {
		oldDoc = DesktopWidgetDocument{}
		exists = false
	}
	newDoc = normalizeDesktopWidgetDocument(newDoc)
	newDoc.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	if err := validateDesktopWidgetDocument(newDoc); err != nil {
		return nil, err
	}
	w := s.writeDocument
	if w == nil {
		w = writeDesktopWidgetDocumentFile
	}
	if err := w(s.path, newDoc); err != nil {
		return nil, err
	}
	s.cached = cloneDesktopWidgetDocument(newDoc)
	s.cachedExists = true
	s.cachedLoadErr = nil
	s.loaded = true
	return func() error {
		s.mu.Lock()
		defer s.mu.Unlock()
		if exists {
			if err := w(s.path, oldDoc); err != nil {
				return err
			}
			s.cached = cloneDesktopWidgetDocument(oldDoc)
			s.cachedExists = true
			s.cachedLoadErr = nil
			s.loaded = true
			return nil
		}
		if err := os.Remove(s.path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		s.cached = normalizeDesktopWidgetDocument(DesktopWidgetDocument{})
		s.cachedExists = false
		s.cachedLoadErr = nil
		s.loaded = true
		return nil
	}, nil
}
