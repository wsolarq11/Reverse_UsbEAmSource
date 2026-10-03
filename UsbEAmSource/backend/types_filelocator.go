// AUTO-RECONSTRUCTED TYPES — DOMAIN: filelocator
// 研究用途
package main

import (
	"context"
	"io"
	"regexp"
	"sync"
)

type FileLocatorConfig struct {
	FileNameQuery         string              `json:"fileNameQuery,omitempty"`
	FileNameMode          string              `json:"fileNameMode,omitempty"`
	FileNameMatchCase     bool                `json:"fileNameMatchCase,omitempty"`
	FileNameHistory       []string            `json:"fileNameHistory,omitempty"`
	ContainsTextQuery     string              `json:"containsTextQuery,omitempty"`
	ContainsTextMode      string              `json:"containsTextMode,omitempty"`
	ContainsTextMatchCase bool                `json:"containsTextMatchCase,omitempty"`
	ContainsTextHistory   []string            `json:"containsTextHistory,omitempty"`
	BooleanScope          string              `json:"booleanScope,omitempty"`
	SearchRoot            string              `json:"searchRoot,omitempty"`
	SearchRootHistory     []string            `json:"searchRootHistory,omitempty"`
	MaxSearchFileSizeMB   int                 `json:"maxSearchFileSizeMB,omitempty"`
	IncludeSubfolders     bool                `json:"includeSubfolders"`
	ActiveFilterID        string              `json:"activeFilterId,omitempty"`
	SavedFilters          []FileLocatorFilter `json:"savedFilters,omitempty"`
}

type FileLocatorFilter struct {
	ID     string `json:"id"`
	Remark string `json:"remark"`
	Value  string `json:"value"`
	Type   string `json:"type"`
}

type FileLocatorLineMatch struct {
	LineNumber int                      `json:"lineNumber"`
	Text       string                   `json:"text"`
	Ranges     []FileLocatorTextRange   `json:"ranges"`
	Before     []FileLocatorPreviewLine `json:"before"`
	After      []FileLocatorPreviewLine `json:"after"`
}

type FileLocatorPreviewLine struct {
	LineNumber int                    `json:"lineNumber"`
	Text       string                 `json:"text"`
	Ranges     []FileLocatorTextRange `json:"ranges,omitempty"`
}

type FileLocatorResultDetail struct {
	Path          string                 `json:"path"`
	Name          string                 `json:"name"`
	Directory     string                 `json:"directory"`
	Size          int64                  `json:"size"`
	ModifiedAt    string                 `json:"modifiedAt"`
	LineCount     int                    `json:"lineCount"`
	MatchCount    int                    `json:"matchCount"`
	StoredMatches int                    `json:"storedMatches"`
	Truncated     bool                   `json:"truncated"`
	TextAvailable bool                   `json:"textAvailable"`
	Reason        string                 `json:"reason"`
	Matches       []FileLocatorLineMatch `json:"matches"`
}

type FileLocatorResultSummary struct {
	Path       string `json:"path"`
	Name       string `json:"name"`
	Directory  string `json:"directory"`
	Size       int64  `json:"size"`
	ModifiedAt string `json:"modifiedAt"`
	MatchCount int    `json:"matchCount"`
	Preview    string `json:"preview"`
}

type FileLocatorState struct {
	Request              FileLocatorConfig          `json:"request"`
	Running              bool                       `json:"running"`
	Paused               bool                       `json:"paused"`
	Completed            bool                       `json:"completed"`
	Cancelled            bool                       `json:"cancelled"`
	StartedAt            string                     `json:"startedAt"`
	FinishedAt           string                     `json:"finishedAt"`
	ElapsedMilliseconds  int64                      `json:"elapsedMilliseconds"`
	CurrentPath          string                     `json:"currentPath"`
	CheckedItemCount     int                        `json:"checkedItemCount"`
	CheckedBytes         int64                      `json:"checkedBytes"`
	SearchedItemCount    int                        `json:"searchedItemCount"`
	SearchedBytes        int64                      `json:"searchedBytes"`
	MatchedFileCount     int                        `json:"matchedFileCount"`
	MatchedBytes         int64                      `json:"matchedBytes"`
	TextMatchCount       int                        `json:"textMatchCount"`
	SkippedBinaryCount   int                        `json:"skippedBinaryCount"`
	SkippedOversizeCount int                        `json:"skippedOversizeCount"`
	ResultLimit          int                        `json:"resultLimit"`
	ResultTruncated      bool                       `json:"resultTruncated"`
	LastError            string                     `json:"lastError"`
	Results              []FileLocatorResultSummary `json:"results"`
}

