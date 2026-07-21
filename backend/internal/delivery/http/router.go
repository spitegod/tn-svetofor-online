package http

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"tn/backend/internal/apperror"
	"tn/backend/internal/model"
	"tn/backend/internal/service"
)

type Router struct {
	classification  *service.ClassificationService
	systemCatalog   *service.SystemCatalogService
	systemDocuments *service.SystemDocumentService
	orders          *service.OrderService
	navParser       *service.NavParserService
	healthChecker   healthChecker
}

type healthChecker interface {
	PingContext(context.Context) error
}

const (
	maxJSONBodySize       = 4 << 20
	maxTableUploadSize    = 32 << 20
	maxWorkbookUploadSize = 64 << 20
	maxFilterValueBytes   = 500
)

func NewRouter(classification *service.ClassificationService, systemCatalog *service.SystemCatalogService, systemDocuments *service.SystemDocumentService, orders *service.OrderService, navParser *service.NavParserService, healthChecker healthChecker) http.Handler {
	router := &Router{classification: classification, systemCatalog: systemCatalog, systemDocuments: systemDocuments, orders: orders, navParser: navParser, healthChecker: healthChecker}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", router.health)
	mux.HandleFunc("GET /api/orders", router.listOrders)
	mux.HandleFunc("POST /api/orders", router.createOrder)
	mux.HandleFunc("POST /api/orders/import", router.importOrderWorkbook)
	mux.HandleFunc("PATCH /api/orders/{id}", router.updateOrder)
	mux.HandleFunc("DELETE /api/orders/{id}", router.deleteOrder)
	mux.HandleFunc("GET /api/classification-changes", router.listClassificationChanges)
	mux.HandleFunc("PATCH /api/classification-changes/{id}", router.updateClassificationChange)
	mux.HandleFunc("POST /api/classification-changes/import", router.importClassificationChanges)
	mux.HandleFunc("GET /api/classification-changes/export", router.exportClassificationChanges)
	mux.HandleFunc("GET /api/system-catalog", router.listSystemCatalog)
	mux.HandleFunc("PATCH /api/system-catalog/{id}", router.updateSystemCatalogRow)
	mux.HandleFunc("POST /api/system-catalog/import", router.importSystemCatalog)
	mux.HandleFunc("GET /api/system-catalog/export", router.exportSystemCatalog)
	mux.HandleFunc("POST /api/system-catalog/parse-nav", router.parseNavSystemCatalog)
	mux.HandleFunc("POST /api/nav-parser/runs", router.parseNavSystemCatalog)
	mux.HandleFunc("POST /api/nav-parser/cancel", router.cancelNavParser)
	mux.HandleFunc("GET /api/nav-parser/status", router.navParserStatus)
	mux.HandleFunc("GET /api/nav-parser/runs", router.navParserRuns)
	mux.HandleFunc("GET /api/nav-system-types/{slug}/image", router.navSystemTypeImage)
	mux.HandleFunc("GET /api/nav-parser/settings", router.navParserSettings)
	mux.HandleFunc("PATCH /api/nav-parser/settings", router.updateNavParserSettings)
	mux.HandleFunc("GET /api/system-documents", router.listSystemDocuments)
	mux.HandleFunc("GET /api/system-documents/export", router.exportSystemDocuments)
	mux.HandleFunc("GET /api/system-documents/history", router.systemDocumentHistory)
	mux.HandleFunc("PATCH /api/system-documents/{id}", router.updateSystemDocument)
	mux.HandleFunc("POST /api/system-documents/{id}/attachment", router.uploadSystemDocumentAttachment)
	mux.HandleFunc("GET /api/system-documents/{id}/attachment", router.getSystemDocumentAttachment)
	mux.HandleFunc("DELETE /api/system-documents/{id}/attachment", router.deleteSystemDocumentAttachment)
	mux.HandleFunc("PATCH /api/system-documents/{id}/comparison", router.updateSystemDocumentComparison)
	mux.HandleFunc("PATCH /api/system-documents/comparison", router.updateSystemDocumentComparisonBulk)
	mux.HandleFunc("POST /api/comparison/export", router.exportComparison)

	return requestLoggingMiddleware(recoverMiddleware(securityHeadersMiddleware(mux)))
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	if r.status != 0 {
		return
	}
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(payload []byte) (int, error) {
	if r.status == 0 {
		r.WriteHeader(http.StatusOK)
	}
	return r.ResponseWriter.Write(payload)
}

func requestLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		startedAt := time.Now()
		requestID := newRequestID()
		w.Header().Set("X-Request-ID", requestID)
		recorder := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(recorder, request)
		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}
		log.Printf("request_id=%s method=%s path=%s status=%d duration=%s", requestID,
			request.Method, request.URL.Path, status, time.Since(startedAt).Round(time.Millisecond))
	})
}

