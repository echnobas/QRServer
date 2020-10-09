package main

import (
	"encoding/json"
	"fmt"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/http"
	"os"
)

const maxUploadSize = 10 << 20 // 10mb

type QRJsonRes struct {
	Err *string `json:"error"`
	Result *string `json:"result"`
}

func WriteErr(w http.ResponseWriter, errMsg string, resCode int) {
	w.WriteHeader(resCode)
	w.Header().Set("Content-Type", "application/json")
	response := QRJsonRes{Err: &errMsg, Result: nil}
	data, err := json.Marshal(response)
	if err != nil { WriteErr(w, err.Error(), 500); return }
	_, _ = w.Write(data)
}

func ReadQRCode(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		WriteErr(w, fmt.Sprintf("Could not parse multipart form: %v\n", err), 500)
		return
	}
	file, handler, err := r.FormFile("file")
	if err != nil { WriteErr(w, "Error Retrieving the File", 400); return }
	defer func(){ _ = file.Close() }()
	_, _ = file.Seek(0,0)
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)
	defer func() { _ = file.Close() }()

	img, _, err := image.Decode(file)
	if err != nil { WriteErr(w, err.Error(), 500); return }
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil { WriteErr(w, err.Error(), 500); return }
	qrReader := qrcode.NewQRCodeReader()
	result, err := qrReader.Decode(bmp, nil)
	if err != nil { WriteErr(w, err.Error(), 400); return }
	text := result.String()
	response := QRJsonRes{
		Err:    nil,
		Result: &text,
	}
	jsonResponse, err := json.MarshalIndent(response, "", "\t")
	if err != nil { WriteErr(w, err.Error(), 400); return }
	_, _ = w.Write(jsonResponse)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}
	fmt.Println("Starting server")
	http.HandleFunc("/api/v1/read", ReadQRCode)
	http.HandleFunc("/",
		func(w http.ResponseWriter, r *http.Request) { _,_ = w.Write([]byte("Base")) },
	)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

