package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const httpListenAddr = ":3000"

var (
	pidKabus    int = 0
	pidTradeApp int = 0
)

var cfg Config

var apiKey string
var apiKeyMu sync.RWMutex

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

	rt := gin.Default()
	rt.LoadHTMLGlob("view/*.html")
	rt.Static("/static", "./view")
	rt.GET("/", handleIndexGET)
	rt.GET("/bootauthkabus", handleBootAuthKabusGET)
	rt.GET("/apiauth", handleAPIAuthGET)
	rt.GET("/bootapp", handleBootAppGET)
	rt.GET("/pid", handlePIDGET)

	if err := rt.Run(httpListenAddr); err != nil {
		os.Exit(1)
	}
}

// ----------------------------------------

func handleIndexGET(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{})
}

// ----------------------------------------

// handlePIDGET は、pidKabus と pidTradeApp の値を返します。
//
// 機能:
//   - pidKabus と pidTradeApp の現在値を JSON で返します。
//
// 引数およびその型:
//   - c (*gin.Context): Gin のコンテキストです。
//
// 返り値およびその型:
//   - なし（HTTP レスポンスとして JSON を返します）
func handlePIDGET(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"ok":          true,
		"pidKabus":    pidKabus,
		"pidTradeApp": pidTradeApp,
	})
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
	pid := 0
	var pids []int
	if !wasRunning {
		cmd := exec.Command(exePath)
		cmd.Dir = filepath.Dir(exePath)
		if err := cmd.Start(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": "KabuStation の起動に失敗しました", "error": err.Error()})
			return
		}
		pid = cmd.Process.Pid
		_ = cmd.Process.Release()
		started = true
	} else {
		foundPIDs, pidErr := getProcessPIDsByImageName("KabuS.exe")
		if pidErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": "KabuStation の PID 取得に失敗しました", "error": pidErr.Error()})
			return
		}
		pids = foundPIDs
		if len(pids) > 0 {
			pid = pids[0]
		}
	}

	if started {
		time.Sleep(10 * time.Second)
	}

	pidKabus = pid

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
			"pid":     pid,
			"pids":    pids,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"message": "KabuStation 起動認証（ログイン操作）を実行しました",
		"started": started,
		"pid":     pid,
		"pids":    pids,
	})
}

// ----------------------------------------

// handleAPIAuthGET は、KabuStation の API 認証要求を処理します。
//
// 機能:
//   - KabuStation の API トークンを取得し、サーバ内へ保持します。
//
// 引数およびその型:
//   - c (*gin.Context): Gin のコンテキストです。
//
// 返り値およびその型:
//   - なし（HTTP レスポンスとして JSON を返します）
func handleAPIAuthGET(c *gin.Context) {
	apikey := apitoken()
	if strings.TrimSpace(apikey) == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": "API 認証に失敗しました"})
		return
	}

	if cfg.System.Debug {
		log.Println(apikey) // DEBUG
	}

	setAPIKey(apikey)
	c.JSON(http.StatusOK, gin.H{"ok": true, "message": "API 認証が完了しました"})
}

// ----------------------------------------

