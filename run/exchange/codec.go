package exchange

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"acrux/internal/codec"

	tea "github.com/charmbracelet/bubbletea"
)

type codecOperation int

const (
	codecCompress codecOperation = iota
	codecDecompress
)

type codecModel struct {
	operation codecOperation
	source    string
	destination string
	message   string
	quitting  bool
}

func CodecMenu() (string, error) {
	model := codecModel{}

	result, err := tea.NewProgram(model).Run()
	if err != nil {
		return "", fmt.Errorf("run codec menu: %w", err)
	}

	m, ok := result.(codecModel)
	if !ok {
		return "", errors.New("invalid codec menu result")
	}

	if m.quitting && m.source == "" {
		return "", nil
	}

	return m.destination, nil
}

func (m codecModel) Init() tea.Cmd {
	return nil
}

func (m codecModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m codecModel) View() string {
	if m.quitting {
		return ""
	}

	return "Codec\n\nCompression and decompression operations are handled by the copy workflow."
}

func CompressFile(source, destination string) error {
	if strings.TrimSpace(source) == "" {
		return errors.New("source cannot be empty")
	}

	if strings.TrimSpace(destination) == "" {
		return errors.New("destination cannot be empty")
	}

	if filepath.Clean(source) == filepath.Clean(destination) {
		return errors.New("source and destination are the same file")
	}

	if err := codec.Compress(source, destination); err != nil {
		return fmt.Errorf("compress file: %w", err)
	}

	return nil
}

func DecompressFile(source, destination string) error {
	if strings.TrimSpace(source) == "" {
		return errors.New("source cannot be empty")
	}

	if strings.TrimSpace(destination) == "" {
		return errors.New("destination cannot be empty")
	}

	if filepath.Clean(source) == filepath.Clean(destination) {
		return errors.New("source and destination are the same file")
	}

	if err := codec.Decompress(source, destination); err != nil {
		return fmt.Errorf("decompress file: %w", err)
	}

	return nil
}

func CompressReader(reader io.Reader, writer io.Writer) error {
	if reader == nil {
		return errors.New("reader cannot be nil")
	}

	if writer == nil {
		return errors.New("writer cannot be nil")
	}

	if err := codec.CompressReader(reader, writer); err != nil {
		return fmt.Errorf("compress stream: %w", err)
	}

	return nil
}

func DecompressReader(reader io.Reader, writer io.Writer) error {
	if reader == nil {
		return errors.New("reader cannot be nil")
	}

	if writer == nil {
		return errors.New("writer cannot be nil")
	}

	if err := codec.DecompressReader(reader, writer); err != nil {
		return fmt.Errorf("decompress stream: %w", err)
	}

	return nil
}

func CompressToTemporary(source string) (string, error) {
	if strings.TrimSpace(source) == "" {
		return "", errors.New("source cannot be empty")
	}

	info, err := os.Stat(source)
	if err != nil {
		return "", fmt.Errorf("stat source %q: %w", source, err)
	}

	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("source %q is not a regular file", source)
	}

	tempDir, err := os.MkdirTemp("", "acrux-codec-*")
	if err != nil {
		return "", fmt.Errorf("create temporary directory: %w", err)
	}

	destination := filepath.Join(
		tempDir,
		filepath.Base(source)+".gz",
	)

	if err := codec.Compress(source, destination); err != nil {
		_ = os.RemoveAll(tempDir)
		return "", fmt.Errorf("compress temporary file: %w", err)
	}

	return destination, nil
}

func DecompressToTemporary(source string) (string, error) {
	if strings.TrimSpace(source) == "" {
		return "", errors.New("source cannot be empty")
	}

	info, err := os.Stat(source)
	if err != nil {
		return "", fmt.Errorf("stat source %q: %w", source, err)
	}

	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("source %q is not a regular file", source)
	}

	tempDir, err := os.MkdirTemp("", "acrux-codec-*")
	if err != nil {
		return "", fmt.Errorf("create temporary directory: %w", err)
	}

	baseName := filepath.Base(source)
	baseName = strings.TrimSuffix(baseName, ".gz")

	if baseName == "" {
		baseName = "decompressed"
	}

	destination := filepath.Join(tempDir, baseName)

	if err := codec.Decompress(source, destination); err != nil {
		_ = os.RemoveAll(tempDir)
		return "", fmt.Errorf("decompress temporary file: %w", err)
	}

	return destination, nil
}