package images

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"image"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	intimages "github.com/arran4/goa4web/internal/images"
	"github.com/arran4/goa4web/internal/middleware/apiauth"
	nav "github.com/arran4/goa4web/internal/navigation"
	"github.com/arran4/goa4web/internal/router"
	"github.com/arran4/goa4web/internal/sign"
	"github.com/arran4/goa4web/internal/upload"
)

func verifyMiddleware(prefix string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := mux.Vars(r)["id"]
			if !intimages.ValidID(id) {
				w.WriteHeader(http.StatusForbidden)
				handlers.RenderErrorPage(w, r, fmt.Errorf("invalid id"))
				return
			}

			query := r.URL.Query()
			sig := query.Get("sig")
			tsStr := query.Get("ts")
			nonce := query.Get("nonce")

			cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

			var err error
			var data string

			if tsStr != "" || nonce != "" {
				var opts []sign.SignOption
				if tsStr != "" {
					ts, _ := strconv.ParseInt(tsStr, 10, 64)
					opts = append(opts, sign.WithExpiry(time.Unix(ts, 0)))
				}
				if nonce != "" {
					opts = append(opts, sign.WithNonce(nonce))
				}
				// Verify "image:ID" or "cache:ID"
				data = prefix + id
				err = sign.Verify(data, sig, cd.ImageSignKey, opts...)
			} else {
				query.Del("ts")
				query.Del("sig")
				data = id
				if encoded := query.Encode(); encoded != "" {
					data = data + "?" + encoded
				}
				if prefix != "" {
					data = prefix + data
				}
				err = sign.Verify(data, sig, cd.ImageSignKey, sign.WithOutNonce())
			}

			if cd.ImageSignKey == "" || err != nil {
				w.WriteHeader(http.StatusForbidden)
				handlers.RenderErrorPage(w, r, fmt.Errorf("forbidden"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RegisterRoutes attaches the image endpoints to r.
func RegisterRoutes(r *mux.Router, cfg *config.RuntimeConfig) []nav.RouterOptions {
	ir := r.PathPrefix("/images").Subrouter()
	ir.Use(handlers.IndexMiddleware(CustomIndex))
	ir.HandleFunc("/upload/image", handlers.TaskHandler(uploadImageTask)).
		Methods(http.MethodPost).
		MatcherFunc(handlers.RequiresAnAccount()).
		MatcherFunc(uploadImageTask.Matcher())
	ir.HandleFunc("/pasteimg.js", handlers.PasteImageJS(cfg)).Methods(http.MethodGet)
	ir.Handle("/image/{id}", verifyMiddleware("image:")(http.HandlerFunc(serveImage))).
		Methods(http.MethodGet)

	api := ir.PathPrefix("/api").Subrouter()
	api.Use(apiauth.APIKeyAuthMiddleware)

	apiRead := api.NewRoute().Subrouter()
	apiRead.Use(apiauth.RequireScope("images:read"))
	apiRead.HandleFunc("", APIListGallery).Methods(http.MethodGet)

	apiWrite := api.NewRoute().Subrouter()
	apiWrite.Use(apiauth.RequireScope("images:write"))
	apiWrite.HandleFunc("/upload", APIUploadImage).Methods(http.MethodPost)

	ir.Handle("/cache/{id}", verifyMiddleware("cache:")(http.HandlerFunc(serveCache))).
		Methods(http.MethodGet)
	return nil
}

func serveImage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if !intimages.ValidID(id) {
		w.WriteHeader(http.StatusForbidden)
		handlers.RenderErrorPage(w, r, fmt.Errorf("invalid id"))
		return
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cfg := cd.Config
	sub1, sub2 := id[:2], id[2:4]

	full := filepath.Join(cfg.ImageUploadDir, sub1, sub2, id)
	info, err := os.Stat(full)
	if os.IsNotExist(err) {
		var opts []templates.Option
		if cfg != nil && cfg.TemplatesDir != "" {
			opts = append(opts, templates.WithDir(cfg.TemplatesDir))
		}
		w.Header().Set("Content-Type", "image/svg+xml")
		http.ServeContent(w, r, "missing_image.svg", time.Time{}, bytes.NewReader(templates.GetMissingImageData(opts...)))
		return
	}

	// Get preferred dimension
	safeDim := ""
	if cd.Config != nil {
		if dims := cd.Config.SafeImageDimensions(); len(dims) > 0 {
			safeDim = dims[0]
		}
	}
	if pref, err := cd.Preference(); err == nil && pref != nil && pref.ImageSafeDimension.Valid {
		safeDim = pref.ImageSafeDimension.String
	}

	maxW, maxH := 0, 0
	if safeDim != "" {
		maxW, maxH, _ = intimages.ParseDimension(safeDim)
	}

	key := path.Join(sub1, sub2, id)

	// Check size before reading into memory
	if cfg != nil && info.Size() > int64(cfg.ImageMaxResizeBytes) && maxW > 0 && maxH > 0 {
		if up, cacheProvider := upload.ProviderFromConfig(cfg), upload.CacheProviderFromConfig(cfg); up != nil && cacheProvider != nil {
			safeKey := key + "_safe_" + safeDim
			safeBytes, safeErr := cacheProvider.Read(r.Context(), safeKey)
			if safeErr == nil {
				http.ServeContent(w, r, id, info.ModTime(), bytes.NewReader(safeBytes))
				return
			}

			origBytes, err := up.Read(r.Context(), key)
			if err == nil {
				if config, _, err := image.DecodeConfig(bytes.NewReader(origBytes)); err == nil {
					if config.Width <= maxW && config.Height <= maxH {
						if err := cacheProvider.Write(r.Context(), safeKey, origBytes); err == nil {
							recordUploadedImageDerivative(r.Context(), cd, path.Base(safeKey), id, origBytes, config.Height, config.Width)
						}
						http.ServeContent(w, r, id, info.ModTime(), bytes.NewReader(origBytes))
						return
					}
				}
				img, _, err := image.Decode(bytes.NewReader(origBytes))
				if err == nil {
					ext := filepath.Ext(id)
					generator := "bild"
					if cfg != nil && cfg.ImageThumbnailGenerator != "" {
						generator = cfg.ImageThumbnailGenerator
					}
					resizedBytes, err := intimages.GenerateSafeSize(img, ext, generator, maxW, maxH)
					if err == nil {
						if err := cacheProvider.Write(r.Context(), safeKey, resizedBytes); err == nil {
							height, width, dimensionsErr := intimages.DimensionsWithinBounds(img, maxH, maxW)
							if dimensionsErr == nil {
								recordUploadedImageDerivative(r.Context(), cd, path.Base(safeKey), id, resizedBytes, height, width)
							}
						}
						http.ServeContent(w, r, id, info.ModTime(), bytes.NewReader(resizedBytes))
						return
					}
				}
			}
		}
	}

	http.ServeFile(w, r, full)
}

func recordUploadedImageDerivative(ctx context.Context, cd *common.CoreData, cacheID, imageID string, body []byte, height, width int) {
	source, err := cd.UploadedImageByImageID(imageID)
	if err != nil {
		log.Printf("find cached image source: %v", err)
		return
	}
	if err := cd.RecordUploadedImageDerivative(ctx, cacheID, source, body, height, width); err != nil {
		log.Printf("record uploaded image cache entry: %v", err)
	}
}

func getSafeDimensions(cd *common.CoreData) (string, int, int) {
	safeDim := ""
	if cd.Config != nil {
		if dims := cd.Config.SafeImageDimensions(); len(dims) > 0 {
			safeDim = dims[0]
		}
	}
	if pref, err := cd.Preference(); err == nil && pref != nil && pref.ImageSafeDimension.Valid {
		safeDim = pref.ImageSafeDimension.String
	}

	maxW, maxH := 0, 0
	if safeDim != "" {
		maxW, maxH, _ = intimages.ParseDimension(safeDim)
	}
	return safeDim, maxW, maxH
}

func enforceSafeImageSize(w http.ResponseWriter, r *http.Request, cd *common.CoreData, p upload.Provider, id, key, safeDim string, maxW, maxH int) ([]byte, bool, error) {
	safeKey := key + "_safe_" + safeDim
	safeData, safeErr := p.Read(r.Context(), safeKey)
	if safeErr == nil {
		http.ServeContent(w, r, id, time.Now(), bytes.NewReader(safeData))
		return nil, true, nil
	}

	data, err := p.Read(r.Context(), key)
	if err == nil {
		cfg := cd.Config
		if cfg != nil && len(data) > cfg.ImageMaxResizeBytes {
			if config, _, err := image.DecodeConfig(bytes.NewReader(data)); err == nil {
				if config.Width <= maxW && config.Height <= maxH {
					if err := p.Write(r.Context(), safeKey, data); err == nil {
						if err := cd.RecordDerivedImageCacheEntry(r.Context(), path.Base(safeKey), data); err != nil {
							log.Printf("record safe image cache entry: %v", err)
						}
					}
					http.ServeContent(w, r, id, time.Now(), bytes.NewReader(data))
					return nil, true, nil
				}
			}
			img, _, err := image.Decode(bytes.NewReader(data))
			if err == nil {
				ext := filepath.Ext(id)
				generator := "bild"
				if cfg != nil && cfg.ImageThumbnailGenerator != "" {
					generator = cfg.ImageThumbnailGenerator
				}
				safeData, err := intimages.GenerateSafeSize(img, ext, generator, maxW, maxH)
				if err == nil {
					if err := p.Write(r.Context(), safeKey, safeData); err == nil {
						if err := cd.RecordDerivedImageCacheEntry(r.Context(), path.Base(safeKey), safeData); err != nil {
							log.Printf("record safe image cache entry: %v", err)
						}
					}
					data = safeData
				}
			}
		}
	}
	return data, false, err
}

func regenerateMissingThumbnail(ctx context.Context, cd *common.CoreData, p upload.Provider, id, key, originalID string, thumbnailSize config.ThumbnailSize) []byte {
	origKey := path.Join(originalID[:2], originalID[2:4], originalID)

	if origBytes, origErr := p.Read(ctx, origKey); origErr == nil {
		if data := generateAndSaveThumbnail(ctx, cd, p, id, key, originalID, origBytes, thumbnailSize, true); data != nil {
			return data
		}
	}

	if up := upload.ProviderFromConfig(cd.Config); up != nil {
		if origBytes, origErr := up.Read(ctx, origKey); origErr == nil {
			if data := generateAndSaveThumbnail(ctx, cd, p, id, key, originalID, origBytes, thumbnailSize, false); data != nil {
				return data
			}
		}
	}

	return nil
}

func generateAndSaveThumbnail(ctx context.Context, cd *common.CoreData, p upload.Provider, id, key, originalID string, origBytes []byte, thumbnailSize config.ThumbnailSize, fromCache bool) []byte {
	img, _, err := image.Decode(bytes.NewReader(origBytes))
	if err != nil {
		return nil
	}

	ext := filepath.Ext(originalID)
	generator := "bild"
	if cd.Config != nil && cd.Config.ImageThumbnailGenerator != "" {
		generator = cd.Config.ImageThumbnailGenerator
	}

	thumbBytes, err := intimages.GenerateThumbnailWithinBounds(img, ext, generator, thumbnailSize.Height, thumbnailSize.Width)
	if err != nil {
		return nil
	}

	if err := p.Write(ctx, key, thumbBytes); err == nil {
		height, width, dimErr := intimages.DimensionsWithinBounds(img, thumbnailSize.Height, thumbnailSize.Width)
		if dimErr != nil {
			log.Printf("thumbnail dimensions: %v", dimErr)
		}

		if fromCache {
			parent, parentErr := cd.ImageCacheEntry(ctx, originalID)
			if parentErr != nil {
				log.Printf("find cached thumbnail source: %v", parentErr)
			} else if recordErr := cd.RecordCachedImageThumbnail(ctx, id, parent, thumbBytes, height, width); recordErr != nil {
				log.Printf("record cached image thumbnail entry: %v", recordErr)
			}
		} else if !fromCache {
			source, sourceErr := cd.UploadedImageByImageID(originalID)
			if sourceErr != nil {
				log.Printf("find thumbnail source image: %v", sourceErr)
			} else if recordErr := cd.RecordUploadedImageThumbnail(ctx, id, source, thumbBytes, height, width); recordErr != nil {
				log.Printf("record uploaded image thumbnail cache entry: %v", recordErr)
			}
		}
		return thumbBytes
	}

	return nil
}

func serveCache(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if !intimages.ValidID(id) {
		w.WriteHeader(http.StatusForbidden)
		handlers.RenderErrorPage(w, r, fmt.Errorf("invalid id"))
		return
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cfg := cd.Config
	originalID, thumbnailSize, isThumbnail := thumbnailRequest(id, cfg)
	ok, err := cd.PrepareImageCacheEntryForServe(r.Context(), id)
	if err != nil || !ok {
		entry, entryErr := cd.ImageCacheEntry(r.Context(), id)
		if entryErr != nil && entryErr != sql.ErrNoRows {
			entry = nil
		}
		serveMissingCacheImage(w, r, cfg, id, entry)
		return
	}
	sub1, sub2 := id[:2], id[2:4]
	key := path.Join(sub1, sub2, id)
	safeDim, maxW, maxH := getSafeDimensions(cd)

	if p := upload.CacheProviderFromConfig(cfg); p != nil {
		var data []byte
		var err error

		if maxW > 0 && maxH > 0 && !isThumbnail {
			var served bool
			data, served, err = enforceSafeImageSize(w, r, cd, p, id, key, safeDim, maxW, maxH)
			if served {
				return
			}
		} else {
			data, err = p.Read(r.Context(), key)
		}

		if err != nil && isThumbnail {
			data = regenerateMissingThumbnail(r.Context(), cd, p, id, key, originalID, thumbnailSize)
		}

		if data == nil {
			http.NotFound(w, r)
			return
		}
		http.ServeContent(w, r, id, time.Now(), bytes.NewReader(data))
		return
	}

	full := filepath.Join(cfg.ImageCacheDir, sub1, sub2, id)
	http.ServeFile(w, r, full)
}

// thumbnailRequest identifies a permitted thumbnail request and its source image.
func thumbnailRequest(id string, cfg *config.RuntimeConfig) (string, config.ThumbnailSize, bool) {
	ext := filepath.Ext(id)
	if ext == "" {
		return "", config.ThumbnailSize{}, false
	}
	base := strings.TrimSuffix(id, ext)
	sizes := cfg.ThumbnailSizes()
	if before, ok := strings.CutSuffix(base, "_thumb"); ok {
		originalID := before + ext
		return originalID, sizes[0], intimages.ValidID(originalID)
	}
	index := strings.LastIndex(base, "_thumb_")
	if index < 0 {
		return "", config.ThumbnailSize{}, false
	}
	dimensions := strings.Split(base[index+len("_thumb_"):], "x")
	if len(dimensions) != 2 {
		return "", config.ThumbnailSize{}, false
	}
	width, widthErr := strconv.Atoi(dimensions[0])
	height, heightErr := strconv.Atoi(dimensions[1])
	if heightErr != nil || widthErr != nil || height <= 0 || width <= 0 {
		return "", config.ThumbnailSize{}, false
	}
	size := config.ThumbnailSize{Width: width, Height: height}
	allowed := slices.Contains(sizes, size)
	if !allowed {
		return "", config.ThumbnailSize{}, false
	}
	originalID := base[:index] + ext
	return originalID, size, intimages.ValidID(originalID)
}

func ServeMissingImage(w http.ResponseWriter, r *http.Request, cfg *config.RuntimeConfig) {
	var opts []templates.Option
	if cfg != nil && cfg.TemplatesDir != "" {
		opts = append(opts, templates.WithDir(cfg.TemplatesDir))
	}
	w.Header().Set("Content-Type", "image/svg+xml")
	http.ServeContent(w, r, "missing_image.svg", time.Time{}, bytes.NewReader(templates.GetMissingImageData(opts...)))
}

func serveMissingCacheImage(w http.ResponseWriter, r *http.Request, cfg *config.RuntimeConfig, id string, entry *db.ImageCacheEntry) {
	var opts []templates.Option
	if cfg != nil && cfg.TemplatesDir != "" {
		opts = append(opts, templates.WithDir(cfg.TemplatesDir))
	}
	data := missingCacheImageData(cfg, id, entry)
	w.Header().Set("Content-Type", "image/svg+xml")
	http.ServeContent(w, r, "missing_image.svg", time.Time{}, bytes.NewReader(templates.GetMissingImageSVG(data, opts...)))
}

func missingCacheImageData(cfg *config.RuntimeConfig, id string, entry *db.ImageCacheEntry) templates.MissingImageData {
	minWidth := imageCachePlaceholderMinWidth(cfg)
	minHeight := imageCachePlaceholderMinHeight(cfg)
	data := templates.MissingImageData{
		Title:  "Image Pending",
		Line1:  "cache:" + id,
		Line2:  "waiting for download",
		Width:  minWidth,
		Height: minHeight,
	}
	if entry == nil {
		data.Title = "Missing Image"
		data.DetailLines = wrapMissingImageLines(data.Width, "cache: "+id, "metadata: not found")
		return data
	}
	if entry.Width.Valid && entry.Width.Int32 > 0 {
		data.Width = max(int(entry.Width.Int32), minWidth)
	}
	if entry.Height.Valid && entry.Height.Int32 > 0 {
		data.Height = max(int(entry.Height.Int32), minHeight)
	}
	var detailLines []string
	detailLines = append(detailLines, "cache: "+id)
	switch entry.Status {
	case "failed":
		data.Title = "Image Unavailable"
		data.Line2 = "download failed"
	case "pending":
		data.Title = "Image Pending"
		data.Line2 = "waiting for download"
	default:
		data.Title = "Missing Image"
		data.Line2 = entry.Status
	}
	if entry.Status != "" {
		detailLines = append(detailLines, "status: "+entry.Status)
	}
	if entry.SourceUrl.Valid && entry.SourceUrl.String != "" {
		data.Line1 = originSummary(entry.SourceUrl.String)
		data.Description = "Original image: " + entry.SourceUrl.String
		detailLines = append(detailLines, wrapMissingImageLine(data.Width, "source: ", entry.SourceUrl.String)...)
	}
	if entry.ErrorMessage.Valid && entry.ErrorMessage.String != "" {
		data.Line2 = fmt.Sprintf("attempt %d failed: %s", entry.RetryCount, entry.ErrorMessage.String)
		if entry.NextAttemptAt.Valid {
			data.Line2 = fmt.Sprintf("attempt %d failed; retry %s", entry.RetryCount, entry.NextAttemptAt.Time.Format(time.RFC3339))
		}
		if data.Description != "" {
			data.Description += "\n"
		}
		data.Description += "Last error: " + entry.ErrorMessage.String
		if entry.LastAttemptAt.Valid {
			data.Description += "\nLast attempt: " + entry.LastAttemptAt.Time.Format(time.RFC3339)
		}
		detailLines = append(detailLines, fmt.Sprintf("attempts: %d", entry.RetryCount))
		if entry.LastAttemptAt.Valid {
			detailLines = append(detailLines, "last attempt: "+entry.LastAttemptAt.Time.Format(time.RFC3339))
		}
		if entry.NextAttemptAt.Valid {
			detailLines = append(detailLines, "next retry: "+entry.NextAttemptAt.Time.Format(time.RFC3339))
		}
		detailLines = append(detailLines, wrapMissingImageLine(data.Width, "error: ", entry.ErrorMessage.String)...)
	}
	if entry.ErrorMessage.Valid && entry.ErrorMessage.String != "" {
		data.DetailLines = trimMissingImageLines(detailLines)
		return data
	}
	if entry.ContentType.Valid && entry.SizeBytes.Valid {
		data.Line2 = fmt.Sprintf("%s, %d bytes", entry.ContentType.String, entry.SizeBytes.Int64)
	} else if entry.ContentType.Valid {
		data.Line2 = entry.ContentType.String
	} else if entry.SizeBytes.Valid {
		data.Line2 = fmt.Sprintf("%d bytes", entry.SizeBytes.Int64)
	}
	if data.Line2 != "" {
		detailLines = append(detailLines, data.Line2)
	}
	data.DetailLines = trimMissingImageLines(detailLines)
	return data
}

func imageCachePlaceholderMinWidth(cfg *config.RuntimeConfig) int {
	if cfg == nil || cfg.ImageCachePlaceholderMinWidth <= 0 {
		return config.DefaultImageCachePlaceholderMinWidth
	}
	return cfg.ImageCachePlaceholderMinWidth
}

func imageCachePlaceholderMinHeight(cfg *config.RuntimeConfig) int {
	if cfg == nil || cfg.ImageCachePlaceholderMinHeight <= 0 {
		return config.DefaultImageCachePlaceholderMinHeight
	}
	return cfg.ImageCachePlaceholderMinHeight
}

func wrapMissingImageLines(width int, lines ...string) []string {
	var out []string
	for _, line := range lines {
		out = append(out, wrapMissingImageLine(width, "", line)...)
	}
	return trimMissingImageLines(out)
}

func wrapMissingImageLine(width int, prefix, value string) []string {
	value = strings.TrimSpace(value)
	prefix = strings.TrimSpace(prefix)
	if prefix != "" {
		prefix += " "
	}
	if value == "" {
		return nil
	}
	maxChars := max((width-48)/7, 32)
	if len(prefix)+len(value) <= maxChars {
		return []string{prefix + value}
	}
	var out []string
	remaining := value
	firstPrefix := prefix
	nextPrefix := strings.Repeat(" ", len(prefix))
	for remaining != "" {
		linePrefix := firstPrefix
		if len(out) > 0 {
			linePrefix = nextPrefix
		}
		limit := max(maxChars-len(linePrefix), 12)
		if len(remaining) <= limit {
			out = append(out, linePrefix+remaining)
			break
		}
		cut := strings.LastIndexAny(remaining[:limit], "/?&=.-_")
		if cut < limit/2 {
			cut = limit
		} else {
			cut++
		}
		out = append(out, linePrefix+remaining[:cut])
		remaining = strings.TrimLeft(remaining[cut:], " ")
	}
	return out
}

func trimMissingImageLines(lines []string) []string {
	const maxPlaceholderDetailLines = 10 // Maximum diagnostic rows shown inside the SVG placeholder.
	if len(lines) <= maxPlaceholderDetailLines {
		return lines
	}
	return append(lines[:maxPlaceholderDetailLines-1], "…")
}

func originSummary(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return raw
	}
	if u.Path == "" || u.Path == "/" {
		return u.Host
	}
	return u.Host + "/" + path.Base(u.Path)
}

// Register registers the images router module.
func Register(reg *router.Registry) {
	reg.RegisterModule("images", nil, func(r *mux.Router, cfg *config.RuntimeConfig) []nav.RouterOptions {
		return RegisterRoutes(r, cfg)
	})
}
