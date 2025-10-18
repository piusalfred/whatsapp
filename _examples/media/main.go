package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/piusalfred/whatsapp/config"
	examples "github.com/piusalfred/whatsapp/examples"
	"github.com/piusalfred/whatsapp/media"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

type MediaService struct {
	logger *slog.Logger
	client *media.BaseClient
	reader config.ReaderFunc
}

func NewMediaService(configFilePath string) *MediaService {
	lh := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}).WithGroup("media")

	ms := &MediaService{
		logger: slog.New(lh),
		reader: examples.LoadConfigFromFile(),
	}

	clientOptions := []whttp.CoreClientOption[any]{
		whttp.WithCoreClientHTTPClient[any](examples.HttpClient()),
	}

	coreClient := whttp.NewSender[any](clientOptions...)
	ms.client = media.NewBaseClient(ms.reader, coreClient)

	return ms
}

func main() {
	ctx := context.Background()
	ms := NewMediaService("../api.env")

	image, err := downloadImageFromNet(ctx)
	if err != nil {
		ms.logger.LogAttrs(ctx, slog.LevelError, "failed to download image", slog.String("error", err.Error()))
		return
	}

	// write the image to a file in the current directory call it awesome-puppies.jpg
	file, err := os.Create("awesome-puppies.jpg")
	if err != nil {
		ms.logger.LogAttrs(ctx, slog.LevelError, "failed to create file", slog.String("error", err.Error()))
		return
	}

	defer file.Close()

	// write the image to the file
	_, err = io.Copy(file, image)
	if err != nil {
		ms.logger.LogAttrs(ctx, slog.LevelError, "failed to write image to file", slog.String("error", err.Error()))
		return
	}

	// upload media
	uploadMediaResponse, err := ms.client.Upload(ctx, &media.UploadRequest{
		MediaType: media.TypeImageJPEG,
		Filepath:  file.Name(),
	})
	if err != nil {
		ms.logger.LogAttrs(ctx, slog.LevelError, "failed to upload image", slog.String("error", err.Error()))
		return
	}

	ms.logger.LogAttrs(ctx, slog.LevelInfo, "image uploaded successfully",
		slog.String("media_id", uploadMediaResponse.ID),
	)

	info, err := ms.client.GetInfo(ctx, &media.BaseRequest{
		MediaID: "1869943773825841",
	})
	if err != nil {
		ms.logger.LogAttrs(ctx, slog.LevelError, "failed to get media info", slog.String("error", err.Error()))
		return
	}

	ms.logger.LogAttrs(ctx, slog.LevelInfo, "media info retrieved successfully",
		slog.Any("info", info),
	)

	decoder := whttp.ResponseDecoderFunc(func(ctx context.Context, response *http.Response) error {
		// read body an dump it to the current directory as hahahahah-madeit.jpg
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		defer response.Body.Close()

		file, err := os.Create("hahahahah-madeit.jpg")
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}

		defer file.Close()

		_, err = file.Write(body)
		if err != nil {
			return fmt.Errorf("failed to write to file: %w", err)
		}

		ms.logger.LogAttrs(ctx, slog.LevelInfo, "file downloaded successfully", slog.String("file", file.Name()))

		return nil
	})

	if err := ms.client.DownloadByMediaID(ctx, &media.BaseRequest{
		MediaID:            info.ID,
		RestrictToOwnMedia: false,
		PhoneNumberID:      "",
	}, decoder, media.WithDownloadRetries(1)); err != nil {
		ms.logger.LogAttrs(ctx, slog.LevelError, "failed to download media", slog.String("error", err.Error()))
	}

	ms.logger.LogAttrs(ctx, slog.LevelInfo, "media downloaded successfully")

	// delete the file
	deleteMediaResponse, err := ms.client.Delete(ctx, &media.BaseRequest{
		MediaID: info.ID,
	})
	if err != nil {

		ms.logger.LogAttrs(ctx, slog.LevelError, "failed to delete media", slog.String("error", err.Error()))
		return
	}

	ms.logger.LogAttrs(ctx, slog.LevelInfo, "media deleted successfully",
		slog.Any("media_id", deleteMediaResponse))
}

func downloadImageFromNet(ctx context.Context) (io.Reader, error) {
	link := "https://images.pexels.com/photos/1108099/pexels-photo-1108099.jpeg?cs=srgb&dl=pexels-chevanon-1108099.jpg&fm=jpg&w=1920&h=1440&_gl=1*kv6c3l*_ga*MjYxNDAxNzkwLjE3NDU1NDY0NzM.*_ga_8JE65Q40S6*MTc0NTU0NjQ3My4xLjEuMTc0NTU0NjQ5MS4wLjAuMA.."

	resp, err := http.Get(link)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download image: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	defer resp.Body.Close()

	return io.NopCloser(strings.NewReader(string(body))), nil
}
