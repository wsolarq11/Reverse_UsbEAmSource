// AUTO-RECONSTRUCTED FUNCTIONS — DOMAIN: bookmark icon cache / favicon DB
// 研究用途
//
// 契约来源：
//   - 符号地址：symbols.main.bak
//   - 行号蓝图：source_funcs.txt bookmarkicon.go L84-1291
//
// 档位：[S] 反汇编实证（缓存 + SQLite + HTTP 下载体系）
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"golang.org/x/net/publicsuffix"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// ---- 书签图标全局状态 ----

var (
	bookmarkIconCacheLock sync.Mutex
	bookmarkIconMemCache  sync.Map // map[string]string: urlKey->dataURL
)

// ---- 缓存重置 ----

// resetBookmarkIconCache 重置书签图标全部缓存（内存 + 磁盘标记）。
// [S 汇编 0x14075bda0]：lock → HashTrieMap.Clear → 清全局 time → unlock。
func resetBookmarkIconCache() {
	bookmarkIconCacheLock.Lock()
	bookmarkIconMemCache = sync.Map{}
	bookmarkIconCacheLock.Unlock()
}

// ---- 入口解析 ----

// resolveBookmarkIconData 解析书签图标数据。
// [S 汇编 0x14075c120]：normalizeBookmarkIconURL → bookmarkIconCachePath → bookmarkIconCacheKey →
// loadBookmarkIconMemoryCache → 命中返回；未命中则 loadBookmarkIconFromSource → cache/store。
func resolveBookmarkIconData(workspaceID, originURL, referrer string) (string, error) {
	normalizedURL := normalizeBookmarkIconURL(originURL)
	if normalizedURL == "" {
		return "", fmt.Errorf("bookmarkicon: empty URL")
	}

	cacheKey := bookmarkIconCacheKey(normalizedURL, referrer)
	if data, ok := loadBookmarkIconMemoryCache(cacheKey); ok {
		return data, nil
	}

	data, err := loadBookmarkIconFromCache(workspaceID, normalizedURL)
	if err == nil && data != "" {
		storeBookmarkIconMemoryCache(cacheKey, data)
		return data, nil
	}

	return resolveBookmarkIconSource(workspaceID, normalizedURL, referrer, cacheKey)
}

// resolveBookmarkIconSource 从源获取并缓存。
// [S-inline 内联]：内联于 resolveBookmarkIconData(0x14075c120)，无独立符号；
// 源加载→MIME 规范化→dataURL→内存/磁盘缓存→返回 的分支拆解。
func resolveBookmarkIconSource(workspaceID, normalizedURL, referrer, cacheKey string) (string, error) {
	data, mimeType, err := loadBookmarkIconFromSource(workspaceID, normalizedURL, referrer)
	if err != nil || data == nil {
		return "", err
	}

	contentType := normalizeBookmarkIconMimeType(mimeType)
	if dataURL := bookmarkIconDataURLWithContentType(data, contentType); dataURL != "" {
		storeBookmarkIconMemoryCache(cacheKey, dataURL)
		if cp := bookmarkIconCachePath(workspaceID, normalizedURL); cp != "" {
			_ = writeBookmarkIconCache(cp, referrer, data, contentType)
		}
		return dataURL, nil
	}
	return "", fmt.Errorf("bookmarkicon: invalid icon content")
}

// ---- 内存缓存 ----

// cacheBookmarkIconResult 缓存书签图标结果到内存。
// [S 汇编 0x14075c6a0]：参数 key/val → HashTrieMap 存储。
func cacheBookmarkIconResult(key, dataURL string) {
	bookmarkIconMemCache.Store(key, dataURL)
}

// loadBookmarkIconMemoryCache 从内存缓存加载。
// [S 汇编 0x14075c7a0]：HashTrieMap.Load(key) → (val, ok)。
func loadBookmarkIconMemoryCache(key string) (string, bool) {
	v, ok := bookmarkIconMemCache.Load(key)
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// storeBookmarkIconMemoryCache 存储到内存缓存。
// [S 汇编 0x14075cc60]：TrimSpace → 非空且 < 32MB → lock → HashTrieMap.Store → unlock。
func storeBookmarkIconMemoryCache(key, dataURL string) {
	if strings.TrimSpace(key) == "" || dataURL == "" {
		return
	}
	if len(dataURL) > 32<<20 {
		return
	}
	bookmarkIconMemCache.Store(key, dataURL)
}

// ---- 磁盘缓存 ----

// writeBookmarkIconCache 写书签图标到磁盘缓存。
// [S 汇编 0x14075d240]：classifyAutomaticWindowsPath("local") → 写 JSON 记录到缓存目录。
func writeBookmarkIconCache(cachePath, referrer string, data []byte, contentType string) error {
	p := strings.TrimSpace(cachePath)
	if p == "" {
		return fmt.Errorf("bookmarkicon: empty cache path")
	}
	dir := filepath.Dir(p)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(p, data, 0644)
}

// pruneBookmarkIconDiskCache 修剪磁盘缓存（保留最新的 N 个条目）。
// [S 汇编 0x14075d860]：os.ReadDir → 按 mtime 排序 → 删除超量条目。
func pruneBookmarkIconDiskCache(cacheDir string, maxEntries int) error {
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return err
	}
	if len(entries) <= maxEntries {
		return nil
	}
	sort.Slice(entries, func(i, j int) bool {
		fi, _ := entries[i].Info()
		fj, _ := entries[j].Info()
		if fi == nil || fj == nil {
			return false
		}
		return fi.ModTime().After(fj.ModTime())
	})
	for _, e := range entries[maxEntries:] {
		os.Remove(filepath.Join(cacheDir, e.Name()))
	}
	return nil
}

