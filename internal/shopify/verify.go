package shopify

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"os"
	"sort"
)

func VerifyHMAC(query url.Values) bool {
	givenHmac := query.Get("hmac")
	query.Del("hmac")
	// Build message from sorted key=value
	keys := make([]string, 0, len(query))
	for k := range query {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	msg := ""
	for i, k := range keys {
		if i > 0 {
			msg += "&"
		}
		msg += k + "=" + query.Get(k)
	}
	mac := hmac.New(sha256.New, []byte(os.Getenv("SHOPIFY_API_SECRET")))
	mac.Write([]byte(msg))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(givenHmac))
}
