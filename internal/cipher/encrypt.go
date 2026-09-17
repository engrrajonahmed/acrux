package cipher

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const nonceSize = 12

func Encrypt(source, destination, key string) error {
	if source == "" {
		return errors.New("source cannot be empty")
	}

	if destination == "" {
		return errors.New("destination cannot be empty")
	}

	if key == "" {
		return errors.New("encryption key cannot be empty")
	}

	if filepath.Clean(source) == filepath.Clean(destination) {
		return errors.New("source and destination are the same file")
	}

	sourceInfo, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("stat source %q: %w", source, err)
	}

	if !sourceInfo.Mode().IsRegular() {
		return fmt.Errorf("source %q is not a regular file", source)
	}

	if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		return fmt.Errorf("create destination directory: %w", err)
	}

	input, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("open source %q: %w", source, err)
	}
	defer input.Close()

	output, err := os.OpenFile(
		destination,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0600,
	)
	if err != nil {
		return fmt.Errorf("create encrypted file %q: %w", destination, err)
	}

	success := false
	defer func() {
		if !success {
			_ = output.Close()
			_ = os.Remove(destination)
		}
	}()

	if err := encryptStream(input, output, key); err != nil {
		return fmt.Errorf("encrypt %q: %w", source, err)
	}

	if err := output.Sync(); err != nil {
		return fmt.Errorf("sync encrypted file %q: %w", destination, err)
	}

	if err := output.Close(); err != nil {
		return fmt.Errorf("close encrypted file %q: %w", destination, err)
	}

	success = true

	return nil
}

func EncryptReader(reader io.Reader, writer io.Writer, key string) error {
	if reader == nil {
		return errors.New("reader cannot be nil")
	}

	if writer == nil {
		return errors.New("writer cannot be nil")
	}

	if key == "" {
		return errors.New("encryption key cannot be empty")
	}

	return encryptStream(reader, writer, key)
}

func encryptStream(reader io.Reader, writer io.Writer, key string) error {
	block, err := aes.NewCipher(deriveKey(key))
	if err != nil {
		return fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCMWithNonceSize(block, nonceSize)
	if err != nil {
		return fmt.Errorf("create GCM cipher: %w", err)
	}

	nonce := make([]byte, nonceSize)

	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("generate encryption nonce: %w", err)
	}

	if _, err := writer.Write(nonce); err != nil {
		return fmt.Errorf("write encryption nonce: %w", err)
	}

	const chunkSize = 1024 * 1024
	buffer := make([]byte, chunkSize)

	for {
		n, readErr := reader.Read(buffer)

		if n > 0 {
			encrypted := gcm.Seal(nil, nonce, buffer[:n], nil)

			if _, err := writer.Write(encrypted); err != nil {
				return fmt.Errorf("write encrypted data: %w", err)
			}

			incrementNonce(nonce)
		}

		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}

			return fmt.Errorf("read source data: %w", readErr)
		}
	}

	return nil
}

func deriveKey(key string) []byte {
	hash := sha256.Sum256([]byte(key))
	return hash[:]
}

func incrementNonce(nonce []byte) {
	for i := len(nonce) - 1; i >= 0; i-- {
		nonce[i]++

		if nonce[i] != 0 {
			return
		}
	}
}