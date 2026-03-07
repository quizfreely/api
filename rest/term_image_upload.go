package rest

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"image"
	"io"
	"net/http"
	"strings"

	"quizfreely/api/auth"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

const (
	maxSizeBefore       = 10 << 20   /* 10 MB */
	maxPixelsBefore     = 20_000_000 /* 20 MP */
	maxWidthHeightAfter = 1200
	webpQualityAfter    = 80
)

func (rh *RESTHandler) UploadTermImage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if rh.Storage == nil {
		render.Status(r, 503)
		render.JSON(w, r, map[string]any{
			"error": "Storage not enabled/configured",
		})
		return
	}

	termID := chi.URLParam(r, "termID")
	if termID == "" {
		render.Status(r, 400)
		render.JSON(w, r, map[string]any{
			"error": "Missing termID in URL",
		})
		return
	}
	side := chi.URLParam(r, "side")
	if side != "term" && side != "def" {
		render.Status(r, 400)
		render.JSON(w, r, map[string]any{
			"error": "Invalid term/def side in URL",
		})
		return
	}

	authedUser := auth.AuthedUserContext(ctx)
	if authedUser == nil || authedUser.ID == nil || *authedUser.ID == "" {
		render.Status(r, 401)
		render.JSON(w, r, map[string]any{
			"error": "not authenticated while trying to upload image for term",
		})
		return
	}

	/* check if user owns term BEFORE processing image */
	var ownsTerm bool
	err := pgxscan.Get(
		ctx,
		rh.DB,
		&ownsTerm,
		`SELECT EXISTS (
			SELECT 1 FROM terms t
			JOIN studysets s ON s.id = t.studyset_id
			WHERE t.id = $1
			AND s.user_id = $2
		)`,
		termID,
		authedUser.ID,
	)
	if err != nil {
		log.Error().Err(err).Msg("error checking term ownership in UploadTermImage")
		render.Status(r, 400)
		render.JSON(w, r, map[string]any{
			"error": "error checking term ownership",
		})
		return
	}
	if !ownsTerm {
		render.Status(r, 403)
		render.JSON(w, r, map[string]any{
			"error": "term not owned by user, can't upload image for term",
		})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxSizeBefore)

	raw, err := io.ReadAll(r.Body)
	if err != nil {
		render.Status(r, 400)
		render.JSON(w, r, map[string]any{
			"error": "failed to read request body",
		})
		return
	}
	if len(raw) == 0 {
		render.Status(r, 400)
		render.JSON(w, r, map[string]any{
			"error": "empty file",
		})
		return
	}

	/* detect actual MIME type, ignoring user-specified Content-Type */
	mime := http.DetectContentType(raw[:512])
	if !isMIMEAllowed(mime) {
		render.Status(r, 400)
		render.JSON(w, r, map[string]any{
			"error": "unsupported image type",
		})
		return
	}

	cfg, _, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil {
		render.Status(r, 400)
		render.JSON(w, r, map[string]any{
			"error": "invalid image",
		})
		return
	}

	if cfg.Width <= 0 || cfg.Height <= 0 {
		render.Status(r, 400)
		render.JSON(w, r, map[string]any{
			"error": "invalid image dimensions",
		})
		return
	}

	if cfg.Width*cfg.Height > maxPixelsBefore {
		render.Status(r, 400)
		render.JSON(w, r, map[string]any{
			"error": "image too large",
		})
		return
	}

	img, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		render.Status(r, 400)
		render.JSON(w, r, map[string]any{
			"error": "failed to decode image",
		})
		return
	}

	img = imaging.Fit(img, maxWidthHeightAfter, maxWidthHeightAfter, imaging.Lanczos)
	var buf bytes.Buffer
	err = webp.Encode(&buf, img, &webp.Options{
		Lossless: false,
		Quality:  webpQualityAfter,
	})
	if err != nil {
		render.Status(r, 400)
		render.JSON(w, r, map[string]any{
			"error": "failed to encode webp",
		})
		return
	}

	hash := sha256.Sum256(buf.Bytes())
	hashStr := hex.EncodeToString(hash[:])[:32]
	objectKey := "images/" + hashStr[:2] + "/" + hashStr[2:4] + "/" + hashStr + ".webp"

	_, err = rh.Storage.PutObject(ctx, &s3.PutObjectInput{
		Bucket:       rh.UsercontentBucket,
		Key:          aws.String(objectKey),
		Body:         bytes.NewReader(buf.Bytes()),
		ContentType:  aws.String("image/webp"),
		CacheControl: aws.String("public, max-age=31536000, immutable"),
	})

	if err != nil {
		log.Error().Err(err).Msg("failed to upload to s3")
		render.Status(r, 500)
		render.JSON(w, r, map[string]any{
			"error": "failed to upload image to storage",
		})
		return
	}

	_, err = rh.DB.Exec(
		ctx,
		"INSERT INTO images (object_key) VALUES ($1) ON CONFLICT DO NOTHING",
		objectKey,
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to insert image object key in DB")
		render.Status(r, 500)
		render.JSON(w, r, map[string]any{
			"error": "failed to insert image object key in DB",
		})
		return
	}

	sql := "UPDATE terms SET term_image_key = $1 WHERE id = $2"
	if side == "def" {
		sql = "UPDATE terms SET def_image_key = $1 WHERE id = $2"
	}
	_, err = rh.DB.Exec(
		ctx,
		sql,
		objectKey,
		termID,
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to update term/def image key in DB")
		render.Status(r, 500)
		render.JSON(w, r, map[string]any{
			"error": "failed to update term/def image key in DB",
		})
		return
	}

	render.JSON(w, r, map[string]interface{}{
		"error": false,
		"data": map[string]interface{}{
			"imageUrl": *rh.UsercontentBaseURL + objectKey,
		},
	})
}

