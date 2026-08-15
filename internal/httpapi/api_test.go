package httpapi

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestDecodeRejectsMultipleJSONObjects 是针对以下 Bug 的回归测试：
// 请求体包含多个 JSON 对象时（如 {"expr":"..."}{"expr":"..."}），
// 接口静默忽略第二个及之后的数据并返回 200，而非返回 400。
func TestDecodeRejectsMultipleJSONObjects(t *testing.T) {
	api := New()
	srv := httptest.NewServer(api.Handler())
	defer srv.Close()

	cases := []struct {
		name string
		path string
		body string
	}{
		{"validate 拼接两个对象", "/api/validate", `{"expr":"* * * * *"}{"expr":"0 0 * * *"}`},
		{"next 拼接两个对象", "/api/next", `{"expr":"* * * * *","from":"2026-01-15T10:00:00Z"}{"expr":"0 0 * * *"}`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resp, err := http.Post(srv.URL+c.path, "application/json", strings.NewReader(c.body))
			if err != nil {
				t.Fatalf("请求出错: %v", err)
			}
			defer resp.Body.Close()
			_, _ = io.ReadAll(resp.Body)
			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("status = %d, want %d（多个 JSON 对象应被拒绝）", resp.StatusCode, http.StatusBadRequest)
			}
		})
	}
}