// ---- dataURL 生成 ----

// bookmarkIconDataURLWithContentType 生成 base64 data URL。
// [S 汇编 0x14075df60]：validateBookmarkIconContent → base64.EncodeToString。
func bookmarkIconDataURLWithContentType(data []byte, contentType string) string {
	mime, ok := validateBookmarkIconContent(data)
	if !ok {
		return ""
	}
	if contentType != "" {
		mime = contentType
	}
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data)
}

// ---- 内容校验 ----

// validateBookmarkIconContent 校验书签图标内容。
// [S 汇编 0x14075e040]：nil/空/>0x80000 返回空；SVG→"image/svg+xml"；ICO→"image/x-icon"；
// PNG→"image/png"；否则 "image/webp"。
func validateBookmarkIconContent(data []byte) (string, bool) {
	if len(data) == 0 || len(data) > 0x80000 {
		return "", false
	}

	if validateBookmarkIconSVG(data) {
		return "image/svg+xml", true
	}
	if validateBookmarkIconICO(data) {
		return "image/x-icon", true
	}
	if len(data) >= 8 && bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47}) {
		return "image/png", true
	}
	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xD8 {
		return "image/jpeg", true
	}
	if len(data) >= 12 && bytes.Equal(data[0:4], []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WEBP")) {
		return "image/webp", true
	}
	return "", false
}

// validateBookmarkIconSVG 检查是否为 SVG 内容。
// [S 汇编 0x14075e2e0]：bytes.TrimSpace → 首字节 '<' → memequal <svg 或 <SVG。
func validateBookmarkIconSVG(data []byte) bool {
	b := bytes.TrimSpace(data)
	if len(b) == 0 {
		return false
	}
	if b[0] != '<' {
		return false
	}
	if len(b) >= 4 {
		lower := bytes.ToLower(b[:4])
		if string(lower) == "<svg" || string(lower) == "<?xm" {
			return true
		}
	}
	return false
}

// validateBookmarkIconICO 检查是否为 ICO 内容。
// [S 汇编 0x14075e820]：w[0]==0 && w[2]==1 → count=w[4] → total=6+count*16。
func validateBookmarkIconICO(data []byte) bool {
	if len(data) < 6 {
		return false
	}
	if data[0] != 0 || data[1] != 0 {
		return false
	}
	if data[2] != 1 || data[3] != 0 {
		return false
	}
	count := int(data[4])
	if count < 1 || count > 63 {
		return false
	}
	total := 6 + count*16
	return len(data) >= total
}

// ---- MIME 类型规范化 ----

// normalizeBookmarkIconMimeType 规范化书签图标 MIME 类型。
// [S 汇编 0x14075ebc0]：TrimSpace → ToLower → Index(';') → TrimSpace → switch 长度/内容。
func normalizeBookmarkIconMimeType(mime string) string {
	m := strings.TrimSpace(strings.ToLower(mime))
	if idx := strings.IndexByte(m, ';'); idx >= 0 {
		m = strings.TrimSpace(m[:idx])
	}
	switch {
	case len(m) < 5:
		return m
	case m == "image/svg+xml" || m == "image/svg":
		return "image/svg+xml"
	case m == "image/x-icon" || m == "image/vnd.microsoft.icon":
		return "image/x-icon"
	case m == "image/png" || m == "image/jpeg" || m == "image/webp":
		return m
	case strings.Contains(m, "svg"):
		return "image/svg+xml"
	case strings.Contains(m, "icon"):
		return "image/x-icon"
	case strings.Contains(m, "png"):
		return "image/png"
	default:
		return m
	}
}

// ---- 磁盘缓存读取 ----

// bookmarkIconContentByteLimit 单条图标内容的字节上限。
// [S 实证 0x1411cd538]：三个查询函数共用同一 int64 常量，实测值 524288（0x80000 = 512 KiB），
// 作为 SQL `length(...) <= ?` 的绑定参数下发到 SQLite；
// 与 validateBookmarkIconContent 的 `len(data) > 0x80000` 上界判定互为交叉验证。
const bookmarkIconContentByteLimit = 0x80000

