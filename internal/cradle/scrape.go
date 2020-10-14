package cradle

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"go.uber.org/zap"
)

func scrapeEndpoint(ctx context.Context, w io.Writer, configFilePath string, endpoint string) {
	log := zap.L()
	var client http.Client
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		log.Error("Failed to create request", zap.String("config-file-path", configFilePath), zap.String("endpoint", endpoint), zap.Error(err))
		_, _ = io.WriteString(w, "### Scraping Target\n")
		_, _ = io.WriteString(w, "### Err: Failed to create request\n")
		_, _ = io.WriteString(w, "### URL: "+endpoint+"\n")
		_, _ = io.WriteString(w, "### Config: "+configFilePath+"\n")
		_, _ = io.WriteString(w, promCommentOut(err.Error()))
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Failed to execute request", zap.String("config-file-path", configFilePath), zap.String("endpoint", endpoint), zap.Error(err))
		_, _ = io.WriteString(w, "### Scraping Target\n")
		_, _ = io.WriteString(w, "### Err: Failed to execute request\n")
		_, _ = io.WriteString(w, "### URL: "+endpoint+"\n")
		_, _ = io.WriteString(w, "### Config: "+configFilePath+"\n")
		_, _ = io.WriteString(w, promCommentOut(err.Error()))
		return
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Warn("Failed to close response body", zap.String("config-file-path", configFilePath), zap.String("endpoint", endpoint), zap.Error(err))
		}
	}()
	var buf bytes.Buffer
	written, err := io.Copy(&buf, resp.Body)
	if err != nil {
		log.Error("Failed to read response body", zap.String("config-file-path", configFilePath), zap.String("endpoint", endpoint), zap.Error(err))
		_, _ = io.WriteString(w, "### Scraping Target\n")
		_, _ = io.WriteString(w, "### Err: Failed to read response body\n")
		_, _ = io.WriteString(w, "### URL: "+endpoint+"\n")
		_, _ = io.WriteString(w, "### Config: "+configFilePath+"\n")
		_, _ = io.WriteString(w, promCommentOut(err.Error()))
	}
	if written != resp.ContentLength && resp.ContentLength >= 0 {
		log.Warn("Body length does not match to content-length header",
			zap.String("config-file-path", configFilePath),
			zap.String("endpoint", endpoint),
			zap.Int64("written", written),
			zap.Int64("content-length", resp.ContentLength))
	}
	_, _ = io.WriteString(w, "### Scraping Target\n")
	_, _ = io.WriteString(w, "### URL: "+endpoint+"\n")
	_, _ = io.WriteString(w, "### Config: "+configFilePath+"\n")
	_, _ = w.Write(buf.Bytes())
}
