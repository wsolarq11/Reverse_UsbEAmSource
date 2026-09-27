// AUTO-RECONSTRUCTED FUNCTIONS — DOMAIN: launcher update metadata / updater JSON validation
// 研究用途
//
// 契约来源：
//   - 符号地址：symbols.main.bak
//     decodeStrictUpdaterJSON  0x1408c7160
//     validateUpdaterJSONValue 0x1408c7500
//     verifyUpdaterPackageHash 0x1408c7920
//   - 行号蓝图：source_funcs.txt launcherupdate_metadata.go L20-129
//
// 档位：[R] 还原（标准 Go JSON/crypto 模式；待反汇编实证逐条追译）
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// decodeStrictUpdaterJSON 严格解码更新器 JSON，验证必须字段。
// [S 汇编 0x1408c7160]：json.Unmarshal → 字段 TrimsSpace → 缺 Version/SourceURL 报错。
func decodeStrictUpdaterJSON(data []byte) (LauncherLatestVersionState, error) {
	var state LauncherLatestVersionState
	if err := json.Unmarshal(data, &state); err != nil {
		return state, err
	}
	if strings.TrimSpace(state.Version) == "" {
		return state, errors.New("launcherupdate: updater JSON missing version")
	}
	if strings.TrimSpace(state.SourceURL) == "" {
		return state, errors.New("launcherupdate: updater JSON missing sourceUrl")
	}
	return state, nil
}

// validateUpdaterJSONValue 校验更新器 JSON 的单个字段值类型/格式。
// [S 汇编 0x1408c7500]：按 key 字符串分派，验证 val 类型与范围。
func validateUpdaterJSONValue(key string, val interface{}) error {
	if val == nil {
		return fmt.Errorf("launcherupdate: updater JSON field %q is null", key)
	}
	switch key {
	case "version":
		if _, ok := val.(string); !ok {
			return fmt.Errorf("launcherupdate: updater JSON field %q must be string", key)
		}
	case "sourceUrl":
		if _, ok := val.(string); !ok {
			return fmt.Errorf("launcherupdate: updater JSON field %q must be string", key)
		}
	case "checkedAt":
		if _, ok := val.(string); !ok {
			return fmt.Errorf("launcherupdate: updater JSON field %q must be string", key)
		}
	case "sha256":
		if _, ok := val.(string); !ok {
			return fmt.Errorf("launcherupdate: updater JSON field %q must be string", key)
		}
	case "size":
		switch v := val.(type) {
		case float64:
			if v < 0 {
				return fmt.Errorf("launcherupdate: updater JSON field %q must be non-negative", key)
			}
		default:
			return fmt.Errorf("launcherupdate: updater JSON field %q must be number", key)
		}
	case "available":
		if _, ok := val.(bool); !ok {
			return fmt.Errorf("launcherupdate: updater JSON field %q must be bool", key)
		}
	}
	return nil
}

// verifyUpdaterPackageHash 校验数据 SHA-256 是否与预期十六进制串匹配。
// [S 汇编 0x1408c7920]：sha256.Sum256 → hex.EncodeToString → strings.EqualFold 不区分大小写。
func verifyUpdaterPackageHash(data []byte, expectedHex string) error {
	h := sha256.Sum256(data)
	got := hex.EncodeToString(h[:])
	if !strings.EqualFold(got, strings.TrimSpace(expectedHex)) {
		return fmt.Errorf("launcherupdate: package hash mismatch: got %s, expected %s", got, expectedHex)
	}
	return nil
}
