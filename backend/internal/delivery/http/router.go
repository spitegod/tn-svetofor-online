package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"tn/backend/internal/model"
	"tn/backend/internal/service"
)

type Router struct {
	classification *service.ClassificationService
	systemCatalog  *service.SystemCatalogService
	orders         *service.OrderService
	navParser      *service.NavParserService
}

func NewRouter(classification *service.ClassificationService, systemCatalog *service.SystemCatalogService, orders *service.OrderService, navParser *service.NavParserService) http.Handler {
	router := &Router{classification: classification, systemCatalog: systemCatalog, orders: orders, navParser: navParser}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", health)
	mux.HandleFunc("GET /api/orders", router.listOrders)
	mux.HandleFunc("POST /api/orders", router.createOrder)
	mux.HandleFunc("PATCH /api/orders/{id}", router.updateOrder)
	mux.HandleFunc("DELETE /api/orders/{id}", router.deleteOrder)
	mux.HandleFunc("GET /api/classification-changes", router.listClassificationChanges)
	mux.HandleFunc("POST /api/classification-changes/import", router.importClassificationChanges)
	mux.HandleFunc("GET /api/classification-changes/export", router.exportClassificationChanges)
	mux.HandleFunc("GET /api/system-catalog", router.listSystemCatalog)
	mux.HandleFunc("POST /api/system-catalog/import", router.importSystemCatalog)
	mux.HandleFunc("GET /api/system-catalog/export", router.exportSystemCatalog)
	mux.HandleFunc("POST /api/system-catalog/parse-nav", router.parseNavSystemCatalog)

	return mux
}

func (r *Router) parseNavSystemCatalog(w http.ResponseWriter, request *http.Request) {
	report, err := r.navParser.Parse(request.Context(), orderIDFromRequest(request))
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}

	writeJSON(w, http.StatusOK, report)
}

func health(w http.ResponseWriter, r *http.Request) {
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
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("decode order: %w", err))
		return
	}

	order, err := r.orders.Create(request.Context(), payload.Name)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
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
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("decode order: %w", err))
		return
	}

	order, err := r.orders.UpdateName(request.Context(), id, payload.Name)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
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
		writeError(w, http.StatusBadRequest, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (r *Router) listClassificationChanges(w http.ResponseWriter, request *http.Request) {
	payload, err := r.classification.List(request.Context(), classificationFilterFromRequest(request))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, payload)
}

func (r *Router) importClassificationChanges(w http.ResponseWriter, request *http.Request) {
	if err := request.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("parse multipart form: %w", err))
		return
	}

	file, _, err := request.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("read uploaded file: %w", err))
		return
	}
	defer file.Close()

	payload, err := r.classification.Import(request.Context(), orderIDFromRequest(request), file)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, payload)
}

func (r *Router) exportClassificationChanges(w http.ResponseWriter, request *http.Request) {
	payload, err := r.classification.Export(request.Context(), classificationFilterFromRequest(request))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", `attachment; filename="classification-changes.xlsx"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
}

func (r *Router) listSystemCatalog(w http.ResponseWriter, request *http.Request) {
	payload, err := r.systemCatalog.List(request.Context(), systemCatalogFilterFromRequest(request))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, payload)
}

func (r *Router) importSystemCatalog(w http.ResponseWriter, request *http.Request) {
	if err := request.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("parse multipart form: %w", err))
		return
	}

	file, _, err := request.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("read uploaded file: %w", err))
		return
	}
	defer file.Close()

	payload, err := r.systemCatalog.Import(request.Context(), orderIDFromRequest(request), file)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, payload)
}

func (r *Router) exportSystemCatalog(w http.ResponseWriter, request *http.Request) {
	payload, err := r.systemCatalog.Export(request.Context(), systemCatalogFilterFromRequest(request))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", `attachment; filename="system-catalog.xlsx"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
}

func classificationFilterFromRequest(request *http.Request) model.ClassificationFilter {
	query := request.URL.Query()
	filter := model.ClassificationFilter{
		OrderID:     orderIDFromRequest(request),
		Query:       query.Get("q"),
		ClassBefore: query.Get("before"),
		ClassAfter:  query.Get("after"),
	}

	if filter.ClassBefore == "Все" {
		filter.ClassBefore = ""
	}
	if filter.ClassAfter == "Все" {
		filter.ClassAfter = ""
	}

	return filter
}

func systemCatalogFilterFromRequest(request *http.Request) model.SystemCatalogFilter {
	query := request.URL.Query()
	filter := model.SystemCatalogFilter{
		OrderID:     orderIDFromRequest(request),
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

	return filter
}

func orderIDFromRequest(request *http.Request) int64 {
	value := request.URL.Query().Get("orderId")
	if value == "" {
		value = request.FormValue("orderId")
	}
	if value == "" {
		return 1
	}

	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 1
	}

	return id
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{
		"error": err.Error(),
	})
}
