package router

import (
	"net/http"

	"github.com/schliz/deckel/internal/auth"
	"github.com/schliz/deckel/internal/handler/admin"
)

// RegisterAdmin wires the /admin/* routes. All routes require the admin
// role (auth.RequireAdmin) on top of CSRF + base middleware.
func RegisterAdmin(mux *http.ServeMux, h *admin.Handler, withCSRF func(http.Handler) http.Handler) {
	adminOnly := func(handler http.Handler) http.Handler {
		return withCSRF(auth.RequireAdmin(handler))
	}

	mux.Handle("GET /admin/menu", adminOnly(h.Wrap(h.AdminMenuPage)))
	mux.Handle("GET /admin/menu/batch", adminOnly(h.Wrap(h.MenuBatchPage)))
	mux.Handle("GET /admin/menu/batch/export", adminOnly(h.Wrap(h.MenuBatchExport)))
	mux.Handle("POST /admin/menu/batch/upload", adminOnly(h.Wrap(h.MenuBatchUpload)))
	mux.Handle("POST /admin/menu/batch/apply", adminOnly(h.Wrap(h.MenuBatchApply)))
	mux.Handle("POST /admin/categories", adminOnly(h.Wrap(h.CreateCategory)))
	mux.Handle("GET /admin/categories/{id}/edit", adminOnly(h.Wrap(h.EditCategoryForm)))
	mux.Handle("POST /admin/categories/{id}/update", adminOnly(h.Wrap(h.UpdateCategory)))
	mux.Handle("POST /admin/categories/{id}/reorder", adminOnly(h.Wrap(h.ReorderCategory)))
	mux.Handle("DELETE /admin/categories/{id}", adminOnly(h.Wrap(h.DeleteCategory)))
	mux.Handle("POST /admin/categories/{id}/items", adminOnly(h.Wrap(h.CreateItem)))
	mux.Handle("GET /admin/items/{id}/edit", adminOnly(h.Wrap(h.EditItemForm)))
	mux.Handle("POST /admin/items/{id}/update", adminOnly(h.Wrap(h.UpdateItem)))
	mux.Handle("POST /admin/items/{id}/reorder", adminOnly(h.Wrap(h.ReorderItem)))
	mux.Handle("POST /admin/items/{id}/delete", adminOnly(h.Wrap(h.SoftDeleteItem)))
	mux.Handle("GET /admin/users", adminOnly(h.Wrap(h.AdminUserList)))
	mux.Handle("GET /admin/users/{id}/confirm-toggle", adminOnly(h.Wrap(h.ConfirmToggleModal)))
	mux.Handle("POST /admin/users/{id}/toggle-barteamer", adminOnly(h.Wrap(h.ToggleBarteamer)))
	mux.Handle("POST /admin/users/{id}/toggle-active", adminOnly(h.Wrap(h.ToggleActive)))
	mux.Handle("POST /admin/users/{id}/toggle-spending-limit", adminOnly(h.Wrap(h.ToggleSpendingLimit)))
	mux.Handle("GET /admin/users/{id}/deposit", adminOnly(h.Wrap(h.DepositModal)))
	mux.Handle("POST /admin/users/{id}/deposit", adminOnly(h.Wrap(h.RegisterDeposit)))
	mux.Handle("GET /admin/transactions", adminOnly(h.Wrap(h.AdminTransactionList)))
	mux.Handle("GET /admin/transactions/{id}/cancel", adminOnly(h.Wrap(h.AdminCancelModal)))
	mux.Handle("POST /admin/transactions/{id}/cancel", adminOnly(h.Wrap(h.AdminCancelTransaction)))
	mux.Handle("GET /admin/stats", adminOnly(h.Wrap(h.AdminStatsPage)))
	mux.Handle("GET /admin/settings", adminOnly(h.Wrap(h.AdminSettingsPage)))
	mux.Handle("POST /admin/settings", adminOnly(h.Wrap(h.SaveSettings)))
	mux.Handle("POST /admin/settings/send-reminders", adminOnly(h.Wrap(h.SendReminders)))
}
