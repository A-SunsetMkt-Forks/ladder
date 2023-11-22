package requestmodifers

import (
	"fmt"
	"ladder/proxychain"
	"math/rand"
	"strings"
	"time"
)

// SpoofReferrerFromBaiduSearch modifies the referrer header
// pretending to be from a BaiduSearch
func SpoofReferrerFromBaiduSearch(url string) proxychain.RequestModification {
	return func(chain *proxychain.ProxyChain) error {
		// https://www.baidu.com/link?url=5biIeDvUIihawf3Zbbysach2Xn4H3w3FzO6LZKgSs-B5Yt4M4RUFikokOk5zetf2&wd=&eqid=9da80d8208009b8480000706655d5ed6
		referrer := fmt.Sprintf("https://baidu.com/link?url=%s", generateRandomBaiduURL())
		chain.AddRequestModifications(
			SpoofReferrer(referrer),
			SetRequestHeader("sec-fetch-site", "cross-site"),
			SetRequestHeader("sec-fetch-dest", "document"),
			SetRequestHeader("sec-fetch-mode", "navigate"),
		)
		return nil
	}
}

// utility functions ==================

func generateRandomString(charset string, length int) string {
	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	var stringBuilder strings.Builder
	for i := 0; i < length; i++ {
		stringBuilder.WriteByte(charset[seededRand.Intn(len(charset))])
	}
	return stringBuilder.String()
}

func generateRandomBaiduURL() string {
	const alphanumericCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const hexCharset = "0123456789abcdef"
	randomAlphanumeric := generateRandomString(alphanumericCharset, 30) // Length before "-"
	randomHex := generateRandomString(hexCharset, 16)                   // Length of eqid
	return randomAlphanumeric + "-" + "&wd=&eqid=" + randomHex
}
