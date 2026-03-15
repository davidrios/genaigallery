package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"genai-gallery-backend/internal/config"
	"genai-gallery-backend/internal/database"
	"genai-gallery-backend/internal/handlers"
)

func serveStaticZip(r *gin.Engine) bool {
	exePath, err := os.Executable()
	if err != nil {
		return false
	}

	exe := filepath.Base(exePath)
	if strings.Index(exe, "genaigallery") != 0 {
		return false
	}

	file, err := os.Open(exePath)
	if err != nil {
		log.Fatalln("Could not read self executable")
		return false
	}

	info, err := file.Stat()
	if err != nil {
		log.Fatalln("Could not stat self executable")
		return false
	}

	reader, err := zip.NewReader(file, info.Size())
	if err != nil {
		log.Fatalf("error reading self zip: %v\n", err)
		return false
	}

	rootFs, err := fs.Sub(reader, ".")
	if err != nil {
		log.Fatalf("error creating FS for self zip: %v\n", err)
		return false
	}
	rootHttpFs := http.FS(rootFs)

	r.StaticFileFS("/favicon.ico", "favicon.ico", rootHttpFs)

	f, err := rootFs.Open("index.html")
	if err != nil {
		log.Fatalf("index.html not found in zip")
		return false
	}
	defer f.Close()

	indexContent, err := io.ReadAll(f)
	if err != nil {
		log.Fatalf("error reading index.html")
		return false
	}

	r.GET("/", func(c *gin.Context) {
		c.Data(200, "text/html; charset=utf-8", indexContent)
	})

	assetsFs, err := fs.Sub(reader, "assets")
	if err != nil {
		log.Fatalf("error creating FS for self zip: %v\n", err)
		return false
	}

	r.StaticFS("/assets", http.FS(assetsFs))

	log.Println("Serving static files from self zip")
	return true
}

func main() {
	config.InitConfig()
	database.InitDB(config.DBPath)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:" + config.Port},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type"},
		AllowCredentials: true,
	}))

	r.Static(config.StaticImagesRoot, config.ImagesDir)

	api := r.Group("/api")
	{
		api.GET("/image/:id", handlers.GetImage)
		api.GET("/browse", handlers.Browse)
		api.POST("/upload", handlers.Upload)
	}

	servingZip := serveStaticZip(r)

	// Determine the port to start with
	startPort := 5775
	if config.Port != "" {
		// If the user specified a port explicitly, use that exactly
		if p, err := strconv.Atoi(config.Port); err == nil {
			startPort = p
		}
	}

	var listener net.Listener
	var err error
	var finalPort int

	for port := startPort; port < startPort+100; port++ {
		addr := fmt.Sprintf(":%d", port)
		listener, err = net.Listen("tcp", addr)
		if err == nil {
			finalPort = port
			break
		}
		// If user explicitly asked for a port and it failed, don't try next ports
		if config.Port != "" {
			break
		}
	}

	if listener == nil {
		log.Fatalf("Could not find a free port to start the server. Last error: %v", err)
	}

	config.Port = strconv.Itoa(finalPort)
	serverUrl := fmt.Sprintf("http://localhost:%d", finalPort)
	log.Printf("Server starting on %s", serverUrl)

	go func() {
		if err := r.RunListener(listener); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	if servingZip {
		time.Sleep(100 * time.Millisecond)
		openBrowser(serverUrl)
	}

	// Block main thread to keep server alive
	select {}
}

func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		log.Printf("Failed to open browser automatically: %v", err)
	}
}