func newRequestID() string {
	var value [12]byte
	if _, err := rand.Read(value[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return hex.EncodeToString(value[:])
}

func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		next.ServeHTTP(w, request)
	})
}

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("panic while serving %s %s: %v", request.Method, request.URL.Path, recovered)
				writeError(w, http.StatusInternalServerError, fmt.Errorf("internal server error"))
			}
		}()
		next.ServeHTTP(w, request)
	})
}

func (r *Router) exportComparison(w http.ResponseWriter, request *http.Request) {
	var payload model.ComparisonExport
	if err := decodeJSON(w, request, &payload); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("decode comparison export: %w", err))
		return
	}

	data, err := r.systemDocuments.ExportComparison(request.Context(), payload)
	if err != nil {
		writeActionError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", `attachment; filename="comparison.xlsx"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (r *Router) listSystemDocuments(w http.ResponseWriter, request *http.Request) {
	filter, ok := systemDocumentFilterFromRequest(w, request)
	if !ok {
		return
	}
	payload, err := r.systemDocuments.List(request.Context(), filter)
	if err != nil {
		writeActionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, payload)
}

func (r *Router) exportSystemDocuments(w http.ResponseWriter, request *http.Request) {
	filter, ok := systemDocumentFilterFromRequest(w, request)
	if !ok {
		return
	}
	payload, err := r.systemDocuments.Export(request.Context(), filter)
	if err != nil {
		writeActionError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", `attachment; filename="system-documents.xlsx"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
}

func (r *Router) systemDocumentHistory(w http.ResponseWriter, request *http.Request) {
	payload, err := r.systemDocuments.History(request.Context(), request.URL.Query().Get("code"), request.URL.Query().Get("systemName"))
	if err != nil {
		writeActionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, payload)
}

func (r *Router) updateSystemDocument(w http.ResponseWriter, request *http.Request) {
	id, err := strconv.ParseInt(request.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid system document id: %w", err))
		return
	}
	var payload struct {
		Comment string `json:"comment"`
	}
	if err := decodeJSON(w, request, &payload); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("decode system document: %w", err))
		return
	}
	orderID, ok := requireOrderID(w, request)
	if !ok {
		return
	}
	row, err := r.systemDocuments.UpdateComment(request.Context(), id, orderID, payload.Comment)
	if err != nil {
		writeActionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, row)
}

