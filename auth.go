package main

import (
	"strings"

	kabusapi "github.com/masayoshi4649/KabuStationWebApp/kabusapi"
)

// apitoken は、KabuStation の API トークンを取得します。
//
// 機能:
//   - KabuStation API へ認証リクエストを送り、トークンを取得します。
//
// 引数およびその型:
//   - なし
//
// 返り値およびその型:
//   - (string): 取得したトークンです（失敗時は空文字です）。
func apitoken() string {
	apiPassword := strings.TrimSpace(cfg.System.Apipw)
	if apiPassword == "" {
		return ""
	}

	kabusapi.SetBaseURL("http://localhost:18080/kabusapi")

	code, tok, err := kabusapi.PostAuthToken(
		kabusapi.ReqPostAuthToken{APIPassword: apiPassword},
	)
	if err != nil || code < 200 || code >= 300 || tok.Token == "" {
		return ""
	}
	return tok.Token
}
