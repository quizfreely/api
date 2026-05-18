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
		reader, err = rh.crawlbaseReq(reqBody.URL, ctx)
	} else if err != nil && rh.UseCrawlbase && rh.UseZyte {
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
	ctx, cancel := context.WithTimeout(reqCtx, 60*time.Second)
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
	ctx, cancel := context.WithTimeout(reqCtx, 60*time.Second)
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
		return nil, errors.New("no __NEXT_DATA__ :(")
	}

	var nextData struct {
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
	}
	if err := json.Unmarshal([]byte(rawNextData), &nextData); err != nil {
		return nil, err
	}

	var termDefPairs [][]string
	for _, item := range nextData.StudiableData.StudiableItems {
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

