package template

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/xhofe/wopan-sdk-go"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestInitCredentials(t *testing.T) {
	for _, tc := range []struct {
		name, access, refresh  string
		wantRefresh, wantError bool
	}{{"refresh only", "", "refresh", true, false}, {"short access with refresh", "short", "refresh", true, false}, {"both", "0123456789abcdef", "refresh", false, false}, {"access only", "0123456789abcdef", "", false, false}, {"neither", "", "", false, true}, {"short access only", "short", "", false, true}} {
		t.Run(tc.name, func(t *testing.T) {
			d := &Wopan{Addition: Addition{AccessToken: tc.access, RefreshToken: tc.refresh}}
			var calls []string
			client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				if r.UserAgent() != wopan.DefaultUA {
					t.Fatalf("dispatcher User-Agent = %q", r.UserAgent())
				}
				var req struct {
					Header struct {
						Key string `json:"key"`
					} `json:"header"`
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatal(err)
				}
				key := req.Header.Key
				calls = append(calls, key)
				data := `{}`
				switch key {
				case "AppRefreshToken":
					data = `{"access_token":"fedcba9876543210","refresh_token":"new-refresh"}`
				case "AppQueryUser":
					data = `{"userId":"test-user"}`
				case "GetZoneInfo":
					data = `{"url":"https://example.invalid"}`
				case "FamilyUserCurrentEncode":
					data = `{"defaultHomeId":42}`
				}
				return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"STATUS":"200","RSP":{"RSP_CODE":"0000","DATA":` + data + `}}`))}, nil
			})}
			d.client = wopan.DefaultWithRefreshToken(tc.refresh)
			d.client.SetAccessToken(tc.access)
			d.client.SetHttpClient(client)
			d.client.OnRefreshToken(func(a, r string) { d.AccessToken = a; d.RefreshToken = r })
			err := d.initClient(context.Background())
			if tc.wantError {
				if err == nil || len(calls) != 0 {
					t.Fatalf("err=%v calls=%v", err, calls)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if (calls[0] == "AppRefreshToken") != tc.wantRefresh {
				t.Fatalf("calls %v", calls)
			}
			if d.defaultFamilyID != "42" {
				t.Fatalf("family %q", d.defaultFamilyID)
			}
			if tc.wantRefresh && (d.AccessToken != "fedcba9876543210" || d.RefreshToken != "new-refresh") {
				t.Fatal("refreshed tokens not propagated")
			}
		})
	}
}
