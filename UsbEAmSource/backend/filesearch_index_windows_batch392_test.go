package main

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func Test_loadVolumeIndexBucketContext_empty_path(t *testing.T) {
	_, _, err := loadVolumeIndexBucketContext(context.Background(), "   ", nil, false)
	if err == nil {
		t.Fatal("expected error for empty path")
	}
	if !errors.Is(err, errVolumeIndexBucketPathEmpty) {
		t.Fatalf("expected errVolumeIndexBucketPathEmpty, got %v", err)
	}
}

func Test_loadVolumeIndexBucketContext_missing_file(t *testing.T) {
	_, _, err := loadVolumeIndexBucketContext(context.Background(), filepath.Join(t.TempDir(), "nonexistent.bkt"), nil, false)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func Test_saveVolumeIndexBucketContext_nil_v(t *testing.T) {
	err := saveVolumeIndexBucketContext(context.Background(), nil, "path", namePrefixBucketCounts{})
	if err != nil {
		t.Fatalf("expected nil error for nil v, got %v", err)
	}
}

func Test_saveVolumeIndexBucketContext_empty_path(t *testing.T) {
	v := &VolumeIndex{RootFRN: 1, JournalID: 2, LastUSN: 3, GeneratedAt: time.Now()}
	err := saveVolumeIndexBucketContext(context.Background(), v, "   ", namePrefixBucketCounts{})
	if err != nil {
		t.Fatalf("expected nil error for empty path, got %v", err)
	}
}

func Test_roundTripBucketContext(t *testing.T) {
	dir := t.TempDir()
	bktPath := filepath.Join(dir, "test.bkt")
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	v := &VolumeIndex{
		RootFRN:     0x1234,
		JournalID:   0x5678,
		LastUSN:     100,
		GeneratedAt: now,
	}
	counts := namePrefixBucketCounts{
		twoByte: make([]int32, 0x10000),
	}
	counts.firstByte['a'] = 10
	counts.firstByte['b'] = 20
	counts.twoByte[int('a')<<8|int('b')] = 42

	if err := saveVolumeIndexBucketContext(context.Background(), v, bktPath, counts); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	want := volumeIndexBucketIdentity{
		rootFRN:     0x1234,
		journalID:   0x5678,
		lastUSN:     100,
		generatedAt: timeToUnixNano(now),
	}
	loaded, entryCount, err := loadVolumeIndexBucketContext(context.Background(), bktPath, &want, true)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	_ = entryCount
	if loaded == nil {
		t.Fatal("loaded nil")
	}
	if loaded.firstByte['a'] != 10 || loaded.firstByte['b'] != 20 {
		t.Fatalf("firstByte mismatch: a=%d b=%d", loaded.firstByte['a'], loaded.firstByte['b'])
	}
	if len(loaded.twoByte) != 0x10000 {
		t.Fatalf("twoByte len=%d expected 0x10000", len(loaded.twoByte))
	}
	if loaded.twoByte[int('a')<<8|int('b')] != 42 {
		t.Fatalf("twoByte[ab]=%d expected 42", loaded.twoByte[int('a')<<8|int('b')])
	}
}

func Test_buildVolumeIndexCheckpointSnapshotLocked_nil(t *testing.T) {
	ns, np, si, ac, err := buildVolumeIndexCheckpointSnapshotLocked(nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ns != nil || np != nil || si != nil || ac != 0 {
		t.Fatal("expected zero returns for nil v")
	}
}

func Test_buildVolumeIndexCheckpointSnapshotLocked_no_delta(t *testing.T) {
	v := &VolumeIndex{
		Nodes:         []IndexNode{{FRN: 1, NameOffset: 0, NameLen: 3}},
		NamePool:      []byte("abc"),
		SortedIndices: []int32{0},
		activeCount:   1,
	}
	ns, np, si, ac, err := buildVolumeIndexCheckpointSnapshotLocked(v)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(ns) != 1 || ns[0].FRN != 1 {
		t.Fatalf("nodes mismatch: %+v", ns)
	}
	if !bytes.Equal(np, []byte("abc")) {
		t.Fatalf("namePool mismatch: %v", np)
	}
	if len(si) != 1 || si[0] != 0 {
		t.Fatalf("sortedIndices mismatch: %v", si)
	}
	if ac != 1 {
		t.Fatalf("activeCount=%d expected 1", ac)
	}
}

func Test_buildVolumeIndexCheckpointSnapshotLocked_with_delta(t *testing.T) {
	v := &VolumeIndex{deltaByFRN: map[uint64]int32{1: 0}}
	_, _, _, _, err := buildVolumeIndexCheckpointSnapshotLocked(v)
	if err == nil {
		t.Fatal("expected error for delta-present path")
	}
}
