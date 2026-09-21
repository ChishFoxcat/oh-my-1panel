package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
)

// ParseRSAPublicKey 解析面板下发的 PEM 公钥（PKIX 或 PKCS1）。
func ParseRSAPublicKey(pemData string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, errors.New("公钥 PEM 解析失败")
	}
	if parsed, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		pub, ok := parsed.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("面板公钥不是 RSA 类型")
		}
		return pub, nil
	}
	pub, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("面板公钥解析失败：%w", err)
	}
	return pub, nil
}

// EncryptPassword 按 1Panel 的约定加密登录密码。
//
// 报文格式 keyCipher:ivBase64:ciphertextBase64：
//   - aesKey 为 32 个十六进制字符（UTF-8 编码后 32 字节，即 AES-256 密钥）
//   - keyCipher = base64(RSA/PKCS1v15(aesKey))
//   - ciphertext = AES-256-CBC/PKCS7(password, aesKey, iv)，iv 随机 16 字节
//
// 对应 1Panel 的 core/utils/encrypt.DecryptPassword。
func EncryptPassword(password string, pub *rsa.PublicKey) (string, error) {
	if password == "" {
		return "", nil
	}
	aesKey, err := randomHex(16)
	if err != nil {
		return "", err
	}
	encryptedKey, err := rsa.EncryptPKCS1v15(rand.Reader, pub, []byte(aesKey))
	if err != nil {
		return "", fmt.Errorf("RSA 加密会话密钥失败：%w", err)
	}
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return "", fmt.Errorf("生成初始向量失败：%w", err)
	}
	block, err := aes.NewCipher([]byte(aesKey))
	if err != nil {
		return "", fmt.Errorf("初始化 AES 失败：%w", err)
	}
	padded := pkcs7Pad([]byte(password), aes.BlockSize)
	ciphertext := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(ciphertext, padded)

	return fmt.Sprintf("%s:%s:%s",
		base64.StdEncoding.EncodeToString(encryptedKey),
		base64.StdEncoding.EncodeToString(iv),
		base64.StdEncoding.EncodeToString(ciphertext),
	), nil
}

func randomHex(byteLen int) (string, error) {
	buf := make([]byte, byteLen)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("生成随机密钥失败：%w", err)
	}
	return hex.EncodeToString(buf), nil
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padded := make([]byte, len(data)+padding)
	copy(padded, data)
	for i := len(data); i < len(padded); i++ {
		padded[i] = byte(padding)
	}
	return padded
}
