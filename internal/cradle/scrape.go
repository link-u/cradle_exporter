package cradle

import (
	"io"
	"net/http"

	"go.uber.org/zap"
)

func scrapeEndpoint(w io.Writer, configFilePath string, endpoint string) error {
	log := zap.L()
	resp, err := http.Get(endpoint)
	if err != nil {
		return err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Warn("Failed to close response body", zap.String("config-file-path", configFilePath), zap.String("endpoint", endpoint), zap.Error(err))
		}
	}()
	written, err := io.Copy(w, resp.Body)
	if err != nil {
		log.Warn("Failed to read response body", zap.String("config-file-path", configFilePath), zap.String("endpoint", endpoint), zap.Error(err))
		return err
	}
	if written != resp.ContentLength && resp.ContentLength >= 0 {
		log.Warn("Body length does not match to content-length header",
			zap.String("config-file-path", configFilePath),
			zap.String("endpoint", endpoint),
			zap.Int64("written", written),
			zap.Int64("content-length", resp.ContentLength))
	}
	return err
}
