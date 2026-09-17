package cipher

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func Decrypt(source, destination, key string) error {
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
		return fmt.Errorf("open encrypted source %q: %w", source, err)
	}
	defer input.Close()

	output, err := os.OpenFile(
		destination,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0600,
	)
	if err != nil {
		return fmt.Errorf("create decrypted file %q: %w", destination, err)
	}

	success := false
	defer func() {
		if !success {
			_ = output.Close()
			_ = os.Remove(destination)
		}
	}()

	if err := decryptStream(input, output, key); err != nil {
		return fmt.Errorf("decrypt %q: %w", source, err)
	}

	if err := output.Sync(); err != nil {
		return fmt.Errorf("sync decrypted file %q: %w", destination, err)
	}

	if err := output.Close(); err != nil {
		return fmt.Errorf("close decrypted file %q: %w", destination, err)
	}

	success = true

	return nil
}

func DecryptReader(reader io.Reader, writer io.Writer, key string) error {
	if reader == nil {
		return errors.New("reader cannot be nil")
	}

	if writer == nil {
		return errors.New("writer cannot be nil")
	}

	if key == "" {
		return errors.New("encryption key cannot be empty")
	}

	return decryptStream(reader, writer, key)
}

func decryptStream(reader io.Reader, writer io.Writer, key string) error {
	block, err := aes.NewCipher(deriveKey(key))
	if err != nil {
		return fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCMWithNonceSize(block, nonceSize)
	if err != nil {
		return fmt.Errorf("create GCM cipher: %w", err)
	}

	nonce := make([]byte, nonceSize)

	if _, err := io.ReadFull(reader, nonce); err != nil {
		return fmt.Errorf("read encryption nonce: %w", err)
	}

	const chunkSize = 1024 * 1024
	ciphertextSize := chunkSize + gcm.Overhead()
	buffer := make([]byte, ciphertextSize)

	for {
		n, err := io.ReadFull(reader, buffer)

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			if errors.Is(err, io.ErrUnexpectedEOF) {
				if n == 0 {
					break
				}

				decrypted, openErr := gcm.Open(nil, nonce, buffer[:n], nil)
				if openErr != nil {
					return fmt.Errorf("authenticate encrypted data: %w", openErr)
				}

				if _, writeErr := writer.Write(decrypted); writeErr != nil {
					return fmt.Errorf("write decrypted data: %w", writeErr)
				}

				break
			}

			return fmt.Errorf("read encrypted data: %w", err)
		}

		decrypted, err := gcm.Open(nil, nonce, buffer[:n], nil)
		if err != nil {
			return fmt.Errorf("authenticate encrypted data: %w", err)
		}

		if _, err := writer.Write(decrypted); err != nil {
			return fmt.Errorf("write decrypted data: %w", err)
		}

		incrementNonce(nonce)
	}

	return nil
}