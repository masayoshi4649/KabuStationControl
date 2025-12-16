package main

import (
	"bufio"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const httpListenAddr = ":3000"

var cfg Config

func main() {
	// -c もしくは --config で指定可能に
	var confPath string
	flag.StringVar(&confPath, "c", "auth.toml", "path to config file")
	flag.StringVar(&confPath, "config", "auth.toml", "path to config file (alias)")
	flag.Parse()

	var err error
	cfg, err = loadConfig(confPath)
	if err != nil {
		log.Printf("failed to load config (%s): %v\n", confPath, err)
		log.Println("Enterキーを押してください...")
		_, _ = bufio.NewReader(os.Stdin).ReadString('\n') // 改行が来るまでブロック
		os.Exit(1)
	}

	rt := gin.Default()

	rt.LoadHTMLGlob("view/*.html")
	rt.Static("/static", "./view")

	rt.GET("/", handleIndexGET)
	rt.GET("/bootauthkabus", handleBootAuthKabusGET)
	rt.GET("/bootapp", handleBootAppGET)

	rt.Run(httpListenAddr)
}

// ----------------------------------------

func handleIndexGET(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{})
}

// ----------------------------------------

// handleBootAuthKabusGET は、KabuStation の起動→待機→認証要求を処理します。
//
// 機能:
//   - KabuStation を起動し、10 秒待機した後にログイン操作スクリプトを実行します。
//
// 引数およびその型:
//   - c (*gin.Context): Gin のコンテキストです。
//
// 返り値およびその型:
//   - なし（HTTP レスポンスとして JSON を返します）
func handleBootAuthKabusGET(c *gin.Context) {
	exePath, err := resolveKabuStationExePath()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": err.Error()})
		return
	}

	wasRunning, runningErr := isKabuStationRunning()
	if runningErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": "KabuStation の起動状態確認に失敗しました", "error": runningErr.Error()})
		return
	}

	started := false
	if !wasRunning {
		if err := exec.Command(exePath).Start(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": "KabuStation の起動に失敗しました", "error": err.Error()})
			return
		}
		started = true
	}

	if started {
		time.Sleep(10 * time.Second)
	}

	out, runErr := runPowerShellFile(
		filepath.Join("cmd", "Click-KabuStationLogin.ps1"),
		"-ExePath", exePath,
		"-TimeoutSeconds", "60",
	)
	if runErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":      false,
			"message": "KabuStation 起動認証（ログイン操作）の実行に失敗しました",
			"error":   runErr.Error(),
			"output":  out,
			"started": started,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"message": "KabuStation 起動認証（ログイン操作）を実行しました",
		"output":  out,
		"started": started,
	})
}

// ----------------------------------------

// handleBootAppGET は、TradeWebApp の起動要求を処理します。
//
// 機能:
//   - 既定の URL もしくは設定済み URL を既定ブラウザで開きます。
//
// 引数およびその型:
//   - c (*gin.Context): Gin のコンテキストです。
//
// 返り値およびその型:
//   - なし（HTTP レスポンスとして JSON を返します）
func handleBootAppGET(c *gin.Context) {
	url := resolveTradeWebAppURL()
	if strings.TrimSpace(url) == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": "TradeWebApp の URL が空です"})
		return
	}

	if err := openURL(url); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": "TradeWebApp の起動（URL を開く）に失敗しました", "error": err.Error(), "url": url})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "message": "TradeWebApp を起動しました（URL を開きました）", "url": url})
}

// ----------------------------------------

// resolveKabuStationExePath は、KabuStation の実行ファイルパスを解決します。
//
// 機能:
//   - 設定ファイル（PATH.KABUSTATION_EXE）→環境変数（KABUSTATION_EXE）→既定パスの順に探索します。
//
// 引数およびその型:
//   - なし
//
// 返り値およびその型:
//   - (string): 実行ファイルパスです。
//   - (error): 見つからない場合や不正な場合のエラーです。
func resolveKabuStationExePath() (string, error) {
	candidates := []string{
		strings.TrimSpace(cfg.Path.KabuStationExe),
		strings.TrimSpace(os.Getenv("KABUSTATION_EXE")),
	}

	localAppData := strings.TrimSpace(os.Getenv("LOCALAPPDATA"))
	if localAppData != "" {
		candidates = append(candidates, filepath.Join(localAppData, "kabuStation", "KabuS.exe"))
	}

	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", errors.New("KabuStation の実行ファイルが見つかりません。設定ファイルの PATH.KABUSTATION_EXE もしくは環境変数 KABUSTATION_EXE を設定してください。")
}

// ----------------------------------------

// isKabuStationRunning は、KabuStation が起動済みかどうかを判定します。
//
// 機能:
//   - Windows の tasklist を用いて KabuS.exe の存在を確認します。
//
// 引数およびその型:
//   - なし
//
// 返り値およびその型:
//   - (bool): 起動している場合は true です。
//   - (error): 判定に失敗した場合のエラーです。
func isKabuStationRunning() (bool, error) {
	out, err := exec.Command("tasklist.exe", "/FI", "IMAGENAME eq KabuS.exe").CombinedOutput()
	if err != nil {
		return false, err
	}
	return strings.Contains(strings.ToLower(string(out)), "kabus.exe"), nil
}

// ----------------------------------------

// resolveTradeWebAppURL は、TradeWebApp の URL を解決します。
//
// 機能:
//   - 設定ファイル（PATH.TRADEWEBAPP_URL）→環境変数（TRADEWEBAPP_URL）→既定 URL の順に解決します。
//
// 引数およびその型:
//   - なし
//
// 返り値およびその型:
//   - (string): TradeWebApp の URL です。
func resolveTradeWebAppURL() string {
	candidate := strings.TrimSpace(cfg.Path.TradeWebAppURL)
	if candidate != "" {
		return candidate
	}

	candidate = strings.TrimSpace(os.Getenv("TRADEWEBAPP_URL"))
	if candidate != "" {
		return candidate
	}

	return "http://localhost:5173/"
}

// ----------------------------------------

// runPowerShellFile は、PowerShell スクリプトを実行し、標準出力と標準エラーをまとめて返します。
//
// 機能:
//   - `powershell.exe -NoProfile -ExecutionPolicy Bypass -File <script> ...` を実行します。
//
// 引数およびその型:
//   - scriptPath (string): スクリプトパスです。
//   - args (...string): スクリプトへ渡す引数です。
//
// 返り値およびその型:
//   - (string): 標準出力と標準エラーを連結した文字列です。
//   - (error): 実行に失敗した場合のエラーです。
func runPowerShellFile(scriptPath string, args ...string) (string, error) {
	commandArgs := []string{"-NoProfile", "-ExecutionPolicy", "Bypass", "-File", scriptPath}
	commandArgs = append(commandArgs, args...)

	out, err := exec.Command("powershell.exe", commandArgs...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// ----------------------------------------

// openURL は、既定ブラウザで指定 URL を開きます。
//
// 機能:
//   - Windows の `rundll32 url.dll,FileProtocolHandler` を用いて URL を開きます。
//
// 引数およびその型:
//   - url (string): 開く URL です。
//
// 返り値およびその型:
//   - (error): 失敗した場合のエラーです。
func openURL(url string) error {
	return exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", url).Start()
}
