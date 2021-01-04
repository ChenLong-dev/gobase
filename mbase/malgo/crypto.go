/*
 * @Description: 
 * @Author: Chen Long
 * @Date: 2020-12-16 14:15:45
 * @LastEditTime: 2020-12-16 14:15:45
 * @LastEditors: Chen Long
 * @Reference: 
 */


 package malgo

 import (
	 "bytes"
	 "crypto/aes"
	 "crypto/cipher"
	 "crypto/des"
	 "crypto/rc4"
	 "errors"
	 "fmt"
	 "time"
 )
 
 func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	 padding := blockSize - len(ciphertext)%blockSize
	 padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	 return append(ciphertext, padtext...)
 }
 func PKCS5UnPadding(origData []byte) []byte {
	 length := len(origData)
 
	 if length == 0 {
		 return origData
	 }
 
	 // 去掉最后一个字节 unpadding 次
	 unpadding := int(origData[length-1])
 
	 return origData[:(length - unpadding)]
 }
 func ZeroPadding(ciphertext []byte, blockSize int) []byte {
	 padding := blockSize - len(ciphertext)%blockSize
	 if padding == blockSize {
		 return ciphertext
	 } else {
		 padtext := bytes.Repeat([]byte{byte(0)}, padding)
		 return append(ciphertext, padtext...)
	 }
 }
 
 func cryptMbase(key, data []byte) ([]byte, error) {
	 klen := len(key)
	 dlen := len(data)
	 if klen == 0 || dlen == 0 {
		 return data, nil
	 }
 
	 rdata := make([]byte, dlen)
 
	 for i := 0; i+klen <= dlen; i += klen {
		 for j, kd := range key {
			 rdata[i+j] = data[i+j] ^ kd
		 }
	 }
 
	 reslen := dlen % klen
	 dpos := dlen - reslen
	 for k := 0; k < reslen; k++ {
		 rdata[dpos+k] = data[dpos+k] ^ key[k]
	 }
 
	 return rdata, nil
 }
 
 func EncryptMoa(key, data []byte) ([]byte, error) {
	 return cryptMbase(key, data)
 }
 func DecryptMoa(key, data []byte) ([]byte, error) {
	 return cryptMbase(key, data)
 }
 
 func EncryptAesCBC(key, data []byte, iv []byte) ([]byte, error) {
	 if len(data) == 0 {
		 return data, nil
	 }
 
	 klen := len(key)
	 if klen >= 32 {
		 key = key[:32]
	 } else if klen > 24 {
		 padtext := bytes.Repeat([]byte{byte(0)}, 32-klen)
		 key = append(key, padtext...)
	 } else if klen > 16 {
		 padtext := bytes.Repeat([]byte{byte(0)}, 24-klen)
		 key = append(key, padtext...)
	 } else {
		 padtext := bytes.Repeat([]byte{byte(0)}, 16-klen)
		 key = append(key, padtext...)
	 }
 
	 block, err := aes.NewCipher(key)
	 if err != nil {
		 return data, err
	 }
 
	 blockSize := block.BlockSize()
	 ivlen := len(iv)
	 if ivlen < blockSize {
		 padtext := bytes.Repeat([]byte{byte(0)}, blockSize-ivlen)
		 iv = append(iv, padtext...)
	 } else {
		 iv = iv[:blockSize]
	 }
 
	 data = PKCS5Padding(data, blockSize)
	 blockMode := cipher.NewCBCEncrypter(block, iv)
 
	 rdata := make([]byte, len(data))
 
	 blockMode.CryptBlocks(rdata, data)
 
	 return rdata, nil
 }
 
 func DecryptAesCBC(key, data []byte, iv []byte) ([]byte, error) {
	 dlen := len(data)
	 if dlen == 0 {
		 return data, nil
	 }
	 if dlen%16 != 0 {
		 return data, errors.New("data len not 16 align")
	 }
 
	 klen := len(key)
	 if klen >= 32 {
		 key = key[:32]
	 } else if klen > 24 {
		 padtext := bytes.Repeat([]byte{byte(0)}, 32-klen)
		 key = append(key, padtext...)
	 } else if klen > 16 {
		 padtext := bytes.Repeat([]byte{byte(0)}, 24-klen)
		 key = append(key, padtext...)
	 } else {
		 padtext := bytes.Repeat([]byte{byte(0)}, 16-klen)
		 key = append(key, padtext...)
	 }
 
	 block, err := aes.NewCipher(key)
	 if err != nil {
		 return data, err
	 }
 
	 blockSize := block.BlockSize()
	 ivlen := len(iv)
	 if ivlen < blockSize {
		 padtext := bytes.Repeat([]byte{byte(0)}, blockSize-ivlen)
		 iv = append(iv, padtext...)
	 } else {
		 iv = iv[:blockSize]
	 }
 
	 blockMode := cipher.NewCBCDecrypter(block, iv)
 
	 rdata := make([]byte, len(data))
	 blockMode.CryptBlocks(rdata, data)
	 rdata = PKCS5UnPadding(rdata)
 
	 return rdata, nil
 }
 
 //	由使用者确保data长度是16的整数倍，若不是，自动追加0到对齐
