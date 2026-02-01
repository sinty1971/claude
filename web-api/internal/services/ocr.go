package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"web-api/internal/core"
)

type OcrService struct {
	// name はサービス名
	name string
}

func NewOcrService(cs *ContainerService) *OcrService {
	return &OcrService{
		name: "OcrService",
	}
}

func (srv *OcrService) Name() string {
	return srv.name
}

// GenerateHandler はサービスのハンドラを生成します
func (srv *OcrService) GenerateHandler() (
	servicePath string,
	handler http.Handler,
	serviceName string) {

	// gRPC パスとハンドラの生成
	servicePath = "/api/ocr"
	handler = http.HandlerFunc(ocrHandler)

	// サービス名の取得
	serviceName = strings.Join([]string{"api", srv.name}, ".")

	return
}

func (srv *OcrService) Start() error {
	// 初期データのロード処理などをここに実装予定
	return nil
}

func (srv *OcrService) Cleanup() {
	// クリーンアップ処理などをここに実装予定
}

func ocrHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}
	log.Println("Received OCR request")

	// フォームデータからファイルを取得
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Failed to get file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 一時ファイルを作成 (元の拡張子を保持)
	tempFile, err := os.CreateTemp("", "ocr-upload-*"+filepath.Ext(header.Filename))
	if err != nil {
		http.Error(w, "Failed to create temporary file", http.StatusInternalServerError)
		return
	}
	// このdeferは重要: PDF->PNG変換後も元のPDF一時ファイルが削除されるようにする
	defer os.Remove(tempFile.Name())

	// アップロードされたファイルの内容を一時ファイルにコピー
	_, err = io.Copy(tempFile, file)
	if err != nil {
		tempFile.Close() // コピーエラーでもファイルを閉じる
		http.Error(w, "Failed to save uploaded file", http.StatusInternalServerError)
		return
	}
	tempFile.Close() // これ以降はファイルパスで操作するため、ファイルを閉じる

	// OCR処理に渡す最終的な画像ファイルのパス
	ocrImagePath := tempFile.Name()

	// PDFかどうかを判定し、画像に変換する
	if filepath.Ext(strings.ToLower(header.Filename)) == ".pdf" {
		// 出力PNGファイルのパスプレフィックスを定義
		// (例: C:\Users\...\ocr-upload-123.pdf -> C:\Users\...\ocr-upload-123.pdf)
		pngFilePrefix := strings.TrimSuffix(tempFile.Name(), filepath.Ext(tempFile.Name()))

		log.Printf("Converting PDF %s to PNG...", tempFile.Name())
		// pdftoppm を使ってPDFの最初の1ページをPNG画像に変換
		// -f 1: 最初のページ
		// -l 1: 最初のページまで (範囲指定)
		// -singlefile: ファイル名にページ番号(-1など)をつけない
		// -png: PNG形式で出力
		cmd := exec.Command(core.Config.PdfToImageExecutablePath, "-f", "1", "-l", "1", "-singlefile", "-png", tempFile.Name(), pngFilePrefix)

		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			errMsg := fmt.Sprintf("Failed to convert PDF to PNG. Make sure 'poppler' is installed and in your PATH. Error: %v, Stderr: %s", err, stderr.String())
			log.Println(errMsg)
			http.Error(w, errMsg, http.StatusInternalServerError)
			return
		}

		// 変換後のPNGファイルパスをOCR処理用に設定
		ocrImagePath = pngFilePrefix + ".png"
		if _, err := os.Stat(ocrImagePath); os.IsNotExist(err) {
			errMsg := fmt.Sprintf("Converted PNG file not found at expected path: %s", ocrImagePath)
			log.Println(errMsg)
			http.Error(w, errMsg, http.StatusInternalServerError)
			return
		}

		// 変換されたPNGファイルも後で削除する
		defer os.Remove(ocrImagePath)
		log.Printf("Converted PDF to %s", ocrImagePath)
	}

	// 画像の向き補正（EXIF Orientation に従う）
	// if err := fixImageOrientation(ocrImagePath); err != nil {
	// 	log.Printf("fixImageOrientation warning: %v", err)
	// }

	// 1リクエストごとにOCRプロセスを起動
	result, err := recognizeOnce(core.Config.OcrServerExecutablePath, ocrImagePath)
	if err != nil {
		log.Printf("OCR error: %v", err)
		// エラー詳細をクライアントに返す
		errorDetails := fmt.Sprintf(`{"error": "OCR failed", "details": "%s"}`, err.Error())
		http.Error(w, errorDetails, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	// result は JSON 文字列 (日本語化済み) のはずなのでパースして
	// `text` フィールドをすべて収集して返す。見つからなければ元の result を返す。
	var parsed interface{}
	if err := json.Unmarshal([]byte(result), &parsed); err == nil {
		texts := make([]string, 0)
		var collectTexts func(interface{})
		collectTexts = func(v interface{}) {
			switch vv := v.(type) {
			case map[string]interface{}:
				for k, uv := range vv {
					if k == "text" {
						switch tv := uv.(type) {
						case string:
							if tv != "" {
								texts = append(texts, tv)
							}
						case []interface{}:
							for _, e := range tv {
								if s, ok := e.(string); ok && s != "" {
									texts = append(texts, s)
								}
							}
						default:
							// 非文字列は文字列化して追加
							if b, err := json.Marshal(uv); err == nil {
								texts = append(texts, string(b))
							}
						}
					} else {
						collectTexts(uv)
					}
				}
			case []interface{}:
				for _, item := range vv {
					collectTexts(item)
				}
			}
		}

		collectTexts(parsed)
		if len(texts) > 0 {
			out := map[string]interface{}{"text": texts}
			if b, err := json.Marshal(out); err == nil {
				_, _ = w.Write(b)
				return
			}
		}
	}

	// パース失敗や text 未検出時は元のレスポンスをそのまま返す
	_, _ = w.Write([]byte(result))
}

// 1リクエストごとにOCRプロセスを起動
func recognizeOnce(exePath, imagePath string) (string, error) {
	// --image_path <path> 形式でコマンドを実行
	args := []string{"-image_path", imagePath}
	// 設定で言語が指定されていれば渡す（例: "japan"）
	if core.Config != nil && core.Config.OcrLanguage != "" {
		args = append(args, "-use_angle_cls", "true")
		args = append(args, "-cls", "true")
		args = append(args, "-rec_model_dir", "models/japan_PP-OCRv3_rec_infer")
		args = append(args, "-rec_char_dict_path", "models/dict_japan.txt")
	}
	cmd := exec.Command(exePath, args...)
	// OCR実行ファイルのディレクトリを作業ディレクトリに設定
	cmd.Dir = filepath.Dir(exePath)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("OCR error: %v, stderr: %s", err, errBuf.String())
	}

	// OCR生出力をログ出力
	// 日本語を可読化するため、JSONを一度デコード→再エンコード
	var v map[string]interface{}
	raw := outBuf.Bytes()

	// OCRツールの出力には、JSONの前に情報メッセージが含まれる場合があるため、
	// JSONオブジェクトの開始を示す最初の '{' を探す。
	jsonStart := bytes.IndexByte(raw, '{')
	if jsonStart == -1 {
		log.Printf("No JSON object found in OCR output: %s", string(raw))
		return outBuf.String(), nil
	}

	// '{' 以降をJSONデータとしてパースする
	if err := json.Unmarshal(raw[jsonStart:], &v); err != nil {
		// JSONでなければそのまま返す
		log.Printf("json.Unmarshal failed: %v", err)
		return outBuf.String(), nil
	}

	// Goのjson.MarshalはデフォルトでUnicode文字を \uXXXX 形式にエスケープする。
	// そのため、一度Marshal処理を行った後、エスケープされた文字列をUnquoteして元の文字に戻す。
	pretty, err := json.Marshal(v)
	if err != nil {
		// Marshalに失敗した場合は、デコード前のJSONを返す
		return outBuf.String(), nil
	}

	finalResult := unquoteUnicode(string(pretty))
	return finalResult, nil
}

var reUnicode = regexp.MustCompile(`\\\\u[0-9a-fA-F]{4}`)

// unquoteUnicodeは、JSON文字列内の \uXXXX エスケープシーケンスを

// 対応するUTF-8文字に変換します。

func unquoteUnicode(s string) string {

	return reUnicode.ReplaceAllStringFunc(s, func(match string) string {

		// `match` は `\uXXXX` 形式

		// `match[2:]` で `XXXX` の部分を取り出す

		code, err := strconv.ParseInt(match[2:], 16, 32)

		if err != nil {

			// パースに失敗した場合は元の文字列を返す

			return match

		}

		// 16進数のコードポイントをruneに変換し、さらにstringに変換する

		return string(rune(code))

	})

}