func isMIMEAllowed(m string) bool {
	switch {
	case strings.HasPrefix(m, "image/jpeg"):
		return true
	case strings.HasPrefix(m, "image/png"):
		return true
	case strings.HasPrefix(m, "image/webp"):
		return true
	default:
		return false
	}
}

func (rh *RESTHandler) RemoveTermImage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	termID := chi.URLParam(r, "termID")
	if termID == "" {
		render.Status(r, 400)
		render.JSON(w, r, map[string]any{
			"error": "Missing termID in URL",
		})
		return
	}
	side := chi.URLParam(r, "side")
	if side != "term" && side != "def" {
		render.Status(r, 400)
		render.JSON(w, r, map[string]any{
			"error": "Invalid term/def side in URL",
		})
		return
	}

	authedUser := auth.AuthedUserContext(ctx)
	if authedUser == nil || authedUser.ID == nil || *authedUser.ID == "" {
		render.Status(r, 401)
		render.JSON(w, r, map[string]any{
			"error": "not authenticated while trying to remove image from term",
		})
		return
	}

	sql := `UPDATE terms t SET term_image_key = null
		WHERE id = $1 AND EXISTS (
			SELECT 1 FROM studysets s WHERE s.id = t.studyset_id AND s.user_id = $2
		)`
	if side == "def" {
		sql = `UPDATE terms t SET def_image_key = null
			WHERE id = $1 AND EXISTS (
				SELECT 1 FROM studysets s WHERE s.id = t.studyset_id AND s.user_id = $2
			)`
	}
	_, err = rh.DB.Exec(
		ctx,
		sql,
		termID,
		authedUser.ID,
	)
	if err != nil {
		log.Error().Err(err).Msg("RemoveTermImage: DB error updating term's image key")
		render.Status(r, 500)
		render.JSON(w, r, map[string]any{
			"error": "DB error updating term's image key",
		})
		return
	}

	render.JSON(w, r, map[string]interface{}{
		"error": false,
	})
}