type FileLocatorTextRange struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

type fileLocatorBinaryNode struct {
	op    int
	left  fileLocatorBooleanNode
	right fileLocatorBooleanNode
}

type fileLocatorBooleanExpression struct {
	root             fileLocatorBooleanNode
	positiveMatchers []fileLocatorTermMatcher
}

type fileLocatorBooleanNode interface {
	Eval(string) bool
}

type fileLocatorContextReader struct {
	ctx    context.Context
	reader io.Reader
}

// Read 带 context 的读取（io.Reader 接口方法）。
// [S-sig 0x140a04f20, 192B]：receiver nil 则 panicwrap；否则解包 reader(+0x00:8)
// 与 ctx(+0x10:8) 两接口，ctx 与读取参数重组后间接调用底层 Read。
// 体待 ctx 注入语义专项还原。
func (r *fileLocatorContextReader) Read(p []byte) (int, error) {
	_, _ = r, p
	return 0, nil
}

type fileLocatorNearNode struct {
	left     fileLocatorBooleanNode
	right    fileLocatorBooleanNode
	distance int
}

type fileLocatorNotNode struct {
	child fileLocatorBooleanNode
}

type fileLocatorPathFilter struct {
	includes []*fileLocatorStringMatcher
	excludes []*fileLocatorStringMatcher
}

type fileLocatorPlainMatcher struct {
	query     string
	queryFold string
	matchCase bool
}

type fileLocatorPreparedSearch struct {
	request         FileLocatorConfig
	roots           []string
	fileNameMatcher *fileLocatorStringMatcher
	contentMatcher  *fileLocatorStringMatcher
	pathFilter      *fileLocatorPathFilter
}

type fileLocatorRegexMatcher struct {
	regex *regexp.Regexp
}

type fileLocatorRuneWindow struct {
	start int
	end   int
}

type fileLocatorService struct {
	lock         sync.RWMutex
	pauseCond    *sync.Cond
	generation   uint64
	state        FileLocatorState
	detailByPath map[string]FileLocatorResultDetail
	cancel       func()
	contentScan  chan struct{}
}

// fileLocatorSearchProgress 搜索进度累加器（runSearch 栈上局部，walkRoot/processFile 共享写入，
// finishSearch 收口时并入 FileLocatorState）。字节布局经 processFile 汇编 0x1407d15a0 实证
// （0x78 字节）：+0x10/+0x18/+0x20/+0x28/+0x30/+0x38/+0x40 计数器槽（自增/累加）、
// +0x58 results 切片、+0x70 resultTruncated。字段级语义 [P]（计数器名按 FileLocatorState 推断，
// +0x00/+0x08/+0x48/+0x50 四槽未实证），本轮仅需类型存在以支撑 walkRoot/processFile [S-sig]。
type fileLocatorSearchProgress struct {
	_                  uint64                    // +0x00 语义待专项
	_                  uint64                    // +0x08 语义待专项
	checkedItemCount   int                       // +0x10
	checkedBytes       int64                     // +0x18
	searchedItemCount  int                       // +0x20
	searchedBytes      int64                     // +0x28
	matchedBytes       int64                     // +0x30
	textMatchCount     int                       // +0x38
	skippedBinaryCount int                       // +0x40
	_                  uint64                    // +0x48 语义待专项
	_                  uint64                    // +0x50 语义待专项
	results            []FileLocatorResultSummary // +0x58
	truncated          bool                      // +0x70
}

type fileLocatorStringMatcher struct {
	query            string
	queryFold        string
	mode             string
	matchCase        bool
	booleanScope     string
	regex            *regexp.Regexp
	plainWildcard    *regexp.Regexp
	booleanExpr      *fileLocatorBooleanExpression
	wholeWordMatcher fileLocatorTermMatcher
}

type fileLocatorTermMatcher interface {
	Exists(string) bool
	Ranges(string) []FileLocatorTextRange
}

type fileLocatorTermNode struct {
	matcher fileLocatorTermMatcher
}

type fileLocatorToken struct {
	kind     int
	value    string
	distance int
}

type fileLocatorWholeWordMatcher struct {
	queryRunes []int32
	queryFold  []int32
	matchCase  bool
}