// queryBookmarkIconBitmap 从 Chromium 系 favicon 库查询页面图标位图。
// [S 汇编 0x140760560, 512B] 实证流程：
//
//	db==nil → nil,false；strings.TrimSpace(pageURL) 空 → nil,false →
//	isFirefox==true → SQL(336B @0x140c96c82，instr 前缀匹配 + icon_mapping/favicon_bitmaps 连接)；
//	否则 → SQL(300B @0x140c96a0a，page_url 精确匹配) →
//	绑定参数 [int64(524288), pageURL] → QueryRow →
//	Scan(&[]byte) → Scan 出错或 len==0 → nil,false →
//	`_, ok := validateBookmarkIconContent(data)` → 返回 (data, ok)。
func queryBookmarkIconBitmap(db *sql.DB, pageURL string, isFirefox bool) ([]byte, bool) {
	if db == nil {
		return nil, false
	}
	trimmed := strings.TrimSpace(pageURL)
	if trimmed == "" {
		return nil, false
	}

	var query string
	if isFirefox {
		query = `
			SELECT CASE WHEN length(ib.image_data) <= ? THEN ib.image_data ELSE NULL END
			FROM icon_mapping im
			JOIN favicon_bitmaps ib ON ib.icon_id = im.icon_id
			WHERE instr(im.page_url, ?) = 1
			ORDER BY LENGTH(im.page_url) DESC, COALESCE(ib.width, 0) DESC, COALESCE(ib.height, 0) DESC, COALESCE(ib.last_updated, 0) DESC
			LIMIT 1
		`
	} else {
		query = `
			SELECT CASE WHEN length(ib.image_data) <= ? THEN ib.image_data ELSE NULL END
			FROM icon_mapping im
			JOIN favicon_bitmaps ib ON ib.icon_id = im.icon_id
			WHERE im.page_url = ?
			ORDER BY COALESCE(ib.width, 0) DESC, COALESCE(ib.height, 0) DESC, COALESCE(ib.last_updated, 0) DESC
			LIMIT 1
		`
	}

	var data []byte
	if err := db.QueryRow(query, int64(bookmarkIconContentByteLimit), trimmed).Scan(&data); err != nil {
		return nil, false
	}
	if len(data) == 0 {
		return nil, false
	}
	_, ok := validateBookmarkIconContent(data)
	return data, ok
}

// queryFirefoxBookmarkIconBitmap 从 Firefox moz_pages_w_icons 查询页面图标位图。
// [S 汇编 0x1407608c0, 512B] 实证流程：
//
//	db==nil 或 strings.TrimSpace(imagePath) 空 → nil,false →
//	exact==true → SQL(332B @0x140c96b36，instr 前缀匹配)；
//	否则 → SQL(296B @0x140c968e2，page_url 精确匹配) →
//	绑定 [int64(524288), imagePath] → QueryRow → Scan(&[]byte) →
//	Scan 出错或 len==0 → nil,false → validateBookmarkIconContent → (data, ok)。
func queryFirefoxBookmarkIconBitmap(db *sql.DB, imagePath string, exact bool) ([]byte, bool) {
	if db == nil {
		return nil, false
	}
	trimmed := strings.TrimSpace(imagePath)
	if trimmed == "" {
		return nil, false
	}

	var query string
	if exact {
		query = `
			SELECT CASE WHEN length(i.data) <= ? THEN i.data ELSE NULL END
			FROM moz_pages_w_icons pw
			JOIN moz_icons_to_pages ip ON ip.page_id = pw.id
			JOIN moz_icons i ON i.id = ip.icon_id
			WHERE instr(pw.page_url, ?) = 1
			ORDER BY LENGTH(pw.page_url) DESC, COALESCE(i.width, 0) DESC, COALESCE(i.expire_ms, 0) DESC
			LIMIT 1
		`
	} else {
		query = `
			SELECT CASE WHEN length(i.data) <= ? THEN i.data ELSE NULL END
			FROM moz_pages_w_icons pw
			JOIN moz_icons_to_pages ip ON ip.page_id = pw.id
			JOIN moz_icons i ON i.id = ip.icon_id
			WHERE pw.page_url = ?
			ORDER BY COALESCE(i.width, 0) DESC, COALESCE(i.expire_ms, 0) DESC
			LIMIT 1
		`
	}

	var data []byte
	if err := db.QueryRow(query, int64(bookmarkIconContentByteLimit), trimmed).Scan(&data); err != nil {
		return nil, false
	}
	if len(data) == 0 {
		return nil, false
	}
	_, ok := validateBookmarkIconContent(data)
	return data, ok
}

// queryFirefoxBookmarkIconRoot 从 Firefox moz_icons 按 root 查询图标位图。
// [S 汇编 0x140760ac0, 416B] 实证流程：
//
//	db==nil 或 strings.TrimSpace(rootURL) 空 → nil,false →
//	strings.TrimSpace(rootURL) 二次规整 → SQL(183B @0x140c962ab，WHERE i.root = ?) →
//	绑定 [int64(524288), rootURL] → QueryRow → Scan(&[]byte) →
//	Scan 出错或 len==0 → nil,false → validateBookmarkIconContent → (data, ok)。
func queryFirefoxBookmarkIconRoot(db *sql.DB, rootURL string) ([]byte, bool) {
	if db == nil {
		return nil, false
	}
	trimmed := strings.TrimSpace(rootURL)
	if trimmed == "" {
		return nil, false
	}
	trimmed = strings.TrimSpace(trimmed)

	const query = `
		SELECT CASE WHEN length(i.data) <= ? THEN i.data ELSE NULL END
		FROM moz_icons i
		WHERE i.root = ?
		ORDER BY COALESCE(i.width, 0) DESC, COALESCE(i.expire_ms, 0) DESC
		LIMIT 1
	`

	var data []byte
	if err := db.QueryRow(query, int64(bookmarkIconContentByteLimit), trimmed).Scan(&data); err != nil {
		return nil, false
	}
	if len(data) == 0 {
		return nil, false
	}
	_, ok := validateBookmarkIconContent(data)
	return data, ok
}

