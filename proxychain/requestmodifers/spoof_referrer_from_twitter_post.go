package requestmodifers

import (
	"ladder/proxychain"
)

// SpoofReferrerFromTwitterPost modifies the referrer header
// pretending to be from a twitter post
func SpoofReferrerFromTwitterPost(url string) proxychain.RequestModification {
	return func(chain *proxychain.ProxyChain) error {
		chain.AddRequestModifications(
			SpoofReferrer("https://t.co/"),
			SetRequestHeader("sec-fetch-site", "cross-site"),
			SetRequestHeader("sec-fetch-dest", "document"),
			SetRequestHeader("sec-fetch-mode", "navigate"),
		)
		return nil
	}
}
