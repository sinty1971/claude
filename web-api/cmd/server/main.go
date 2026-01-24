package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"golang.org/x/term"

	"web-api/internal/core"
	"web-api/internal/services"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	// サーバー設定の初期化
	parseConfiguration()

	// コマンドライン引数の解析
	parseCommandLineFlags()

	// コンテナサービスの作成
	cs := services.NewContainerService()
	defer cs.CleanupAll()

	// 各サービスの作成と登録
	cs.AddService(services.NewDirectoryService(cs))
	cs.AddService(services.NewCompanyCategoryService(cs))
	cs.AddService(services.NewCompanyService(cs))
	cs.AddService(services.NewKojiService(cs))
	cs.AddService(services.NewMemberService(cs))
	cs.AddService(services.NewOcrService(cs))

	// サービスのハンドラを取得してマルチプレクサに登録
	mux, err := cs.GenerateMux()
	if err != nil {
		log.Fatalf("Failed to generate mux: %v", err)
	}

	// サービスの起動
	err = cs.Start()
	if err != nil {
		log.Fatalf("Failed to start services: %v", err)
	}
	defer cs.CleanupAll()

	// HTTP サーバーの設定
	httpServer := &http.Server{
		Addr:    *httpAddr,
		Handler: h2c.NewHandler(cors(mux), &http2.Server{}),
	}

	// HTTP サーバーの起動
	go func() {
		log.Printf("HTTP gRPC サーバーを %s で起動します", httpServer.Addr)
		err := httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP サーバーでエラーが発生しました: %v", err)
		}
	}()

	// HTTPS サーバーの設定と起動（有効な場合）
	var httpsServer *http.Server
	if *enableTLS {
		if err := ensureCertificate(*certPath, *keyPath); err != nil {
			log.Fatalf("TLS 証明書の準備に失敗しました: %v", err)
		}

		httpsServer = &http.Server{
			Addr:      *httpsAddr,
			Handler:   cors(mux),
			TLSConfig: &tls.Config{MinVersion: tls.VersionTLS12},
		}

		go func() {
			log.Printf("HTTPS gRPC サーバーを %s で起動します", httpsServer.Addr)
			err := httpsServer.ListenAndServeTLS(*certPath, *keyPath)
			if err != nil && err != http.ErrServerClosed {
				log.Fatalf("HTTPS サーバーでエラーが発生しました: %v", err)
			}
		}()
	}

	// シャットダウンの待機
	shutdown(httpServer, httpsServer)
}

// parseConfiguration はサーバーの設定を解析して初期化します
func parseConfiguration() {
	// サーバーのデフォルト設定を定義します。
	core.Config.DirectoryBaseDirPath = "{ROOT}"
	core.Config.CompanyBaseDirPath = "{ROOT}/1 会社"
	core.Config.CompanyWatcherMaxDepth = 0
	core.Config.KojiBaseDirPath = "{ROOT}/2 工事"
	core.Config.KojiWatcherMaxDepth = 0
	core.Config.MemberBaseDirPath = "{ROOT}/1 会社"
	core.Config.MemberWatcherMaxDepth = 0
	core.Config.MaximumWorkers = 16
	core.Config.MinimumWorkers = 2
	core.Config.CpuMultiplier = 2
	core.Config.OcrServerExecutablePath = "{USERPROFILE}/.local/share/PaddleOCR-json_v1.4.1/paddleocr-json.exe"
	core.Config.PdfToImageExecutablePath = "{USERPROFILE}/.local/share/poppler-25.12.0/Library/bin/pdftoppm.exe"
	// デフォルトで日本語モデルを優先する
	core.Config.OcrLanguage = "japan"

	// プレースホルダーの展開
	core.Config.ExpandPlaceholders()
}

// コマンドライン引数
var (
	httpAddr  = flag.String("http-addr", ":9090", "HTTP (h2c) の待受アドレス")
	httpsAddr = flag.String("https-addr", ":9443", "HTTPS の待受アドレス")
	enableTLS = flag.Bool("enable-tls", false, "true の場合は HTTPS も起動します")
	certPath  = flag.String("cert", "cert.pem", "TLS 証明書のパス")
	keyPath   = flag.String("key", "key.pem", "TLS 秘密鍵のパス")
)

// コマンドライン引数の処理
func parseCommandLineFlags() {
	log.Printf("Usage of %s:", os.Args[0])
	flag.PrintDefaults()
	// コマンドライン引数の解析
	flag.Parse()

}

// 自己署名証明書の生成
func ensureCertificate(certFile, keyFile string) error {
	if _, err := os.Stat(certFile); err == nil {
		if _, err := os.Stat(keyFile); err == nil {
			return nil
		}
	}

	if err := os.MkdirAll(filepath.Dir(certFile), 0o755); err != nil {
		return err
	}

	log.Printf("自己署名証明書を %s に生成します", certFile)

	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			CommonName:   "Pathist gRPC Server",
			Organization: []string{"Pathist gRPC Server"},
		},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
		IPAddresses: []net.IP{
			net.ParseIP("127.0.0.1"),
			net.ParseIP("::1"),
		},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	if err := writePEM(certFile, "CERTIFICATE", certDER); err != nil {
		return err
	}

	keyDER := x509.MarshalPKCS1PrivateKey(priv)
	if err := writePEM(keyFile, "RSA PRIVATE KEY", keyDER); err != nil {
		return err
	}

	return nil
}

func writePEM(path, typ string, data []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()

	return pem.Encode(file, &pem.Block{Type: typ, Bytes: data})
}

// CORS ミドルウェア
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Connect-Protocol-Version")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func shutdown(httpServer *http.Server, httpsServer *http.Server) {
	// シグナルの受信を待機するコンテキストを作成
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 標準入力で 'q' を受け取ったら停止する（Enter 不要: 端末が有効な場合は raw モードで1キー読み取り）
	go func() {
		fd := int(os.Stdin.Fd())
		// ターミナルでない場合は既存の行読み取りにフォールバック
		if !term.IsTerminal(fd) {
			buf := make([]byte, 1)
			for {
				n, err := os.Stdin.Read(buf)
				if err != nil || n == 0 {
					return
				}
				if buf[0] == 'q' || buf[0] == 'Q' {
					log.Printf("'q' を受信しました。シャットダウンを開始します。")
					stop()
					return
				}
			}
		}

		oldState, err := term.MakeRaw(fd)
		if err != nil {
			// raw モードにできなければフォールバック
			buf := make([]byte, 1)
			for {
				n, err := os.Stdin.Read(buf)
				if err != nil || n == 0 {
					return
				}
				if buf[0] == 'q' || buf[0] == 'Q' {
					log.Printf("'q' を受信しました。シャットダウンを開始します。")
					stop()
					return
				}
			}
		}
		defer term.Restore(fd, oldState)

		buf := make([]byte, 1)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil || n == 0 {
				return
			}
			if buf[0] == 'q' || buf[0] == 'Q' {
				log.Printf("'q' を受信しました。シャットダウンを開始します。")
				stop()
				return
			}
		}
	}()

	// シグナル受信を待機
	<-ctx.Done()
	log.Printf("停止シグナルを受信しました。サーバーをシャットダウンします。")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := httpServer.Shutdown(shutdownCtx)
	if err != nil {
		log.Printf("HTTP サーバーの停止に失敗しました: %v", err)
	}

	if httpsServer != nil {
		err = httpsServer.Shutdown(shutdownCtx)
		if err != nil {
			log.Printf("HTTPS サーバーの停止に失敗しました: %v", err)
		}
	}

	log.Printf("gRPC サーバーを停止しました。")

}
