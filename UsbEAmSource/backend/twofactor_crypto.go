// AUTO-RECONSTRUCTED — DOMAIN: twofactor crypto/password
// 研究用途
package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// randomBytes 生成 n 字节密码学安全随机数。
// [S 汇编实证 0x1409d9760, 0x80]：make([]byte,n)→crypto/rand.Read(b)；
// Read 返回错误时返回 (nil, err)，否则 (b, nil)。
func randomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}

// zeroTwoFactorBytes 将字节切片清零并原样返回。
// [S 汇编实证 0x1409d97e0, 0x80]：len>0 时 memclrNoHeapPointers(ptr,len)（等价 clear），
// 返回原始切片三元组。
func zeroTwoFactorBytes(b []byte) []byte {
	clear(b)
	return b
}

// isTwoFactorPasswordConfigured 判断二因素密码是否已配置。
// [S 汇编实证 0x1409d1da0, 0xc0]：TrimSpace(Salt) 非空 且 TrimSpace(Verifier) 非空。
func isTwoFactorPasswordConfigured(p TwoFactorPasswordConfig) bool {
	return strings.TrimSpace(p.Salt) != "" && strings.TrimSpace(p.Verifier) != ""
}

// deriveTwoFactorKey 由密码与盐派生 AES-256 密钥。
// [S 汇编实证 0x1409d2320, 0xe0]：argon2.IDKey([]byte(TrimSpace(password)), salt,
// time=3, memory=32768(0x8000), threads=4, keyLen=32(0x20))，即 Argon2id。
func deriveTwoFactorKey(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(strings.TrimSpace(password)), salt, 3, 32768, 4, 32)
}

// buildTwoFactorPasswordVerifier 计算密码校验串。
// [S 汇编实证 0x1409d2400, 0x160]：data = "usbeam-two-factor-password:"(27B) + key；
// sha256.Sum256(data) → base64.StdEncoding.EncodeToString(前 32B)。
func buildTwoFactorPasswordVerifier(key []byte) string {
	const salt = "usbeam-two-factor-password:"
	data := make([]byte, 0, len(salt)+len(key))
	data = append(data, salt...)
	data = append(data, key...)
	sum := sha256.Sum256(data)
	return base64.StdEncoding.EncodeToString(sum[:])
}

// verifyTwoFactorPassword 校验密码并返回派生密钥。
// [S 汇编实证 0x1409d1e60, 0x4c0]：normalize→isConfigured→KDF 须 "argon2id-v1"；
// Salt base64 解码须 16B，Verifier base64 解码须 32B；deriveTwoFactorKey 后
// buildTwoFactorPasswordVerifier 与 p.Verifier 常量比较；相等返回 key，否则报错。
func verifyTwoFactorPassword(p TwoFactorPasswordConfig, password string) ([]byte, error) {
	p = normalizeTwoFactorPasswordConfig(p)
	if !isTwoFactorPasswordConfigured(p) {
		return nil, errors.New("请先设置二步验证密码")
	}
	if p.KDF != "argon2id-v1" {
		return nil, fmt.Errorf("%w: 二步验证密码算法不受支持", errTwoFactorDataCorrupted)
	}
	salt, err := base64.StdEncoding.DecodeString(p.Salt)
	if err != nil || len(salt) != 16 {
		return nil, errors.New("二步验证密码配置损坏")
	}
	verifier, err := base64.StdEncoding.DecodeString(p.Verifier)
	if err != nil || len(verifier) != 32 {
		return nil, errors.New("二步验证密码配置损坏")
	}
	key := deriveTwoFactorKey(strings.TrimSpace(password), salt)
	if buildTwoFactorPasswordVerifier(key) != p.Verifier {
		return nil, errors.New("二步验证密码不正确")
	}
	return key, nil
}

// encryptTwoFactorSecret AES-GCM 加密明文，返回 nonce 与密文的 base64 编码。
// [S 汇编实证 0x1409d2560, 0x220]：aes.NewCipher(key)→cipher.NewGCM；
// nonce=randomBytes(NonceSize)；ciphertext=gcm.Seal(nil, nonce, plaintext, nil)；
// 返回 (base64(nonce), base64(ciphertext), nil)，任一步失败返回 ("", "", err)。
func encryptTwoFactorSecret(plaintext, key []byte) (nonceB64, ciphertextB64 string, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", err
	}
	nonce, err := randomBytes(gcm.NonceSize())
	if err != nil {
		return "", "", err
	}
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(nonce),
		base64.StdEncoding.EncodeToString(ciphertext),
		nil
}

// decryptTwoFactorSecret AES-GCM 解密条目密钥并归一化。
// [S 汇编实证 0x1409d2780, 0x500]：validateTwoFactorStoredEntry → aes.NewCipher →
// cipher.NewGCM；nonce base64 解码须 == NonceSize；ciphertext base64 解码须
// Overhead<=len<=65536；gcm.Open(nil, nonce, ciphertext, nil)；plaintext 空报错；
// 返回 normalizeStoredTwoFactorSecretBytes(entry.Kind, plaintext)。
func decryptTwoFactorSecret(entry TwoFactorEntryConfig, key []byte) ([]byte, error) {
	if err := validateTwoFactorStoredEntry(entry); err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce, err := base64.StdEncoding.DecodeString(strings.TrimSpace(entry.SecretNonce))
	if err != nil || len(nonce) != gcm.NonceSize() {
		return nil, fmt.Errorf("%w: 二步验证条目的随机向量损坏", errTwoFactorDataCorrupted)
	}
	ciphertext, err := base64.StdEncoding.DecodeString(strings.TrimSpace(entry.SecretCiphertext))
	if err != nil || gcm.Overhead() > len(ciphertext) || len(ciphertext) > 65536 {
		return nil, fmt.Errorf("%w: 二步验证条目的密文损坏", errTwoFactorDataCorrupted)
	}
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: 二步验证条目认证失败", errTwoFactorDataCorrupted)
	}
	if len(plaintext) == 0 {
		return nil, fmt.Errorf("%w: 二步验证条目密钥为空", errTwoFactorDataCorrupted)
	}
	return normalizeStoredTwoFactorSecretBytes(entry.Kind, plaintext), nil
}