func EncryptAesECB(key, data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	klen := len(key)
	if klen >= 32 {
		key = key[:32]
	} else if klen > 24 {
		padtext := bytes.Repeat([]byte{byte(0)}, 32-klen)
		key = append(key, padtext...)
	} else if klen > 16 {
		padtext := bytes.Repeat([]byte{byte(0)}, 24-klen)
		key = append(key, padtext...)
	} else {
		padtext := bytes.Repeat([]byte{byte(0)}, 16-klen)
		key = append(key, padtext...)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return data, err
	}

	blockSize := block.BlockSize()

	data = ZeroPadding(data, blockSize)
	blockMode := NewECBEncrypter(block)
	//blockMode := cipher.NewCBCEncrypter(block, iv)

	rdata := make([]byte, len(data))
	blockMode.CryptBlocks(rdata, data)

	return rdata, nil
}

//	注意，如果加密前数据不是16字节长度对齐的，加密函数添加的0填充会返回给用户，自行识别
func DecryptAesECB(key, data []byte) ([]byte, error) {
	dlen := len(data)
	if dlen == 0 {
		return data, nil
	}
	if dlen%16 != 0 {
		return data, errors.New("data len not 16 align")
	}

	klen := len(key)
	if klen >= 32 {
		key = key[:32]
	} else if klen > 24 {
		padtext := bytes.Repeat([]byte{byte(0)}, 32-klen)
		key = append(key, padtext...)
	} else if klen > 16 {
		padtext := bytes.Repeat([]byte{byte(0)}, 24-klen)
		key = append(key, padtext...)
	} else {
		padtext := bytes.Repeat([]byte{byte(0)}, 16-klen)
		key = append(key, padtext...)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return data, err
	}

	blockMode := NewECBDecrypter(block)

	rdata := make([]byte, len(data))
	blockMode.CryptBlocks(rdata, data)

	return rdata, nil
}

//	DES
func EncryptDesCBC(key, data []byte, iv []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	klen := len(key)
	if klen >= 8 {
		key = key[:8]
	} else {
		padtext := bytes.Repeat([]byte{byte(0)}, 8-klen)
		key = append(key, padtext...)
	}

	block, err := des.NewCipher(key)
	if err != nil {
		return data, err
	}

	blockSize := block.BlockSize()
	ivlen := len(iv)
	if ivlen < blockSize {
		padtext := bytes.Repeat([]byte{byte(0)}, blockSize-ivlen)
		iv = append(iv, padtext...)
	} else {
		iv = iv[:blockSize]
	}

	data = PKCS5Padding(data, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, iv)

	rdata := make([]byte, len(data))
	blockMode.CryptBlocks(rdata, data)

	return rdata, nil
}

func DecryptDesCBC(key, data []byte, iv []byte) ([]byte, error) {
	dlen := len(data)
	if dlen == 0 {
		return data, nil
	}
	if dlen%16 != 0 {
		return data, errors.New("data len not 16 align")
	}

	klen := len(key)
	if klen >= 8 {
		key = key[:8]
	} else {
		padtext := bytes.Repeat([]byte{byte(0)}, 8-klen)
		key = append(key, padtext...)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return data, err
	}

	blockSize := block.BlockSize()
	ivlen := len(iv)
	if ivlen < blockSize {
		padtext := bytes.Repeat([]byte{byte(0)}, blockSize-ivlen)
		iv = append(iv, padtext...)
	} else {
		iv = iv[:blockSize]
	}

	blockMode := cipher.NewCBCDecrypter(block, iv)

	rdata := make([]byte, len(data))
	blockMode.CryptBlocks(rdata, data)
	rdata = PKCS5UnPadding(rdata)

	return rdata, nil
}

//	RC4
func EncryptRC4(key, data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	klen := len(key)
	if klen > 256 {
		key = key[:256]
	}

	rc4cipher, err := rc4.NewCipher(key)
	if err != nil {
		return data, err
	}

	rdata := make([]byte, len(data))
	rc4cipher.XORKeyStream(rdata, data)
	return rdata, nil
}

func DecryptRC4(key, data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	klen := len(key)
	if klen > 256 {
		key = key[:256]
	}

	rc4cipher, err := rc4.NewCipher(key)
	if err != nil {
		return data, err
	}

	rdata := make([]byte, len(data))
	rc4cipher.XORKeyStream(rdata, data)
	return rdata, nil
}

func cryptMbase2(key, data []byte) {
	klen := len(key)
	dlen := len(data)
	if klen == 0 || dlen == 0 {
		return
	}

	for i := 0; i+klen <= dlen; i += klen {
		for j, kd := range key {
			data[i+j] = data[i+j] ^ kd
		}
	}

	reslen := dlen % klen
	dpos := dlen - reslen
	for k := 0; k < reslen; k++ {
		data[dpos+k] = data[dpos+k] ^ key[k]
	}
}

func EncryptMoa2(key, data []byte) {
	cryptMbase2(key, data)
}

func DecryptMoa2(key, data []byte) {
	cryptMbase2(key, data)
}

var (
	MbaseDefaultKey = []byte("XVZCAEYXFEDFAYADEFYXSEFAYVY")
)

func MakeServerPublicKey() (uint64, []byte) {
	now := uint64(time.Now().UnixNano() / 1000000)
	str := fmt.Sprintf("%d\000", now)
	return now, []byte(str)
}

func MakeServerCryptKey(local uint64, peer []byte) []byte {
	perrInt := uint64(0)
	fmt.Sscanf(string(peer), "%d", &perrInt)
	key := (perrInt * local) + perrInt + local
	if key == 0 {
		key = 0x12345678
	}
	strKey := fmt.Sprintf("%d", key)
	data, _ := EncryptMoa([]byte(strKey), MbaseDefaultKey)
	return data
}