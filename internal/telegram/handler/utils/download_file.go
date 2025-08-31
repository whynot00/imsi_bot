package utils

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/go-telegram/bot"
)

func DownloadFile(ctx context.Context, b *bot.Bot, fileID, destPath string) error {

	file, err := b.GetFile(ctx, &bot.GetFileParams{FileID: fileID})
	if err != nil {
		return fmt.Errorf("GetFile: %w", err)
	}

	url := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", b.Token(), file.FilePath)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("http.Get: %w", err)
	}
	defer resp.Body.Close()

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("os.Create: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}

	return nil
}
