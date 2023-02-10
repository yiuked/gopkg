package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

// ECBEncrypt AES-ECB 加密数据
func ECBEncrypt(originData, key []byte) ([]byte, error) {
	return ecbEncrypt(originData, key)
}

// ECBDecrypt AES-ECB 解密数据
func ECBDecrypt(secretData, key []byte) ([]byte, error) {
	return ecbDecrypt(secretData, key)
}

func ecbEncrypt(originData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	originData = PKCS7Padding(originData, block.BlockSize())
	secretData := make([]byte, len(originData))
	blockMode := newECBEncrypted(block)
	blockMode.CryptBlocks(secretData, originData)
	return secretData, nil
}

func ecbDecrypt(secretData, key []byte) (originByte []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockMode := newECBDecrypted(block)
	originByte = make([]byte, len(secretData))
	blockMode.CryptBlocks(originByte, secretData)
	if len(originByte) == 0 {
		return nil, errors.New("blockMode.CryptBlocks error")
	}
	return PKCS7UnPadding(originByte), nil
}

type ecb struct {
	b         cipher.Block
	blockSize int
}

func newECB(b cipher.Block) *ecb {
	return &ecb{
		b:         b,
		blockSize: b.BlockSize(),
	}
}

type ecbEncrypted ecb

// newECBEncrypted returns a BlockMode which encrypts in electronic code book encrypted
// mode, using the given Block.
func newECBEncrypted(b cipher.Block) cipher.BlockMode {
	return (*ecbEncrypted)(newECB(b))
}

func (x *ecbEncrypted) BlockSize() int { return x.blockSize }

func (x *ecbEncrypted) CryptBlocks(dst, src []byte) {
	if len(src)%x.blockSize != 0 {
		panic("crypto/cipher: input not full blocks")
	}
	if len(dst) < len(src) {
		panic("crypto/cipher: output smaller than input")
	}
	for len(src) > 0 {
		x.b.Encrypt(dst, src[:x.blockSize])
		src = src[x.blockSize:]
		dst = dst[x.blockSize:]
	}
}

type ecbDecrypted ecb

// newECBDecrypted returns a BlockMode which decrypts in electronic code book
// mode, using the given Block.
func newECBDecrypted(b cipher.Block) cipher.BlockMode {
	return (*ecbDecrypted)(newECB(b))
}

func (x *ecbDecrypted) BlockSize() int { return x.blockSize }

func (x *ecbDecrypted) CryptBlocks(dst, src []byte) {
	if len(src)%x.blockSize != 0 {
		panic("crypto/cipher: input not full blocks")
	}
	if len(dst) < len(src) {
		panic("crypto/cipher: output smaller than input")
	}
	for len(src) > 0 {
		x.b.Decrypt(dst, src[:x.blockSize])
		src = src[x.blockSize:]
		dst = dst[x.blockSize:]
	}
}
