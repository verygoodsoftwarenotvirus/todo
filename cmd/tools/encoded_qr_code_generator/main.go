package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"log"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

const (
	base64ImagePrefix = "data:image/png;base64,"
	otpAuthURL        = `otpauth://totp/todo:username?secret=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=&issuer=todo`
)

func main() {
	bmp, err := qrcode.NewQRCodeWriter().EncodeWithoutHint(otpAuthURL, gozxing.BarcodeFormat_QR_CODE, 128, 128)
	if err != nil {
		log.Fatal(err)
	}

	// encode the QR code to PNG.
	var b bytes.Buffer
	if err = png.Encode(&b, bmp); err != nil {
		log.Fatal(err)
	}

	// base64 encode the image for easy HTML use.
	log.Println(fmt.Sprintf("%s%s", base64ImagePrefix, base64.StdEncoding.EncodeToString(b.Bytes())))
}