func (r *Router) uploadSystemDocumentAttachment(w http.ResponseWriter, request *http.Request) {
	id, err := strconv.ParseInt(request.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid system document id: %w", err))
		return
	}
	request.Body = http.MaxBytesReader(w, request.Body, service.MaxSystemDocumentAttachmentSize+(1<<20))
	if err := request.ParseMultipartForm(service.MaxSystemDocumentAttachmentSize); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("read attachment: %w", err))
		return
	}
	if request.MultipartForm != nil {
		defer request.MultipartForm.RemoveAll()
	}
	file, header, err := request.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("attachment file is required: %w", err))
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, service.MaxSystemDocumentAttachmentSize+1))
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("read attachment file: %w", err))
		return
	}
	contentType := header.Header.Get("Content-Type")
	if contentType == "" && len(data) > 0 {
		contentType = http.DetectContentType(data)
	}
	attachment := model.SystemDocumentAttachment{
		Name:        header.Filename,
		ContentType: contentType,
		Data:        data,
	}
	orderID, ok := requireOrderID(w, request)
	if !ok {
		return
	}
	if err := r.systemDocuments.SaveAttachment(request.Context(), id, orderID, attachment); err != nil {
		writeActionError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (r *Router) getSystemDocumentAttachment(w http.ResponseWriter, request *http.Request) {
	id, err := strconv.ParseInt(request.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid system document id: %w", err))
		return
	}
	orderID, ok := requireOrderID(w, request)
	if !ok {
		return
	}
	attachment, err := r.systemDocuments.Attachment(request.Context(), id, orderID)
	if err != nil {
		writeActionError(w, err)
		return
	}
	disposition := "attachment"
	if attachment.ContentType == "application/pdf" {
		disposition = "inline"
	}
	w.Header().Set("Content-Type", attachment.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(attachment.Size, 10))
	w.Header().Set("Content-Disposition", mime.FormatMediaType(disposition, map[string]string{"filename": attachment.Name}))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(attachment.Data)
}

func (r *Router) deleteSystemDocumentAttachment(w http.ResponseWriter, request *http.Request) {
	id, err := strconv.ParseInt(request.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid system document id: %w", err))
		return
	}
	orderID, ok := requireOrderID(w, request)
	if !ok {
		return
	}
	if err := r.systemDocuments.DeleteAttachment(request.Context(), id, orderID); err != nil {
		writeActionError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (r *Router) updateSystemDocumentComparison(w http.ResponseWriter, request *http.Request) {
	id, err := strconv.ParseInt(request.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid system document id: %w", err))
		return
	}
	var payload struct {
		Selected bool `json:"selected"`
	}
	if err := decodeJSON(w, request, &payload); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("decode comparison selection: %w", err))
		return
	}
	orderID, ok := requireOrderID(w, request)
	if !ok {
		return
	}
	if err := r.systemDocuments.UpdateComparison(request.Context(), id, orderID, payload.Selected); err != nil {
		writeActionError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (r *Router) updateSystemDocumentComparisonBulk(w http.ResponseWriter, request *http.Request) {
	var payload struct {
		Selected  bool                      `json:"selected"`
		AllOrders bool                      `json:"allOrders"`
		Systems   []model.SystemDocumentKey `json:"systems"`
	}
	if err := decodeJSON(w, request, &payload); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("decode bulk comparison selection: %w", err))
		return
	}
	orderID, ok := requireOrderID(w, request)
	if !ok {
		return
	}
	if err := r.systemDocuments.UpdateComparisonBulk(request.Context(), orderID, payload.AllOrders, payload.Selected, payload.Systems); err != nil {
		writeActionError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (r *Router) parseNavSystemCatalog(w http.ResponseWriter, request *http.Request) {
	if err := r.navParser.StartManual(); err != nil {
		if errors.Is(err, service.ErrNavParserRunning) {
			writeError(w, http.StatusConflict, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "started"})
}

func (r *Router) cancelNavParser(w http.ResponseWriter, _ *http.Request) {
	if err := r.navParser.Cancel(); err != nil {
		if errors.Is(err, service.ErrNavParserNotRunning) {
			writeError(w, http.StatusConflict, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (r *Router) navParserStatus(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, r.navParser.Status())
}

func (r *Router) navParserRuns(w http.ResponseWriter, request *http.Request) {
	limit := 20
	if rawLimit := request.URL.Query().Get("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil || parsedLimit < 1 || parsedLimit > 100 {
			writeError(w, http.StatusBadRequest, fmt.Errorf("limit must be between 1 and 100"))
			return
		}
		limit = parsedLimit
	}
	runs, err := r.navParser.Runs(request.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, runs)
}

func (r *Router) navSystemTypeImage(w http.ResponseWriter, request *http.Request) {
	image, err := r.navParser.SystemTypeImage(request.Context(), request.PathValue("slug"))
	if err != nil {
		writeActionError(w, err)
		return
	}
	w.Header().Set("Content-Type", image.ContentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(image.Data)))
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(image.Data)
}

func (r *Router) navParserSettings(w http.ResponseWriter, request *http.Request) {
	settings, err := r.navParser.Settings(request.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

func (r *Router) updateNavParserSettings(w http.ResponseWriter, request *http.Request) {
	var payload model.NavParserSettings
	if err := decodeJSON(w, request, &payload); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid nav parser settings"))
		return
	}
	settings, err := r.navParser.UpdateSettings(request.Context(), payload)
	if err != nil {
		writeActionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

func (r *Router) health(w http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(request.Context(), 2*time.Second)
	defer cancel()
	if err := r.healthChecker.PingContext(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "unavailable",
			"error":  "database is unavailable",
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

func (r *Router) listOrders(w http.ResponseWriter, request *http.Request) {
	payload, err := r.orders.List(request.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, payload)
}

func (r *Router) createOrder(w http.ResponseWriter, request *http.Request) {
	var payload struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(w, request, &payload); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("decode order: %w", err))
		return
	}

	order, err := r.orders.Create(request.Context(), payload.Name)
	if err != nil {
		writeActionError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, order)
}

func (r *Router) importOrderWorkbook(w http.ResponseWriter, request *http.Request) {
	request.Body = http.MaxBytesReader(w, request.Body, maxWorkbookUploadSize)
	if err := request.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("parse multipart form: %w", err))
		return
	}
	if request.MultipartForm != nil {
		defer request.MultipartForm.RemoveAll()
	}

	file, header, err := request.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("read uploaded file: %w", err))
		return
	}
	defer file.Close()

	fileName := path.Base(strings.ReplaceAll(header.Filename, "\\", "/"))
	if !strings.EqualFold(path.Ext(fileName), ".xlsx") {
		writeError(w, http.StatusBadRequest, fmt.Errorf("order workbook must be an XLSX file"))
		return
	}
	orderName := strings.TrimSuffix(fileName, path.Ext(fileName))

	order, err := r.orders.ImportWorkbook(request.Context(), orderName, file)
	if err != nil {
		writeActionError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, order)
}

func (r *Router) updateOrder(w http.ResponseWriter, request *http.Request) {
	id, err := strconv.ParseInt(request.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid order id: %w", err))
		return
	}

	var payload struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(w, request, &payload); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("decode order: %w", err))
		return
	}

	order, err := r.orders.UpdateName(request.Context(), id, payload.Name)
	if err != nil {
		writeActionError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, order)
}

func (r *Router) deleteOrder(w http.ResponseWriter, request *http.Request) {
	id, err := strconv.ParseInt(request.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid order id: %w", err))
		return
	}

	if err := r.orders.Delete(request.Context(), id); err != nil {
		writeActionError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (r *Router) listClassificationChanges(w http.ResponseWriter, request *http.Request) {
	filter, ok := classificationFilterFromRequest(w, request)
	if !ok {
		return
	}
	payload, err := r.classification.List(request.Context(), filter)
	if err != nil {
		writeActionError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, payload)
}

func (r *Router) updateClassificationChange(w http.ResponseWriter, request *http.Request) {
	id, err := strconv.ParseInt(request.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid classification change id: %w", err))
		return
	}
	var payload model.ClassificationChange
	if err := decodeJSON(w, request, &payload); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("decode classification change: %w", err))
		return
	}
	orderID, ok := requireOrderID(w, request)
	if !ok {
		return
	}
	row, err := r.classification.Update(request.Context(), id, orderID, payload)
	if err != nil {
		writeActionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, row)
}

func (r *Router) importClassificationChanges(w http.ResponseWriter, request *http.Request) {
	request.Body = http.MaxBytesReader(w, request.Body, maxTableUploadSize)
	if err := request.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("parse multipart form: %w", err))
		return
	}
	if request.MultipartForm != nil {
		defer request.MultipartForm.RemoveAll()
	}

	file, _, err := request.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("read uploaded file: %w", err))
		return
	}
	defer file.Close()

	orderID, ok := requireOrderID(w, request)
	if !ok {
		return
	}
	payload, err := r.classification.Import(request.Context(), orderID, file)
	if err != nil {
		writeActionError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, payload)
}