// queryBookmarkIconBitmap 在租约保护下查询 Chromium 系图标。
// [S 汇编 0x14075fd80, 448B] 实证流程：
//
//	lease==nil → nil,false →
//	lease.mu(@+0x10 readerCount) RLock（快路径 lock xadd；慢路径 SemacquireRWMutexR）→
//	defer RUnlock →
//	lease.released(@+0x20) 为真 或 lease.entry(@+0x18)==nil → nil,false →
//	否则 → queryBookmarkIconBitmap(lease.entry.db(@+0x0), pageURL, exact) 原样透传。
func (l *bookmarkFaviconDBLease) queryBookmarkIconBitmap(pageURL string, exact bool) ([]byte, bool) {
	if l == nil {
		return nil, false
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.released || l.entry == nil {
		return nil, false
	}
	return queryBookmarkIconBitmap(l.entry.db, pageURL, exact)
}

// queryFirefoxBookmarkIconData 在租约保护下按候选序列查询 Firefox 图标。
// [S 汇编 0x14075ffa0, 384B] 实证流程：
//
//	lease==nil → nil,false →
//	lease.mu RLock → defer RUnlock →
//	lease.released 或 lease.entry==nil → nil,false →
//	bookmarkIconCandidates(pageURL) 逐个候选：
//	  先 queryFirefoxBookmarkIconBitmap(db, cand, false)（精确）→ 命中即返回；
//	  再 queryFirefoxBookmarkIconBitmap(db, cand, true)（前缀）→ 命中即返回 →
//	全部未命中 → bookmarkIconRootCandidates(pageURL) 逐个候选：
//	  queryFirefoxBookmarkIconRoot(db, cand) → 命中即返回 →
//	仍无 → nil,false。
func (l *bookmarkFaviconDBLease) queryFirefoxBookmarkIconData(pageURL string) ([]byte, bool) {
	if l == nil {
		return nil, false
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.released || l.entry == nil {
		return nil, false
	}

	for _, candidate := range bookmarkIconCandidates(pageURL) {
		if data, ok := queryFirefoxBookmarkIconBitmap(l.entry.db, candidate, false); ok {
			return data, true
		}
		if data, ok := queryFirefoxBookmarkIconBitmap(l.entry.db, candidate, true); ok {
			return data, true
		}
	}
	for _, candidate := range bookmarkIconRootCandidates(pageURL) {
		if data, ok := queryFirefoxBookmarkIconRoot(l.entry.db, candidate); ok {
			return data, true
		}
	}
	return nil, false
}

// loadBookmarkIconFromCache 从磁盘缓存加载书签图标。
// [S-sig 汇编 0x14075ee20]：classifyAutomaticWindowsPath("local") → resolveBookmarkFaviconDBPath →
// openBookmarkFaviconDB → queryBookmarkIconBitmap。
// openBookmarkFaviconDB 的租约单例装配（1760B asm）待专项还原，当前以骨架占位。
func loadBookmarkIconFromCache(workspaceID, normalizedURL string) (string, error) {
	_ = workspaceID
	_ = normalizedURL
	return "", nil
}

// openBookmarkFaviconDB 打开书签图标 SQLite 数据库并返回租约。
// [S 汇编 0x14075f5e0, 1760B] 全量实证：
//
//	strings.TrimSpace(dbPath) 空 → fmt.Errorf(27B) →
//	classifyAutomaticWindowsPath(...) != "local" → fmt.Errorf(45B) →
//	os.Stat(trimmed) → err 非空 → 错误返回 →
//	全局互斥锁 CAS → HashTrieMap.Load(dbPath) 缓存命中：
//	 entry.modifiedAt == fi.ModTime → 递增 users.Add(1) → 复用 entry → 返回 (*bookmarkFaviconDBLease, nil)
//	 entry.modifiedAt != fi.ModTime → LoadAndDelete → 关闭旧 entry → 走新建路径 →
//	缓存未命中：shouldSnapshot → snapshotSQLiteDatabase → sqliteReadonlyURI →
//	sql.Open → SetMaxOpenConns(1) → SetMaxIdleConns(1) →
//	构造 entry {db, cleanup, modifiedAt, size, users, closeOnce} →
//	重新加锁 → 二次检查（双检锁）→ 存在即关闭新 entry 复用旧 / 不存在则 Swap 写入缓存 →
//	返回 (*bookmarkFaviconDBLease, nil)
func openBookmarkFaviconDB(dbPath string) (*bookmarkFaviconDBLease, error) {
	trimmed := strings.TrimSpace(dbPath)
	if trimmed == "" {
		return nil, fmt.Errorf("bookmarkicon: empty favicon db path")
	}
	kind, _ := classifyAutomaticWindowsPath(trimmed, nil)
	if kind != "local" {
		return nil, fmt.Errorf("bookmarkicon: favicon db path is not a local path: %s", trimmed)
	}
	fi, err := os.Stat(trimmed)
	if err != nil {
		return nil, err
	}

	// 全局缓存单例：bookmarkFaviconDBCache keyed by dbPath。
	// 汇编 HashTrieMap @var (0x14fb460 global CAS + HashTrieMap)
	// 此处以 *sync.Map 替代。

	// 尝试获取现有 entry
	if lease := tryGetOrReuseFaviconDBEntry(trimmed, fi); lease != nil {
		return lease, nil
	}

	return openNewFaviconDBEntry(trimmed, fi)
}

// tryGetOrReuseFaviconDBEntry 尝试从全局缓存获取或复用 favicon DB entry。
// 加锁查找：命中且 modifiedAt 一致则递增引用计数后返回新租约；否则返回 nil 由调用者新建。
// [S-inline 内联]：内联于 openBookmarkFaviconDB(0x14075f5e0, 1760B)，无独立符号；
// 对应汇编「互斥锁 → HashTrieMap.Load → modifiedAt 比对 → 复用/删除旧 entry」分支。
func tryGetOrReuseFaviconDBEntry(dbPath string, fi os.FileInfo) *bookmarkFaviconDBLease {
	bookmarkFaviconDBCacheMu.Lock()
	defer bookmarkFaviconDBCacheMu.Unlock()

	v, ok := bookmarkFaviconDBCache.Load(dbPath)
	if !ok {
		return nil
	}
	entry, ok := v.(*bookmarkFaviconDBCacheEntry)
	if !ok || entry == nil {
		return nil
	}
	if !entry.modifiedAt.Equal(fi.ModTime()) {
		// modifiedAt 变化：删除旧 entry，由调用者新建
		bookmarkFaviconDBCache.Delete(dbPath)
		go closeBookmarkFaviconDBCacheEntry(&bookmarkFaviconDBLease{entry: entry})
		return nil
	}
	entry.users.Add(1)
	return &bookmarkFaviconDBLease{entry: entry}
}

// openNewFaviconDBEntry 打开数据库，构造 entry，存入全局缓存。
// [S-inline 内联]：内联于 openBookmarkFaviconDB(0x14075f5e0, 1760B)，无独立符号；
// 对应汇编「快照→sqliteReadonlyURI→sql.Open→SetMaxConns(1)→构造 entry→双检锁→Swap 写入」分支。
func openNewFaviconDBEntry(dbPath string, fi os.FileInfo) (*bookmarkFaviconDBLease, error) {
	// 快照检查
	if shouldSnapshotBookmarkFaviconDB(nil) {
		if err := snapshotSQLiteDatabase(dbPath); err != nil {
			// 快照失败不阻断主流程（非致命）
			_ = err
		}
	}

	uri := sqliteReadonlyURI(dbPath)
	db, err := sql.Open("sqlite3", uri)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	entry := &bookmarkFaviconDBCacheEntry{
		db:         db,
		cleanup:    nil,
		modifiedAt: fi.ModTime(),
		size:       fi.Size(),
	}
	entry.users.Add(1)

	// 重新加锁做双检
	bookmarkFaviconDBCacheMu.Lock()
	defer bookmarkFaviconDBCacheMu.Unlock()

	if v, ok := bookmarkFaviconDBCache.Load(dbPath); ok {
		if existing, ok := v.(*bookmarkFaviconDBCacheEntry); ok && existing != nil {
			// 并发冲突：其他 goroutine 已插入 entry，关闭新的用旧的
			_ = entry.db.Close()
			entry.users.Done()
			existing.users.Add(1)
			return &bookmarkFaviconDBLease{entry: existing}, nil
		}
	}

	bookmarkFaviconDBCache.Store(dbPath, entry)
	return &bookmarkFaviconDBLease{entry: entry}, nil
}

// bookmarkFaviconDBCache 全局 favicon 数据库条目缓存。
var bookmarkFaviconDBCache sync.Map

// bookmarkFaviconDBCacheMu 保护全局缓存操作。
var bookmarkFaviconDBCacheMu sync.Mutex

// bookmarkIconCandidates 生成页面对应的 Firefox 图标查询候选序列。
// [S-sig 汇编 0x140760c60, 2784B]：入口 normalizeBookmarkIconURL → 逐级剥路径/查询生成候选
// （scheme://host、scheme://host/path 前缀等），返回 []string。
// 候选生成的逐级剥段规则（含闭包 func1 0x1409f7xxx）待专项逐条还原。
func bookmarkIconCandidates(pageURL string) []string {
	normalized := normalizeBookmarkIconURL(pageURL)
	if normalized == "" {
		return nil
	}
	candidates := []string{normalized}
	if u, err := url.Parse(normalized); err == nil && u.Host != "" {
		root := u.Scheme + "://" + u.Host
		if root != normalized {
			candidates = append(candidates, root)
		}
	}
	return candidates
}

// bookmarkIconRootCandidates 生成 Firefox moz_icons.root 查询候选序列。
// [S-sig 汇编 0x140761740, 1248B] 实证入口链：
//
//	normalizeBookmarkIconURL → net/url.Parse → 取 Host(+0x28/+0x30) → net/url.splitHostPort →
//	strings.TrimSpace；再取 Scheme(+0x0/+0x8) → strings.TrimSpace → strings.ToLower（空则用
//	5B 常量 @0x140c35bde）→ runtime.makeslice(4) 预分配 4 元素容量 →
//	填充候选（host 变体 / origin 变体）→ 返回 []string。
//
// 4 元素预分配的逐项装配与 host 变体规则待专项逐条还原。
func bookmarkIconRootCandidates(rootURL string) []string {
	normalized := normalizeBookmarkIconURL(rootURL)
	if normalized == "" {
		return nil
	}
	u, err := url.Parse(normalized)
	if err != nil {
		return nil
	}
	host := strings.TrimSpace(u.Host)
	if host == "" {
		return nil
	}
	scheme := strings.ToLower(strings.TrimSpace(u.Scheme))
	if scheme == "" {
		scheme = "https"
	}
	candidates := make([]string, 0, 4)
	candidates = append(candidates, scheme+"://"+host)
	candidates = append(candidates, host)
	return candidates
}

// ---- 源加载 ----

// loadBookmarkIconFromSource 从源加载书签图标（本地文件 → 缓存 → 下载）。
// [S 汇编 0x14075f140]：classifyAutomaticWindowsPath("local") →
// resolveBookmarkFaviconDBPath → openBookmarkFaviconDB → queryBookmarkIconBitmap。
func loadBookmarkIconFromSource(workspaceID, normalizedURL, referrer string) ([]byte, string, error) {
	p := strings.TrimSpace(normalizedURL)
	if p == "" {
		return nil, "", fmt.Errorf("bookmarkicon: empty URL")
	}
	// 下载
	return downloadBookmarkIcon(normalizedURL)
}

// closeBookmarkFaviconDBCacheEntry 关闭 favicon DB 缓存条目。
// [S 汇编 0x140760180]：通过 lease 释放底层 entry（closeOnce 幂等）。
func closeBookmarkFaviconDBCacheEntry(lease *bookmarkFaviconDBLease) {
	if lease == nil || lease.entry == nil {
		return
	}
	lease.entry.closeOnce.Do(func() {
		if lease.entry.cleanup != nil {
			lease.entry.cleanup()
		}
		if lease.entry.db != nil {
			_ = lease.entry.db.Close()
		}
	})
	lease.mu.Lock()
	lease.released = true
	lease.mu.Unlock()
}

// shouldSnapshotBookmarkFaviconDB 判断是否需要快照 favicon DB。
// [S 汇编 0x140760260]。
func shouldSnapshotBookmarkFaviconDB(lease *bookmarkFaviconDBLease) bool {
	return lease != nil
}

// resolveBookmarkFaviconDBPath 解析书签 favicon 数据库路径。
// [S 汇编 0x1407602c0]：TrimSpace → resolveBookmarkSourceKind → "firefox" → Dir + "favicons.sqlite"。
func resolveBookmarkFaviconDBPath(workspaceID string) string {
	return filepath.Join(workspaceID, "favicons.sqlite")
}

// ---- URL 规范化 ----

// normalizeBookmarkIconURL 规范化书签图标 URL。
// [S 汇编 0x140761e40]：TrimSpace → net/url.Parse → 无 Scheme 补 "https://" → 再 Parse。
func normalizeBookmarkIconURL(rawURL string) string {
	p := strings.TrimSpace(rawURL)
	if p == "" {
		return ""
	}
	u, err := url.Parse(p)
	if err != nil || u.Scheme == "" {
		u, err = url.Parse("https://" + p)
		if err != nil {
			return ""
		}
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return ""
	}
	return u.String()
}

// ---- 下载 ----

// downloadBookmarkIcon 下载书签图标。
// [S 汇编 0x140761fc0]：bookmarkIconDownloadTargets → HTTP GET → validate → base64。
func downloadBookmarkIcon(pageURL string) ([]byte, string, error) {
	targets := bookmarkIconDownloadTargets(pageURL)
	for _, t := range targets {
		resp, err := http.Get(t)
		if err != nil {
			continue
		}
		defer resp.Body.Close()
		data, err := io.ReadAll(io.LimitReader(resp.Body, 0x80000))
		if err != nil {
			continue
		}
		if mime, ok := validateBookmarkIconContent(data); ok {
			return data, mime, nil
		}
	}
	return nil, "", fmt.Errorf("bookmarkicon: no valid icon found")
}

// fetchBookmarkIconTarget 通过 channel 返回书签图标下载目标。
// [S 汇编 0x140762380]：chan send → 超时 4s → 下载 → chan recv。
func fetchBookmarkIconTarget(workspaceID, pageURL string) (string, error) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		normalizedURL := normalizeBookmarkIconURL(pageURL)
		if normalizedURL == "" {
			return
		}
		cacheKey := bookmarkIconCacheKey(normalizedURL, "")
		if _, ok := loadBookmarkIconMemoryCache(cacheKey); ok {
			return
		}
		data, mime, err := downloadBookmarkIcon(normalizedURL)
		if err != nil || data == nil {
			return
		}
		contentType := normalizeBookmarkIconMimeType(mime)
		if dataURL := bookmarkIconDataURLWithContentType(data, contentType); dataURL != "" {
			storeBookmarkIconMemoryCache(cacheKey, dataURL)
		}
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
	}
	return "", nil
}

