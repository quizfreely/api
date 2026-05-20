package rest

import (
	"net/http"
	"net/url"
	"time"
	"context"
	"errors"
	"encoding/base64"
	"encoding/json"
	"io"
	"bytes"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
	"github.com/PuerkitoBio/goquery"
	"os"
	"path/filepath"
)

func (rh *RESTHandler) WebImport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var reqBody struct {
		URL string `json:"url"`
	}
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		render.Status(r, 400)
		render.JSON(w, r, map[string]any{
			"error": "invalid request body",
		})
		return
	}

	var reader io.Reader
	if rh.UseZyte && (rh.TryZyteBeforeCrawlbase || !rh.UseCrawlbase) {
		reader, err = rh.zyteReq(reqBody.URL, ctx)
	} else if rh.UseCrawlbase {
		reader, err = rh.crawlbaseReq(reqBody.URL, ctx)
	} else {
		log.Error().Err(err).Msg("web import unavailable. use_crawlbase and use_zyte are both disabled/false. check config.toml")
		render.Status(r, 503)
		render.JSON(w, r, map[string]any{
			"error": "web import unavailable because use_crawlbase and use_zyte are both disabled/false",
		})
		return
	}

	if err != nil && rh.UseZyte && rh.TryZyteBeforeCrawlbase && rh.UseCrawlbase {
		log.Error().Err(err).Msg("web import zyte err on first try")
		reader, err = rh.crawlbaseReq(reqBody.URL, ctx)
	} else if err != nil && rh.UseCrawlbase && rh.UseZyte {
		log.Error().Err(err).Msg("web import crawlbase err on first try")
		reader, err = rh.zyteReq(reqBody.URL, ctx)
	}

	if err != nil {
		log.Error().Err(err).Msg("web import err w crawlbase/zyte req")
		render.Status(r, 500)
		render.JSON(w, r, map[string]any{
			"error": "error with crawlbase/zyte request",
		})
		return
	}

	terms, err := parse(reader)
	if err != nil {

		/* for debugging only */
		tmpPath, tmpErr := saveToTempFile(reader)
		if tmpErr == nil {
			log.Error().Err(err).Str("file", tmpPath).Msg("web import parsing err")
		} else {
			log.Error().Err(err).Msg("web import parsing err")
		}

		log.Error().Err(err).Msg("web import parsing err")
		render.Status(r, 500)
		render.JSON(w, r, map[string]any{
			"error": "error parsing terms",
		})
		return
	}

	render.JSON(w, r, map[string]any{
		"terms": terms,
	})
}

func (rh *RESTHandler) crawlbaseReq(targetURL string, reqCtx context.Context) (io.Reader, error) {
	log.Trace().Msg("crawlbase attempted")
	ctx, cancel := context.WithTimeout(reqCtx, 90*time.Second)
	defer cancel()

	params := url.Values{}
	params.Add("token", rh.CrawlbaseAPIKey)
	params.Add("url", targetURL)
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://api.crawlbase.com/?" + params.Encode(),
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := rh.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, resp.Body); err != nil {
		return nil, err
	}

	return buf, nil
}

type zyteReqBody struct {
	URL string `json:"url"`
	HTTPResponseBody bool `json:"httpResponseBody"`
}
type zyteRespBody struct {
	HTTPResponseBody string `json:"httpResponseBody"`
}
func (rh *RESTHandler) zyteReq(targetURL string, reqCtx context.Context) (io.Reader, error) {
	log.Trace().Msg("zyte attempted")
	ctx, cancel := context.WithTimeout(reqCtx, 90*time.Second)
	defer cancel()

	reqBodyJSON, err := json.Marshal(
		zyteReqBody{
			URL: targetURL,
			HTTPResponseBody: true,
		},
	)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.zyte.com/v1/extract",
		bytes.NewBuffer(reqBodyJSON),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(rh.ZyteAPIKey, "")

	resp, err := rh.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var respBody zyteRespBody
	err = json.Unmarshal(body, &respBody)
	if err != nil {
		return nil, err
	}

	decodedBody, err := base64.StdEncoding.DecodeString(respBody.HTTPResponseBody)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(decodedBody), nil
}

func parse(reader io.Reader) ([][]string, error) {
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, err
	}

	rawNextData := doc.Find(`script#__NEXT_DATA__`).Text()
	if rawNextData == "" {
		// TODO: hmm
		return nil, errors.New("no __NEXT_DATA__ :(")
	}

	var nextData struct {
		Props struct {
			PageProps struct {
				DehydratedReduxStateKey json.RawMessage `json:"dehydratedReduxStateKey"`
			} `json:"pageProps"`
		} `json:"props"`
	}
	var dhrsk struct {
		StudyModesCommon struct {
			StudiableData struct {
				StudiableItems []struct {
					IsDeleted bool `json:"isDeleted"`
					CardSides []struct {
						SideID int `json:"sideId"`
						Media []struct {
							PlainText string `json:"plainText"`
							RichText *string `json:"richText"`
						} `json:"media"`
					} `json:"cardSides"`
				} `json:"studiableItems"`
			} `json:"studiableData"`
		}`json:"StudyModesCommon"`
	}
	if err := json.Unmarshal([]byte(rawNextData), &nextData); err != nil {
		return nil, err
	}
	trimmedRawDHRSK := bytes.TrimSpace(nextData.Props.PageProps.DehydratedReduxStateKey)
	log.Trace().Str("first char", string(trimmedRawDHRSK[0])).Msg("hmm")
	if len(trimmedRawDHRSK) == 0 {
		// TODO: hmm
		return nil, errors.New("empty dehydratedReduxStateKey in __NEXT_DATA__")
	}
	if trimmedRawDHRSK[0] == '"' {
		var dhrskString string
		if err := json.Unmarshal(nextData.Props.PageProps.DehydratedReduxStateKey, &dhrskString); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(dhrskString), &dhrsk); err != nil {
			return nil, err
		}
	} else if trimmedRawDHRSK[0] == '{' {
		if err := json.Unmarshal(nextData.Props.PageProps.DehydratedReduxStateKey, &dhrsk); err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("unexpected JSON token in dehydratedReduxStateKey in __NEXT_DATA__")
	}

	var termDefPairs [][]string
	for _, item := range dhrsk.StudyModesCommon.StudiableData.StudiableItems {
		sides := len(item.CardSides)
		if sides < 1 {
			continue
		}

		term := ""
		def := ""
		if sides >= 1 {
			term = item.CardSides[0].Media[0].PlainText
		}
		if sides >= 2 {
			def = item.CardSides[1].Media[0].PlainText
		}
		termDefPairs = append(termDefPairs, []string{term, def})
	}

	return termDefPairs, nil
}

// saveToTempFile dumps the raw bytes into a temporary file on the server (for debugging only)
// returns filename (to log)
func saveToTempFile(r io.Reader) (string, error) {
	// Creates a file like /tmp/web-import-failed-123456789.html
	tmpFile, err := os.CreateTemp("", "web-import-failed-*.html")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, r); err != nil {
		return "", err
	}

	// Returns the absolute path so we can log it
	return filepath.Abs(tmpFile.Name())
}