func (r *Router) exportClassificationChanges(w http.ResponseWriter, request *http.Request) {
	filter, ok := classificationFilterFromRequest(w, request)
	if !ok {
		return
	}
	payload, err := r.classification.Export(request.Context(), filter)
	if err != nil {
		writeActionError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", `attachment; filename="classification-changes.xlsx"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
}

func (r *Router) listSystemCatalog(w http.ResponseWriter, request *http.Request) {
	filter, ok := systemCatalogFilterFromRequest(w, request)
	if !ok {
		return
	}
	payload, err := r.systemCatalog.List(request.Context(), filter)
	if err != nil {
		writeActionError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, payload)
}

func (r *Router) updateSystemCatalogRow(w http.ResponseWriter, request *http.Request) {
	id, err := strconv.ParseInt(request.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid system catalog row id: %w", err))
		return
	}
	var payload model.SystemCatalogRow
	if err := decodeJSON(w, request, &payload); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("decode system catalog row: %w", err))
		return
	}
	orderID, ok := requireOrderID(w, request)
	if !ok {
		return
	}
	row, err := r.systemCatalog.Update(request.Context(), id, orderID, payload)
	if err != nil {
		writeActionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, row)
}

func (r *Router) importSystemCatalog(w http.ResponseWriter, request *http.Request) {
	request.Body = http.MaxBytesReader(w, request.Body, maxTableUploadSize)
	if err := request.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("parse multipart form: %w", err))
		return
	}
	if request.MultipartForm != nil {
		defer request.MultipartForm.RemoveAll()
	}

	file, _, err := request.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("read uploaded file: %w", err))
		return
	}
	defer file.Close()

	orderID, ok := requireOrderID(w, request)
	if !ok {
		return
	}
	payload, err := r.systemCatalog.Import(request.Context(), orderID, file)
	if err != nil {
		writeActionError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, payload)
}