// ---- 下载失败缓存 ----

// bookmarkIconDownloadFailureKey 生成下载失败缓存键。
// [S 汇编 0x140762b20]：sha256(URL) → hex[:N]。
func bookmarkIconDownloadFailureKey(pageURL, referrer string) string {
	h := sha256.Sum256([]byte(pageURL + "|" + referrer))
	return hex.EncodeToString(h[:8])
}

// isBookmarkIconDownloadFailureCached 检查下载失败是否已缓存。
// [S 汇编 0x140762c40]。
func isBookmarkIconDownloadFailureCached(pageURL, referrer string) bool {
	key := bookmarkIconDownloadFailureKey(pageURL, referrer)
	_, ok := loadBookmarkIconMemoryCache("fail:" + key)
	return ok
}

// cacheBookmarkIconDownloadFailure 缓存下载失败记录。
// [S 汇编 0x140762de0]。
func cacheBookmarkIconDownloadFailure(pageURL, referrer string) {
	key := bookmarkIconDownloadFailureKey(pageURL, referrer)
	bookmarkIconMemCache.Store("fail:"+key, "1")
}

// deleteBookmarkIconDownloadFailure 删除下载失败记录。
// [S 汇编 0x140763160]。
func deleteBookmarkIconDownloadFailure(pageURL, referrer string) {
	key := bookmarkIconDownloadFailureKey(pageURL, referrer)
	bookmarkIconMemCache.Delete("fail:" + key)
}