// handleBootAppGET は、TradeWebApp の起動要求を処理します。
//
// 機能:
//   - `TRADEAPP.PATH -c TRADEAPP.CONF -k <APIKEY>` で起動します。
//
// 引数およびその型:
//   - c (*gin.Context): Gin のコンテキストです。
//
// 返り値およびその型:
//   - なし（HTTP レスポンスとして JSON を返します）
func handleBootAppGET(c *gin.Context) {

	apikey := getAPIKey()
	if strings.TrimSpace(apikey) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "message": "API認証を先に実行してください"})
		return
	}

	exePath, exeArgs, err := resolveTradeAppExeArgs(apikey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": err.Error()})
		return
	}

	cmd := exec.Command(exePath, exeArgs...)
	cmd.Dir = filepath.Dir(exePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": "TradeWebApp の起動に失敗しました", "error": err.Error()})
		return
	}

	pid := cmd.Process.Pid
	_ = cmd.Process.Release()
	pidTradeApp = pid

	time.Sleep(500 * time.Millisecond)
	if !isProcessAliveByPID(pid) {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": "TradeWebApp が起動直後に終了しました", "pid": pid})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "message": "TradeWebApp を起動しました", "pid": pid})
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

// getProcessPIDsByImageName は、指定した実行ファイル名に一致する PID 一覧を取得します。
//
// 機能:
//   - Windows の tasklist を用いて、対象プロセスの PID を列挙します。
//
// 引数およびその型:
//   - imageName (string): 例）KabuS.exe のようなプロセス名です。
//
// 返り値およびその型:
//   - ([]int): 取得した PID の一覧です（見つからない場合は空配列です）。
//   - (error): 取得や解析に失敗した場合のエラーです。
func getProcessPIDsByImageName(imageName string) ([]int, error) {
	query := strings.TrimSpace(imageName)
	if query == "" {
		return nil, errors.New("プロセス名が空です")
	}

	out, err := exec.Command("tasklist.exe", "/FI", fmt.Sprintf("IMAGENAME eq %s", query), "/FO", "CSV", "/NH").CombinedOutput()
	if err != nil {
		return nil, err
	}

	text := strings.TrimSpace(string(out))
	if text == "" {
		return []int{}, nil
	}
	if strings.HasPrefix(text, "INFO:") {
		return []int{}, nil
	}

	reader := csv.NewReader(strings.NewReader(text))
	reader.FieldsPerRecord = -1

	var pids []int
	for {
		record, readErr := reader.Read()
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return nil, readErr
		}
		if len(record) < 2 {
			continue
		}

		if !strings.EqualFold(strings.TrimSpace(record[0]), query) {
			continue
		}

		pid, convErr := strconv.Atoi(strings.TrimSpace(record[1]))
		if convErr != nil {
			continue
		}
		pids = append(pids, pid)
	}

	return pids, nil
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

// setAPIKey は、API トークンをサーバ内へ保持します。
//
// 機能:
//   - 取得したトークンをグローバルへ設定します。
//
// 引数およびその型:
//   - key (string): API トークンです。
//
// 返り値およびその型:
//   - なし
func setAPIKey(key string) {
	apiKeyMu.Lock()
	defer apiKeyMu.Unlock()

	apiKey = key
}

// ----------------------------------------

// getAPIKey は、保持済みの API トークンを取得します。
//
// 機能:
//   - サーバ内へ保持した API トークンを返します。
//
// 引数およびその型:
//   - なし
//
// 返り値およびその型:
//   - (string): API トークンです（未設定の場合は空文字です）。
func getAPIKey() string {
	apiKeyMu.RLock()
	defer apiKeyMu.RUnlock()

	return apiKey
}

// ----------------------------------------

// resolveTradeAppExeArgs は、TradeWebApp の実行ファイルパスと引数を解決します。
//
// 機能:
//   - 設定ファイル（TRADEAPP.PATH）を参照し、実行ファイルの存在を確認します。
//   - `-c <CONF> -k <APIKEY>` を引数に付与します。
//
// 引数およびその型:
//   - apikey (string): API トークンです。
//
// 返り値およびその型:
//   - (string): 実行ファイルパスです。
//   - ([]string): 実行時引数です。
//   - (error): 見つからない場合や不正な場合のエラーです。
func resolveTradeAppExeArgs(apikey string) (string, []string, error) {
	exePath := strings.TrimSpace(cfg.TradeApp.Path)
	if exePath == "" {
		return "", nil, errors.New("TradeWebApp の実行ファイルパスが空です。設定ファイルの TRADEAPP.PATH を設定してください。")
	}

	if _, err := os.Stat(exePath); err != nil {
		return "", nil, errors.New("TradeWebApp の実行ファイルが見つかりません。設定ファイルの TRADEAPP.PATH を確認してください。")
	}

	confPath := strings.TrimSpace(cfg.TradeApp.Conf)
	if confPath == "" {
		return "", nil, errors.New("TradeWebApp の設定ファイルパスが空です。設定ファイルの TRADEAPP.CONF を設定してください。")
	}

	if _, err := os.Stat(confPath); err != nil {
		return "", nil, errors.New("TradeWebApp の設定ファイルが見つかりません。設定ファイルの TRADEAPP.CONF を確認してください。")
	}

	if strings.TrimSpace(apikey) == "" {
		return "", nil, errors.New("API トークンが空です。API認証を先に実行してください。")
	}

	return exePath, []string{"-c", confPath, "-k", apikey}, nil
}

// ----------------------------------------

// isProcessAliveByPID は、指定 PID のプロセスが起動中かどうかを判定します。
//
// 機能:
//   - Windows の tasklist を用いて PID の存在を確認します。
//
// 引数およびその型:
//   - pid (int): 判定対象の PID です。
//
// 返り値およびその型:
//   - (bool): 起動している場合は true です。
func isProcessAliveByPID(pid int) bool {
	out, err := exec.Command("tasklist.exe", "/FI", fmt.Sprintf("PID eq %d", pid)).CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), fmt.Sprintf(" %d ", pid))
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
