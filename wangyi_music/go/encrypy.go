package music

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"math/big"
	"math/rand"
	"time"
)

const (
	modulus = "00e0b509f6259df8642dbc35662901477df22677ec152b5ff68ace615bb7b725152b3ab17a876aea8a5aa76d2e417629ec4ee341f56135fccf695280104e0312ecbda92557c93870114af6c9d05c4f7f0c3685b7a46bee255932575cce10b424d813cfe4875d3e82047b97ddef52741d546b8e289dc6935b3ece0462db0a22b8e7"
	nonce   = "0CoJUm6Qyw8W8jud"
	pubKey  = "010001"
	keys    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789+/"
	iv      = "0102030405060708"
)

var userAgentList = [19]string{
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B143 Safari/601.1",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B143 Safari/601.1",
	"Mozilla/5.0 (Linux; Android 5.0; SM-G900P Build/LRX21T) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Mobile Safari/537.36",
	"Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Mobile Safari/537.36",
	"Mozilla/5.0 (Linux; Android 5.1.1; Nexus 6 Build/LYZ28E) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Mobile Safari/537.36",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 10_3_2 like Mac OS X) AppleWebKit/603.2.4 (KHTML, like Gecko) Mobile/14F89;GameHelper",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_5) AppleWebKit/603.2.4 (KHTML, like Gecko) Version/10.1.1 Safari/603.2.4",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 10_0 like Mac OS X) AppleWebKit/602.1.38 (KHTML, like Gecko) Version/10.0 Mobile/14A300 Safari/602.1",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.103 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.12; rv:46.0) Gecko/20100101 Firefox/46.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:46.0) Gecko/20100101 Firefox/46.0",
	"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 6.0)",
	"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.0; Trident/4.0)",
	"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Trident/5.0)",
	"Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.2; Win64; x64; Trident/6.0)",
	"Mozilla/5.0 (Windows NT 6.3; Win64, x64; Trident/7.0; rv:11.0) like Gecko",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/42.0.2311.135 Safari/537.36 Edge/13.10586",
	"Mozilla/5.0 (iPad; CPU OS 10_0 like Mac OS X) AppleWebKit/602.1.38 (KHTML, like Gecko) Version/10.0 Mobile/14A300 Safari/602.1",
}

/*
参数加密过程如下：
1）将请求参数使用一个固定的标识字符串进行 AES加密并进行base64编码
2）从 abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789+/ 中随机取出16次字符组成一个秘钥
3）将1得到的结果使用2得到的秘钥串再进行一次 AES加密并base64编码，至此得到第一个参数 params
4）将2中的秘钥倒序后转ascii码，然后转16进制字符串得到一个中间字符串
5）将4得到的中间字符串转为十进制大整数1，将另两个固定的16进制字符串也都转为十进制大整数2、3
6）5中得到了三个大整数，现在执行 pow(大整数1, 大整数2, 大整数3) ，即1的2次幂取模3，得到新的大整数4
7）将6得到的大整数4转为16进制字符串，并在左边补满0补满256位，最终得到第二个参数 encSecKey
*/

// EncParams 传入参数 得到加密后的参数和一个也被加密的秘钥
func EncParams(param string) (string, string, error) {
	// 创建 key
	secKey := createSecretKey(16)
	aes1, err1 := aesEncrypt(param, nonce)
	// 第一次加密 使用固定的 nonce
	if err1 != nil {
		return "", "", err1
	}
	aes2, err2 := aesEncrypt(aes1, secKey)
	// 第二次加密 使用创建的 key
	if err2 != nil {
		return "", "", err2
	}
	// 得到 加密好的 param 以及 加密好的key
	return aes2, rsaEncrypt(secKey, pubKey, modulus), nil
}

// 创建指定长度的key
func createSecretKey(size int) string {
	// 也就是从 a~9 以及 +/ 中随机拿出指定数量的字符拼成一个 key
	rs := ""
	for i := 0; i < size; i++ {
		pos := rand.Intn(len(keys))
		rs += keys[pos : pos+1]
	}
	return rs
}

// 通过 CBC模式的AES加密 用 sKey 加密 sSrc
func aesEncrypt(sSrc string, sKey string) (string, error) {
	iv := []byte(iv)
	block, err := aes.NewCipher([]byte(sKey))
	if err != nil {
		return "", err
	}
	padding := block.BlockSize() - len([]byte(sSrc))%block.BlockSize()
	src := append([]byte(sSrc), bytes.Repeat([]byte{byte(padding)}, padding)...)
	model := cipher.NewCBCEncrypter(block, iv)
	cipherText := make([]byte, len(src))
	model.CryptBlocks(cipherText, src)
	// 最后使用base64编码输出
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// 将 key 也加密
func rsaEncrypt(key string, pubKey string, modulus string) string {
	// 倒序 key
	rKey := ""
	for i := len(key) - 1; i >= 0; i-- {
		rKey += key[i : i+1]
	}
	// 将 key 转 ascii 编码 然后转成 16 进制字符串
	hexRKey := ""
	for _, char := range []rune(rKey) {
		hexRKey += fmt.Sprintf("%x", int(char))
	}
	// 将 16进制 的 三个参数 转为10进制的 bigint
	bigRKey, _ := big.NewInt(0).SetString(hexRKey, 16)
	bigPubKey, _ := big.NewInt(0).SetString(pubKey, 16)
	bigModulus, _ := big.NewInt(0).SetString(modulus, 16)
	// 执行幂乘取模运算得到最终的bigint结果
	bigRs := bigRKey.Exp(bigRKey, bigPubKey, bigModulus)
	// 将结果转为 16进制字符串
	hexRs := fmt.Sprintf("%x", bigRs)
	// 可能存在不满256位的情况，要在前面补0补满256位
	return addPadding(hexRs, modulus)
}

// 补0步骤
func addPadding(encText string, modulus string) string {
	ml := len(modulus)
	for i := 0; ml > 0 && modulus[i:i+1] == "0"; i++ {
		ml--
	}
	num := ml - len(encText)
	prefix := ""
	for i := 0; i < num; i++ {
		prefix += "0"
	}
	return prefix + encText
}

func fakeAgent() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return userAgentList[r.Intn(19)]
}