// ---- 下载目标生成 ----

// bookmarkIconDownloadTargets 生成书签图标下载目标列表。
// [S 汇编 0x140763380]：normalizeBookmarkIconURL → net/url.Parse → 按 scheme/host 生成候选。
func bookmarkIconDownloadTargets(pageURL string) []string {
	u, err := url.Parse(strings.TrimSpace(pageURL))
	if err != nil {
		return nil
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil
	}
	targets := make([]string, 0, 4)
	scheme := u.Scheme
	host := u.Host
	targets = append(targets, fmt.Sprintf("%s://%s/favicon.ico", scheme, host))
	targets = append(targets, fmt.Sprintf("%s://%s/apple-touch-icon.png", scheme, host))
	targets = append(targets, fmt.Sprintf("https://www.google.com/s2/favicons?domain=%s", host))
	targets = append(targets, fmt.Sprintf("https://icons.duckduckgo.com/ip3/%s.ico", host))
	return targets
}

// bookmarkIconSupportsWWWVariant 检查是否支持 WWW 变体。
// [S 汇编 0x140764180]：检查 host 是否以 "www." 开头。
func bookmarkIconSupportsWWWVariant(pageURL string) bool {
	u, err := url.Parse(strings.TrimSpace(pageURL))
	if err != nil {
		return false
	}
	return strings.HasPrefix(u.Host, "www.")
}

