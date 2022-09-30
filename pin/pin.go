package pin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/acearchive/artifact-action/cfg"
	"github.com/acearchive/artifact-action/logger"
	"github.com/acearchive/artifact-action/parse"
	"github.com/ipfs/go-cid"
)

// This package deliberately does not use the reference implementation of the
// IPFS pinning services client because it does not encode metadata properly in
// the request parameters, and this tool relies on the metadata functionality
// to work around the broken paging implementation. See the issue below for
// details:
//
// https://github.com/ipfs/go-pinning-service-http-client/issues/8

var getPinClient func() *http.Client = func() func() *http.Client {
	client := &http.Client{}

	return func() *http.Client {
		return client
	}
}()

// This is deliberately larger than necessary in order to mitigate problems
// associated with paging. See `listExistingPins`.
const pageLimit = "200"

// This can not be > 10 per the API spec.
const batchLimit = 10

const (
	headerAuthorization = "Authorization"
	headerContentType   = "Content-Type"
)

type metaMap map[string]string

func (m metaMap) Encode() (string, error) {
	bytes, err := json.Marshal(m)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

const (
	baseMetaKey = "lgbt.acearchive.artifact-action"
	kindMetaKey = "lgbt.acearchive.artifact-action.kind"
)

func rootMeta() metaMap {
	return metaMap{baseMetaKey: "", kindMetaKey: "root"}
}

func fileMeta() metaMap {
	return metaMap{baseMetaKey: "", kindMetaKey: "file"}
}

func addAuthHeader(request *http.Request, token string) {
	request.Header.Add(headerAuthorization, fmt.Sprintf("Bearer %s", token))
}

func joinUrlPath(base url.URL, elements ...string) url.URL {
	var newPath strings.Builder

	newPath.WriteString(strings.TrimRight(base.Path, "/"))

	for _, elem := range elements {
		newPath.WriteRune('/')
		newPath.WriteString(elem)
	}

	base.Path = newPath.String()

	return base
}

func setQueryParams(base *url.URL, params map[string]string) {
	query := base.Query()

	for key, value := range params {
		query.Set(key, value)
	}

	base.RawQuery = query.Encode()
}

type apiError struct {
	Method  string
	URL     *url.URL
	Status  string
	Reason  string
	Details *string
}

func (e apiError) Error() string {
	description := e.Reason

	if e.Details != nil {
		description += fmt.Sprintf("\n%s", *e.Details)
	}

	return fmt.Sprintf("Returned %s\n%s %s\n\n%s\n", e.Status, e.Method, e.URL.String(), description)
}

func checkResponseError(response *http.Response, goodStatusCodes ...int) error {
	for _, statusCode := range goodStatusCodes {
		if response.StatusCode == statusCode {
			return nil
		}
	}

	var responseObj errorResponse

	if err := json.NewDecoder(response.Body).Decode(&responseObj); err != nil {
		return apiError{
			Method: response.Request.Method,
			URL:    response.Request.URL,
			Status: response.Status,
			Reason: err.Error(),
		}
	}

	response.Body.Close()

	return apiError{
		Method:  response.Request.Method,
		URL:     response.Request.URL,
		Status:  response.Status,
		Reason:  responseObj.Error.Reason,
		Details: responseObj.Error.Details,
	}
}

type pinResponse struct {
	Cid string `json:"cid"`
}

type pinStatusResponse struct {
	Pin pinResponse `json:"pin"`
}

type listPinsResponse struct {
	Results []pinStatusResponse `json:"results"`
}

type errorReasonResponse struct {
	Reason  string  `json:"reason"`
	Details *string `json:"details"`
}

type errorResponse struct {
	Error errorReasonResponse `json:"error"`
}

// listExistingPins returns the subset of the CIDs that have already been
// pinned to the remote by this tool.
//
// The way that we have to go about this is wonky. As of time of writing,
// paging is fundamentally broken in the pinning services API, both in terms of
// the spec and in terms of common implementations. See this issue for details:
// https://github.com/ipfs/kubo/issues/8762
//
// Instead of using paging, we batch together a few CIDs and ask the service to
// list pins with any of those CIDs. If one of the pins has that CID, then the
// CID is already pinned. However, multiple pins can have the same CID, and
// since we can't use paging, they all have to fit in one page.
//
// To address this problem, we tag pins which were created by this tool with
// metadata, and we tell the service to only return pins with metadata
// indicating they were made by this tool. Since this tool avoids re-pinning
// files that have already been pinned, *theoretically* there should only ever
// be one pin per CID given the filter parameters.
//
// However, there's all sorts of ways this could break, such as race conditions
// if two instances of this tool are pinning to the same service with the same
// credentials at the same time. We can mitigate this somewhat by using a
// larger-than-necessary page size, but ultimately this solution will have to
// be best-effort until paging is un-borked.
func listExistingPins(ctx context.Context, endpoint url.URL, token string, fileCids parse.DeduplicatedCIDList) (parse.ContentSet, error) {
	existingContent := make(parse.ContentSet)
	currentIndex := 0

	pinClient := getPinClient()

	for {
		cidBatch := make([]string, 0, batchLimit)

		for _, fileCid := range fileCids[currentIndex:] {
			currentIndex++
			cidBatch = append(cidBatch, fileCid.String())

			if len(cidBatch) == batchLimit {
				break
			}
		}

		metaParam, err := fileMeta().Encode()
		if err != nil {
			return nil, err
		}

		requestUrl := joinUrlPath(endpoint, "pins")
		setQueryParams(&requestUrl, map[string]string{
			"cid":    strings.Join(cidBatch, ","),
			"status": "queued,pinning,pinned",
			"limit":  pageLimit,
			"meta":   metaParam,
		})

		request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestUrl.String(), http.NoBody)
		if err != nil {
			return nil, err
		}

		addAuthHeader(request, token)

		response, err := pinClient.Do(request)
		if err != nil {
			return nil, err
		}

		if response.StatusCode == http.StatusNotFound {
			return existingContent, nil
		}

		if err := checkResponseError(response, http.StatusOK); err != nil {
			return nil, err
		}

		var responseObj listPinsResponse

		if err := json.NewDecoder(response.Body).Decode(&responseObj); err != nil {
			return nil, err
		}

		response.Body.Close()

		for _, pinStatus := range responseObj.Results {
			resultCid, err := cid.Parse(pinStatus.Pin.Cid)
			if err != nil {
				return nil, err
			}

			existingContent[parse.ContentKeyFromCid(resultCid)] = struct{}{}
		}

		if len(cidBatch) < batchLimit {
			break
		}
	}

	logger.Printf("Found %d pins\n", len(existingContent))

	return existingContent, nil
}

