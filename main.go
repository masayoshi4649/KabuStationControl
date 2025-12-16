package main

import (
	"bufio"
	"flag"
	"log"
	"net/http"
	"os"

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

	rt.Run(httpListenAddr)
}

func handleIndexGET(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{})
}