func (r *Router) exportSystemCatalog(w http.ResponseWriter, request *http.Request) {
	filter, ok := systemCatalogFilterFromRequest(w, request)
	if !ok {
		return
	}
	payload, err := r.systemCatalog.Export(request.Context(), filter)
	if err != nil {
		writeActionError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", `attachment; filename="system-catalog.xlsx"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
}

func classificationFilterFromRequest(w http.ResponseWriter, request *http.Request) (model.ClassificationFilter, bool) {
	query := request.URL.Query()
	orderID, ok := requireOrderID(w, request)
	if !ok {
		return model.ClassificationFilter{}, false
	}
	filter := model.ClassificationFilter{
		OrderID:          orderID,
		Query:            query.Get("q"),
		ConstructionType: query.Get("constructionType"),
		ClassBefore:      query.Get("before"),
		ClassAfter:       query.Get("after"),
	}

	if filter.ClassBefore == "Все" {
		filter.ClassBefore = ""
	}
	if filter.ClassAfter == "Все" {
		filter.ClassAfter = ""
	}
	if filter.ConstructionType == "Все" {
		filter.ConstructionType = ""
	}
	if !validateFilterValues(w, filter.Query, filter.ConstructionType, filter.ClassBefore, filter.ClassAfter) {
		return model.ClassificationFilter{}, false
	}

	return filter, true
}

func systemCatalogFilterFromRequest(w http.ResponseWriter, request *http.Request) (model.SystemCatalogFilter, bool) {
	query := request.URL.Query()
	orderID, ok := requireOrderID(w, request)
	if !ok {
		return model.SystemCatalogFilter{}, false
	}
	filter := model.SystemCatalogFilter{
		OrderID:     orderID,
		Query:       query.Get("q"),
		SystemClass: query.Get("class"),
		Curator:     query.Get("curator"),
	}

	if filter.SystemClass == "Все" {
		filter.SystemClass = ""
	}
	if filter.Curator == "Все" || filter.Curator == "Все кураторы" {
		filter.Curator = ""
	}
	if !validateFilterValues(w, filter.Query, filter.SystemClass, filter.Curator) {
		return model.SystemCatalogFilter{}, false
	}

	return filter, true
}

func systemDocumentFilterFromRequest(w http.ResponseWriter, request *http.Request) (model.SystemDocumentFilter, bool) {
	query := request.URL.Query()
	orderID, ok := requireOrderID(w, request)
	if !ok {
		return model.SystemDocumentFilter{}, false
	}
	filter := model.SystemDocumentFilter{
		OrderID:          orderID,
		Query:            query.Get("q"),
		SystemClass:      query.Get("class"),
		Curator:          query.Get("curator"),
		ConstructionType: query.Get("constructionType"),
		SystemType:       query.Get("systemType"),
		ComparisonOnly:   query.Get("comparison") == "true",
	}
	if filter.SystemClass == "Все" {
		filter.SystemClass = ""
	}
	if filter.Curator == "Все" || filter.Curator == "Все кураторы" {
		filter.Curator = ""
	}
	if !validateFilterValues(w, filter.Query, filter.SystemClass, filter.Curator, filter.ConstructionType, filter.SystemType) {
		return model.SystemDocumentFilter{}, false
	}
	return filter, true
}

func validateFilterValues(w http.ResponseWriter, values ...string) bool {
	for _, value := range values {
		if len(value) > maxFilterValueBytes {
			writeError(w, http.StatusUnprocessableEntity, fmt.Errorf("filter value is too long"))
			return false
		}
	}
	return true
}

func requireOrderID(w http.ResponseWriter, request *http.Request) (int64, bool) {
	id, err := orderIDFromRequest(request)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return 0, false
	}
	return id, true
}