// ---- 缓存路径/键生成 ----

// bookmarkIconCachePath 生成书签图标磁盘缓存路径。
// [S 汇编 0x140764260]：bookmarkIconCacheDomain → TrimSpace → filepath.Join(workspace, domain)。
func bookmarkIconCachePath(workspaceID, normalizedURL string) string {
	domain := bookmarkIconCacheDomain(normalizedURL)
	if domain == "" {
		return ""
	}
	baseDir := strings.TrimSpace(workspaceID)
	if baseDir == "" {
		return ""
	}
	return filepath.Join(baseDir, "bookmarkicons", domain)
}

// bookmarkIconCacheDomain 从 URL 提取可注册域名（eTLD+1）。
// [S 汇编 0x140764480, 352B(0x160)]：url.Parse→err 返 ""；TrimSpace(u.Hostname()) 空返 ""；
// publicsuffix.EffectiveTLDPlusOne→err 或 Trim 空返 host；否则 Trim(domain)。
// 证伪纠正：旧体手动 "www." 剥离 + 末两段拼接，实为 publicsuffix 库语义。
func bookmarkIconCacheDomain(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	host := strings.TrimSpace(u.Hostname())
	if host == "" {
		return ""
	}
	domain, err := publicsuffix.EffectiveTLDPlusOne(host)
	if err != nil {
		return host
	}
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return host
	}
	return domain
}

// bookmarkIconCacheKey 生成书签图标缓存键。
// [S 汇编 0x1407643c0]：TrimSpace(URL) → ToLower + TrimSpace(referrer) → ToLower → concat。
func bookmarkIconCacheKey(normalizedURL, referrer string) string {
	u := strings.ToLower(strings.TrimSpace(normalizedURL))
	r := strings.ToLower(strings.TrimSpace(referrer))
	if r == "" {
		return u
	}
	return u + "|" + r
}

// ---- 类型定义 ----
// bookmarkFaviconDB / bookmarkFaviconDBLease 定义在 types_bookmark.go 中。

