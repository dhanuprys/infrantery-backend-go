// Package compression provides zstd compression utilities.
// The encoder and decoder are initialized once and reused across calls,
// as they are safe for concurrent use with EncodeAll/DecodeAll.
package compression

import (
	"fmt"
	"sync"

	"github.com/klauspost/compress/zstd"
)

var (
	encoder     *zstd.Encoder
	decoder     *zstd.Decoder
	encoderOnce sync.Once
	decoderOnce sync.Once
	initEncErr  error
	initDecErr  error
)

func getEncoder() (*zstd.Encoder, error) {
	encoderOnce.Do(func() {
		encoder, initEncErr = zstd.NewWriter(nil,
			zstd.WithEncoderLevel(zstd.SpeedDefault),
		)
	})
	return encoder, initEncErr
}

func getDecoder() (*zstd.Decoder, error) {
	decoderOnce.Do(func() {
		decoder, initDecErr = zstd.NewReader(nil,
			zstd.WithDecoderMaxMemory(256*1024*1024), // 256 MB limit
		)
	})
	return decoder, initDecErr
}

// Compress compresses data using zstd with the default compression level.
func Compress(data []byte) ([]byte, error) {
	enc, err := getEncoder()
	if err != nil {
		return nil, fmt.Errorf("initializing zstd encoder: %w", err)
	}
	return enc.EncodeAll(data, make([]byte, 0, len(data)/2)), nil
}

// Decompress decompresses zstd-compressed data.
func Decompress(data []byte) ([]byte, error) {
	dec, err := getDecoder()
	if err != nil {
		return nil, fmt.Errorf("initializing zstd decoder: %w", err)
	}
	result, err := dec.DecodeAll(data, nil)
	if err != nil {
		return nil, fmt.Errorf("decompressing data: %w", err)
	}
	return result, nil
}