func orderIDFromRequest(request *http.Request) (int64, error) {
	value := request.URL.Query().Get("orderId")
	if value == "" {
		value = request.FormValue("orderId")
	}
	if value == "" {
		return 1, nil
	}

	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("orderId must be a positive integer")
	}

	return id, nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func decodeJSON(w http.ResponseWriter, request *http.Request, destination any) error {
	request.Body = http.MaxBytesReader(w, request.Body, maxJSONBodySize)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return fmt.Errorf("request body must contain one JSON value")
		}
		return err
	}
	return nil
}

func writeError(w http.ResponseWriter, status int, err error) {
	message := err.Error()
	code := "bad_request"
	if status >= http.StatusInternalServerError {
		log.Printf("HTTP %d: %v", status, err)
		message = "Внутренняя ошибка сервера"
		code = "internal_error"
	} else {
		switch status {
		case http.StatusNotFound:
			code = "not_found"
		case http.StatusConflict:
			code = "conflict"
		case http.StatusUnprocessableEntity:
			code = "validation_error"
		}
	}
	writeJSON(w, status, map[string]string{"code": code, "error": message})
}

func writeActionError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	if kind, ok := apperror.KindOf(err); ok {
		switch kind {
		case apperror.Validation:
			status = http.StatusUnprocessableEntity
		case apperror.NotFound:
			status = http.StatusNotFound
		case apperror.Conflict:
			status = http.StatusConflict
		}
	}
	writeError(w, status, err)
}
