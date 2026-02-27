package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/howallet/howallet/internal/middleware"
	"github.com/howallet/howallet/internal/service"
)

type ExportHandler struct {
	exportSvc *service.ExportService
	hhSvc     *service.HouseholdService
}

func NewExportHandler(exportSvc *service.ExportService, hhSvc *service.HouseholdService) *ExportHandler {
	return &ExportHandler{exportSvc: exportSvc, hhSvc: hhSvc}
}

// GET /api/export/csv
func (h *ExportHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromCtx(r.Context())
	hhID := middleware.HouseholdIDFromCtx(r.Context())

	if err := h.hhSvc.CheckMembership(r.Context(), hhID, userID); err != nil {
		ErrorJSON(w, http.StatusForbidden, "not a member of this household")
		return
	}

	var from, to *time.Time
	if v := r.URL.Query().Get("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			from = &t
		}
	}
	if v := r.URL.Query().Get("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			to = &t
		}
	}

	filename := fmt.Sprintf("hoWallet_export_%s.csv", time.Now().Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	if err := h.exportSvc.ExportCSV(r.Context(), w, hhID, from, to); err != nil {
		// Headers already sent, just log
		http.Error(w, "export failed", http.StatusInternalServerError)
	}
}
