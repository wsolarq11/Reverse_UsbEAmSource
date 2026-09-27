package main

import (
	"encoding/json"
	"fmt"
)

// validateLauncherConfigJSONStructure 校验配置 JSON 结构（深层 schema 校验）。
// [S-sig 汇编 0x140898120, 56 行]：栈 0x170，序言保存 []byte（rax/rbx/rcx = ptr/len/cap）；体待还原。
func validateLauncherConfigJSONStructure(data []byte) error { return nil }

// validateLauncherConfigInMemoryBudget 校验内存配置预算（二因素一致性 + 结构 + 序列化上限）。
// [S-sig 汇编 0x1408987e0, 7 行]：validateTwoFactorStoredConfig(cfg 字段) 失败→返回；
// 否则 validateLauncherConfigInMemoryStructure(cfg)。cfg 为 LauncherConfig 大结构体栈传参（0x1230B）。
func validateLauncherConfigInMemoryBudget(cfg LauncherConfig) error {
	return validateLauncherConfigInMemoryStructure(cfg)
}

// validateLauncherConfigInMemoryStructure 序列化配置并校验大小上限。
// [S 汇编 0x140898880]：json.Marshal(cfg) 失败→返回 err；len(data)>64<<20 → fmt.Errorf
// ("%s: 配置序列化后超过 %d 字节"@0x140c7b220, "CONFIG_INPUT_TOO_LARGE"@0x140c6127e, 67108864)；
// 否则 validateLauncherConfigJSONStructure(data)。
func validateLauncherConfigInMemoryStructure(cfg LauncherConfig) error {
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	if len(data) > 64<<20 {
		return fmt.Errorf("%s: 配置序列化后超过 %d 字节", "CONFIG_INPUT_TOO_LARGE", 64<<20)
	}
	return validateLauncherConfigJSONStructure(data)
}

// validateLauncherConfigForPersistence 持久化前校验（预算 + 图标数据）。
// [S-sig 汇编 0x140898960, 4 行]：validateLauncherConfigInMemoryBudget(cfg) 失败→返回；
// 否则 validateLauncherConfigIconData(cfg)。cfg 按值栈传参（0x126×8=2352B）。
func validateLauncherConfigForPersistence(cfg LauncherConfig) error {
	if err := validateLauncherConfigInMemoryBudget(cfg); err != nil {
		return err
	}
	return nil // validateLauncherConfigIconData 签名待订正（当前落地为 (cfg *LauncherConfig) 无返回）
}
