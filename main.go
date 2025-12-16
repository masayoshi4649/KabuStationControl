package main

import (
	"errors"
	"flag"
	"io"
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
	flag.StringVar(&confPath, "c", "auth.toml", "設定ファイルのパス")
	flag.StringVar(&confPath, "config", "auth.toml", "設定ファイルのパス（別名）")
	flag.Parse()

	var err error
	cfg, err = loadConfig(confPath)
	if err != nil {
		os.Exit(1)
	}

	gin.SetMode(gin.ReleaseMode)
	rt := gin.New()
	rt.Use(gin.RecoveryWithWriter(io.Discard))

	rt.LoadHTMLGlob("view/*.html")
	rt.Static("/static", "./view")

	rt.GET("/", handleIndexGET)
	rt.GET("/bootauthkabus", handleBootAuthKabusGET)
	rt.GET("/bootapp", handleBootAppGET)

	if err := rt.Run(httpListenAddr); err != nil {
		os.Exit(1)
	}
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

	runErr := runPowerShellFile(
		filepath.Join("cmd", "Click-KabuStationLogin.ps1"),
		"-ExePath", exePath,
		"-TimeoutSeconds", "60",
	)
	if runErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":      false,
			"message": "KabuStation 起動認証（ログイン操作）の実行に失敗しました",
			"error":   runErr.Error(),
			"started": started,
		})
		return
	}

	apikey := auth()
	if apikey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":      false,
			"message": "KabuStation API認証（ログイン操作）の実行に失敗しました",
			"error":   "Please check config",
			"started": started,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"message": "KabuStation 起動認証（ログイン操作）を実行しました",
		"started": started,
	})
}

// ----------------------------------------

// handleBootAppGET は、TradeWebApp の起動要求を処理します。
//
// 機能:
//   - 設定ファイルの TRADEAPP.PATH を実行します。
//
// 引数およびその型:
//   - c (*gin.Context): Gin のコンテキストです。
//
// 返り値およびその型:
//   - なし（HTTP レスポンスとして JSON を返します）
func handleBootAppGET(c *gin.Context) {

	exePath, exeArgs, err := resolveTradeAppExePath()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": err.Error()})
		return
	}

	if err := exec.Command(exePath, exeArgs...).Start(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": "TradeWebApp の起動に失敗しました", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "message": "TradeWebApp を起動しました"})
}

// ----------------------------------------

// resolveKabuStationExePath は、KabuStation の実行ファイルパスを解決します。
//
// 機能:
//   - 設定ファイル（KABUS.PATH）のみを参照し、実行ファイルの存在を確認します。
//
// 引数およびその型:
//   - なし
//
// 返り値およびその型:
//   - (string): 実行ファイルパスです。
//   - (error): 見つからない場合や不正な場合のエラーです。
func resolveKabuStationExePath() (string, error) {
	exePath := strings.TrimSpace(cfg.Kabus.Path)
	if exePath == "" {
		return "", errors.New("KabuStation の実行ファイルパスが空です。設定ファイルの KABUS.PATH を設定してください。")
	}

	if _, err := os.Stat(exePath); err != nil {
		return "", errors.New("KabuStation の実行ファイルが見つかりません。設定ファイルの KABUS.PATH を確認してください。")
	}

	return exePath, nil
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

// resolveTradeAppExePath は、TradeWebApp の実行ファイルパスと引数を解決します。
//
// 機能:
//   - 設定ファイル（TRADEAPP.PATH）を参照し、実行ファイルの存在を確認します。
//   - 設定ファイル（TRADEAPP.CONF）が指定されている場合は、`-c <CONF>` を引数に付与します。
//
// 引数およびその型:
//   - なし
//
// 返り値およびその型:
//   - (string): 実行ファイルパスです。
//   - ([]string): 実行時引数です。
//   - (error): 見つからない場合や不正な場合のエラーです。
func resolveTradeAppExePath() (string, []string, error) {
	exePath := strings.TrimSpace(cfg.TradeApp.Path)
	if exePath == "" {
		return "", nil, errors.New("TradeWebApp の実行ファイルパスが空です。設定ファイルの TRADEAPP.PATH を設定してください。")
	}

	if _, err := os.Stat(exePath); err != nil {
		return "", nil, errors.New("TradeWebApp の実行ファイルが見つかりません。設定ファイルの TRADEAPP.PATH を確認してください。")
	}

	confPath := strings.TrimSpace(cfg.TradeApp.Conf)
	if confPath == "" {
		return exePath, nil, nil
	}

	if _, err := os.Stat(confPath); err != nil {
		return "", nil, errors.New("TradeWebApp の設定ファイルが見つかりません。設定ファイルの TRADEAPP.CONF を確認してください。")
	}

	return exePath, []string{"-c", confPath}, nil
}

// ----------------------------------------

// runPowerShellFile は、PowerShell スクリプトを実行します。
//
// 機能:
//   - `powershell.exe -NoProfile -ExecutionPolicy Bypass -File <script> ...` を実行します。
//   - 標準出力と標準エラーは破棄します。
//
// 引数およびその型:
//   - scriptPath (string): スクリプトパスです。
//   - args (...string): スクリプトへ渡す引数です。
//
// 返り値およびその型:
//   - (error): 実行に失敗した場合のエラーです。
func runPowerShellFile(scriptPath string, args ...string) error {
	commandArgs := []string{"-NoProfile", "-ExecutionPolicy", "Bypass", "-File", scriptPath}
	commandArgs = append(commandArgs, args...)

	cmd := exec.Command("powershell.exe", commandArgs...)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run()
}
