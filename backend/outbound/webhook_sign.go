package outbound

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
)

// WebhookSignatureHeader is the HTTP header carrying HMAC proof (t=timestamp,v1=hexdigest).
const WebhookSignatureHeader = "X-Mycelo-Signature"

const (
	webhookHeaderDeliveryID = "X-Mycelo-Delivery-Id"
	webhookHeaderEventID    = "X-Mycelo-Event-Id"
	webhookHeaderAttempt    = "X-Mycelo-Attempt"
	webhookHeaderTimestamp  = "X-Mycelo-Timestamp"
)

// BuildWebhookSignatureValue returns a header value: t=<unix_ms>,v1=<hmac_hex>.
// The signed message is: "<unix_ms>.<deliveryID>.<body>". Empty secret yields empty string (no signing).
func BuildWebhookSignatureValue(secret string, deliveryID string, sentAtUnixMs int64, body []byte) string {
	if secret == "" {
		return ""
	}

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(strconv.FormatInt(sentAtUnixMs, 10)))
	_, _ = mac.Write([]byte{'.'})
	_, _ = mac.Write([]byte(deliveryID))
	_, _ = mac.Write([]byte{'.'})
	_, _ = mac.Write(body)

	return fmt.Sprintf("t=%d,v1=%s", sentAtUnixMs, hex.EncodeToString(mac.Sum(nil)))
}
