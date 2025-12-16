package main

import (
	"bufio"
	"log"
	"os"

	kabusapi "github.com/masayoshi4649/KabuStationWebApp/kabusapi"
)

func auth() string {
	kabusapi.SetBaseURL("http://localhost:18080/kabusapi")

	code, tok, err := kabusapi.PostAuthToken(
		kabusapi.ReqPostAuthToken{APIPassword: cfg.System.Apipw},
	)
	if err != nil {
		log.Printf("token error: %v (http=%d)", err, code)
		_, _ = bufio.NewReader(os.Stdin).ReadString('\n') // 改行が来るまでブロック

		return ""
	}
	return tok.Token
}