// snapshotSQLiteDatabase 对 SQLite 数据库执行快照。
// [S 汇编 0x14076fb40, 832B]：
//
//	TrimSpace → classifyAutomaticWindowsPath("local") → os.Stat →
//	非目录检查 → Size ≤ 1GB → os.MkdirTemp("") → filepath.Base →
//	filepath.Join → backupSQLiteDatabase(trimmed, backupPath) →
//	失败则 os.RemoveAll(tmpDir) → 返回 error。
func snapshotSQLiteDatabase(dbPath string) error {
	trimmed := strings.TrimSpace(dbPath)
	if trimmed == "" {
		return fmt.Errorf("bookmarkicon: empty favicon db path")
	}
	kind, _ := classifyAutomaticWindowsPath(trimmed, nil)
	if kind != "local" {
		return fmt.Errorf("bookmarkicon: favicon db path is not a local path: %s", trimmed)
	}
	fi, err := os.Stat(trimmed)
	if err != nil {
		return err
	}
	if fi.Size() == 0 {
		return fmt.Errorf("bookmarkicon: snapshot path is a directory")
	}
	if fi.Size() > 0x40000000 {
		return fmt.Errorf("bookmarkicon: snapshot path too large")
	}

	tmpDir, err := os.MkdirTemp("", "bookmarkicon-snapshot-favicon-*")
	if err != nil {
		return err
	}

	base := filepath.Base(trimmed)
	backupPath := filepath.Join(tmpDir, base)
	if err := backupSQLiteDatabase(trimmed, backupPath); err != nil {
		os.RemoveAll(tmpDir)
		return err
	}
	return nil
}

// sqliteFileURI 构造文件风格 SQLite URI。
// [S 汇编 0x1407707e0, 512B]：
//
//	TrimSpace → filepathlite.Clean → replaceStringByte('\\','/') →
//	url.URL{Scheme:"file"} + path → 可选的 params.Encode() → URL.String()
func sqliteFileURI(dbPath string, params url.Values) string {
	trimmed := strings.TrimSpace(dbPath)
	if trimmed == "" {
		return ""
	}
	cleaned := filepath.ToSlash(filepath.Clean(trimmed))

	u := &url.URL{Scheme: "file"}
	if strings.HasPrefix(cleaned, "//") {
		cleaned = strings.TrimPrefix(cleaned, "//")
		if before, after, found := strings.Cut(cleaned, "://"); found {
			u.Host = before
			u.Path = after
		} else {
			u.Host = cleaned
		}
	} else if len(cleaned) >= 3 && isAlpha(cleaned[0]) && cleaned[1] == ':' && cleaned[2] == '/' {
		u.Path = "///" + cleaned
	} else {
		u.Path = cleaned
	}

	if len(params) > 0 {
		u.RawQuery = params.Encode()
	}
	return u.String()
}

// sqliteReadonlyURI 构造只读 SQLite URI（含 mode=ro 参数）。
// [S 汇编 0x1407705e0, 512B]：
//
//	TrimSpace → url.Values{"mode":{"ro"},"_journal_mode":{"WAL"}} →
//	sqliteFileURI(trimmed, params) → 返回只读 URI string。
func sqliteReadonlyURI(dbPath string) string {
	trimmed := strings.TrimSpace(dbPath)
	if trimmed == "" {
		return ""
	}
	params := url.Values{
		"mode":          {"ro"},
		"_journal_mode": {"WAL"},
	}
	return sqliteFileURI(trimmed, params)
}

// backupSQLiteDatabase 执行 SQLite 在线备份。
// [S 汇编 0x14076fec0]：
//
//	sqliteReadonlyURI → sql.Open("sqlite3", uri) → db.Conn → Conn.Raw(func(driverConn any) error {
//	  cast to sqliteBackuper → NewBackup(backupPath) → Step(256) 循环 → 完成
//	}) → 返回 error。
func backupSQLiteDatabase(dbPath, backupPath string) error {
	uri := sqliteReadonlyURI(dbPath)
	db, err := sql.Open("sqlite3", uri)
	if err != nil {
		return fmt.Errorf("bookmarkicon: backup open db: %w", err)
	}
	defer db.Close()

	conn, err := db.Conn(context.Background())
	if err != nil {
		return fmt.Errorf("bookmarkicon: backup get conn: %w", err)
	}
	defer conn.Close()

	err = conn.Raw(func(driverConn any) error {
		backuper, ok := driverConn.(sqliteBackuper)
		if !ok {
			return fmt.Errorf("bookmarkicon: backup: driver does not support sqliteBackuper")
		}
		backup, err := backuper.NewBackup(backupPath)
		if err != nil {
			return fmt.Errorf("bookmarkicon: backup NewBackup: %w", err)
		}
		for {
			done, err := backup.Step(256)
			if err != nil {
				return fmt.Errorf("bookmarkicon: backup Step: %w", err)
			}
			if done {
				break
			}
		}
		return nil
	})
	return err
}

// isAlpha 检查字节是否为英文字母。[S-inline 内联谓词]
func isAlpha(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

// Release 释放书签图标数据库租约（递减引用计数）。
// [S-sig 汇编 0x14075f4a0, 192B]：
// mu.Lock → released=true → mu.Unlock → entry.users.Done()。
func (l *bookmarkFaviconDBLease) Release() {
	if l == nil {
		return
	}
	l.mu.Lock()
	l.released = true
	l.mu.Unlock()
	if l.entry != nil {
		l.entry.users.Done()
	}
}
