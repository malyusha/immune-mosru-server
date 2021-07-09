package bot

import (
	"fmt"

	"github.com/skip2/go-qrcode"
)

type qrGenerator struct {
	urlPattern string
}

func NewQRGenerator(pattern string) *qrGenerator {
	return &qrGenerator{urlPattern: pattern}
}

func (gen *qrGenerator) GenerateQR(code string) ([]byte, error) {
	qr, err := qrcode.Encode(fmt.Sprintf(gen.urlPattern, code), qrcode.Medium, 256)
	if err != nil {
		return nil, fmt.Errorf("failed to create QR code for code %s: %w", code, err)
	}

	return qr, nil
}