type addPinRequest struct {
	Cid  string  `json:"cid"`
	Meta metaMap `json:"meta"`
}

func (r addPinRequest) Encode() ([]byte, error) {
	return json.Marshal(r)
}

func addPin(ctx context.Context, endpoint url.URL, token string, pinCid cid.Cid, meta metaMap) error {
	if cfg.DryRun() {
		return nil
	}

	requestUrl := joinUrlPath(endpoint, "pins")
	body, err := addPinRequest{Cid: pinCid.String(), Meta: meta}.Encode()
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, requestUrl.String(), bytes.NewReader(body))
	if err != nil {
		return err
	}

	addAuthHeader(request, token)
	request.Header.Add(headerContentType, "application/json")

	response, err := getPinClient().Do(request)
	if err != nil {
		return err
	}

	if err := checkResponseError(response, http.StatusAccepted); err != nil {
		return err
	}

	return response.Body.Close()
}

func pinDeduplicated(ctx context.Context, endpoint url.URL, token string, fileCids parse.DeduplicatedCIDList, rootCid cid.Cid, alreadyPinned parse.ContentSet) error {

	filesToUpload := make([]cid.Cid, 0, len(fileCids))

	for _, id := range fileCids {
		if _, alreadyUploaded := alreadyPinned[parse.ContentKeyFromCid(id)]; !alreadyUploaded {
			filesToUpload = append(filesToUpload, id)
		}
	}

	logger.Printf("Skipping %d files that are already pinned\n", len(fileCids)-len(filesToUpload))
	logger.Printf("Pinning %d files\n", len(filesToUpload))

	for currentIndex, currentCid := range filesToUpload {
		logger.Printf("Pinning (%d/%d): %s\n", currentIndex+1, len(filesToUpload), currentCid.String())

		if err := addPin(ctx, endpoint, token, currentCid, fileMeta()); err != nil {
			return err
		}
	}

	logger.Printf("\nPinning the root directory: /ipfs/%s\n", rootCid.String())

	if err := addPin(ctx, endpoint, token, rootCid, rootMeta()); err != nil {
		return err
	}

	return nil
}

func Pin(ctx context.Context, endpoint, token string, fileCids parse.DeduplicatedCIDList, rootCid cid.Cid) error {
	endpointUrl, err := url.Parse(endpoint)
	if err != nil {
		return err
	}

	if len(endpointUrl.Scheme) == 0 {
		endpointUrl.Scheme = "https"
	}

	existingContent, err := listExistingPins(ctx, *endpointUrl, token, fileCids)
	if err != nil {
		return err
	}

	return pinDeduplicated(ctx, *endpointUrl, token, fileCids, rootCid, existingContent)
}
