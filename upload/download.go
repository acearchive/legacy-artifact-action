package upload

import (
	"context"
	"fmt"
	"github.com/ipfs/go-cid"
	"io"
	"net/http"
	"time"
)

type HttpStatusError struct {
	Method     string
	Url        string
	Status     string
	StatusCode int
}

func (e HttpStatusError) Error() string {
	return fmt.Sprintf("%s \"%s\" returned %s", e.Method, e.Url, e.Status)
}

const DefaultTimeout time.Duration = 1000 * 1000 * 1000 * 15

var defaultClient = http.Client{Timeout: DefaultTimeout}

func responseIsOk(status int) bool {
	return status >= 200 && status < 300
}

func gatewayUrl(id cid.Cid) string {
	return fmt.Sprintf("https://dweb.link/ipfs/%s", id.String())
}

func DownloadContent(ctx context.Context, contentId cid.Cid) (io.ReadCloser, error) {
	cidUrl := gatewayUrl(contentId)

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, cidUrl, http.NoBody)
	if err != nil {
		return nil, err
	}

	response, err := defaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	if !responseIsOk(response.StatusCode) {
		return nil, &HttpStatusError{
			Method:     http.MethodGet,
			Url:        cidUrl,
			StatusCode: response.StatusCode,
			Status:     response.Status,
		}
	}

	return response.Body, nil
}
