// AUTO-RECONSTRUCTED TYPES — DOMAIN: icon
// 研究用途
package main

import (
	"time"
)

type RemoteIconSearchResult struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Collection     string `json:"collection"`
	CollectionName string `json:"collectionName"`
	IconData       string `json:"iconData"`
}

type remoteIconifyCollectionRef struct {
	Name string `json:"name"`
}

type remoteIconifySearchResponse struct {
	Icons       []string                              `json:"icons"`
	Collections map[string]remoteIconifyCollectionRef `json:"collections"`
}

type remoteIconCacheEntry struct {
	path       string
	size       int64
	modifiedAt time.Time
}
