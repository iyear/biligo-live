package live

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
)

func TestEncode(t *testing.T) {
	authParams := map[string]interface{}{
		"platform": "web",
		"protover": 1,
		"roomid":   21852,
		"uid":      int(rand.Float64()*200000000000000.0 + 100000000000000.0),
		"type":     2,
		"key":      "",
	}
	body, _ := json.Marshal(authParams)
	fmt.Println(encode(0, 7, body))
}
func TestLog(t *testing.T) {
	f := "%d %d 666"
	t.Logf("[INFO] "+f, 1, 1)
}
func TestBrotliEncode(t *testing.T) {
	t.Log(brotliEn([]byte("aaaaadsadadsa")))
}
func TestBrotliDecode(t *testing.T) {
	b, err := brotliEn([]byte("aaaaadsadadsa"))
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	t.Log(b)
	b, err = brotliDe(b)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Log(b)
	if string(b) != "aaaaadsadadsa" {
		t.FailNow()
	}
}
