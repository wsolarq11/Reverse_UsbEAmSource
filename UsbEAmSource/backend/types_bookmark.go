// AUTO-RECONSTRUCTED TYPES — DOMAIN: bookmark
// 研究用途
package main

import (
	"database/sql"
	"sync"
	"time"
)

type BookmarkFavoritePath struct {
	SourcePath string `json:"sourcePath"`
	FolderPath string `json:"folderPath"`
}

type BookmarkSection struct {
	Custom  []LinkEntry      `json:"custom"`
	Sources []BookmarkSource `json:"sources"`
}

type BookmarkSource struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Browser string `json:"browser"`
	Path    string `json:"path"`
	Enabled bool   `json:"enabled"`
}

type BookmarkState struct {
	Sources []ResolvedBookmarkSource `json:"sources"`
	Items   []ResolvedBookmarkItem   `json:"items"`
}

type chromiumBookmarkFile struct {
	Roots map[string]chromiumBookmarkNode `json:"roots"`
}

type chromiumBookmarkNode struct {
	Children []chromiumBookmarkNode `json:"children"`
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	URL      string                 `json:"url"`
}

type chromiumBookmarkRoot struct {
	BrowserName string
	Path        string
}

type firefoxBookmarkNode struct {
	ID        int64
	ParentID  int64
	Type      int64
	Title     string
	PageTitle string
	URL       string
	GUID      string
}

type firefoxProfileSource struct {
	BrowserName string
	Profile     string
	Path        string
}

type bookmarkFaviconDBCacheEntry struct {
	db         *sql.DB
	cleanup    func()
	modifiedAt time.Time
	size       int64
	users      sync.WaitGroup
	closeOnce  sync.Once
}

type bookmarkFaviconDBLease struct {
	mu       sync.RWMutex
	entry    *bookmarkFaviconDBCacheEntry
	released bool
}

type bookmarkIconDiskCacheEntry struct {
	path       string
	size       int64
	modifiedAt time.Time
}

type bookmarkIconDownloadFailureEntry struct {
	failedAt time.Time
}

type chromiumBookmarkTraversalItem struct {
	node     chromiumBookmarkNode
	ancestry []string
	depth    int
}

type firefoxBookmarkGraphFrame struct {
	id      int64
	depth   int
	exiting bool
}

type firefoxBookmarkTraversalItem struct {
	node     firefoxBookmarkNode
	ancestry []string
	depth    int
}
