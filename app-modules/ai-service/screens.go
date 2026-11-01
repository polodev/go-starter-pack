package aiservice

import (
	"strconv"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/http"

	adminlayout "goravel/app-modules/admin-layout"
	"goravel/app-modules/ai-service/models"
	"goravel/app/facades"
)

// Index is GET /admin/ai-api-logs: newest first, ?provider=, ?status= and ?q= (purpose or model).
func Index(ctx http.Context) http.Response {
	query := facades.Orm().Query().Model(&models.AiApiLog{}).OrderBy("id", "desc")
	filters := map[string]string{}
	for _, field := range []string{"provider", "status", "q"} {
		filters[field] = strings.TrimSpace(ctx.Request().Query(field))
	}
	if filters["provider"] != "" {
		query = query.Where("provider", filters["provider"])
	}
	if filters["status"] != "" {
		query = query.Where("status", filters["status"])
	}
	if filters["q"] != "" {
		query = query.Where("(purpose LIKE ? OR model LIKE ?)", "%"+filters["q"]+"%", "%"+filters["q"]+"%")
	}
	var logs []models.AiApiLog
	var total int64
	_ = query.Paginate(adminlayout.PageParam(ctx), 50, &logs, &total)
	var spent struct{ Total *float64 }
	_ = facades.DB().Table("ai_api_logs").Select("SUM(cost) as total").Where("created_at >= ?", time.Now().UTC().AddDate(0, 0, -30)).First(&spent)
	return adminlayout.View(ctx, "ai-service::index", "AI API logs", []adminlayout.Crumb{{Label: "AI API logs"}}, map[string]any{
		"logs": logs, "filters": filters, "total": total, "hasSpent": spent.Total != nil, "spent": valueOf(spent.Total),
	})
}

func find(ctx http.Context) (models.AiApiLog, bool) {
	var row models.AiApiLog
	id := adminlayout.IDParam(ctx)
	if id == 0 || facades.Orm().Query().Find(&row, id) != nil || row.ID == 0 {
		return row, false
	}
	return row, true
}

// Show is GET /admin/ai-api-logs/{id}.
func Show(ctx http.Context) http.Response {
	row, ok := find(ctx)
	if !ok {
		return adminlayout.NotFound(ctx)
	}
	return adminlayout.View(ctx, "ai-service::show", "AI API log "+strconv.Itoa(int(row.ID)),
		[]adminlayout.Crumb{{Label: "AI API logs", Href: "/admin/ai-api-logs"}, {Label: strconv.Itoa(int(row.ID))}},
		map[string]any{"row": row, "hasCost": row.Cost != nil, "cost": valueOf(row.Cost), "prompt": models.Text(row.Prompt), "response": models.Text(row.Response), "error": models.Text(row.Error)})
}

// Delete is POST /admin/ai-api-logs/{id}/delete (ai_api_logs.manage).
func Delete(ctx http.Context) http.Response {
	row, ok := find(ctx)
	if !ok {
		return adminlayout.NotFound(ctx)
	}
	if _, err := facades.Orm().Query().Delete(&row); err != nil {
		return ctx.Response().Redirect(http.StatusSeeOther, "/admin/ai-api-logs?error=delete")
	}
	return ctx.Response().Redirect(http.StatusSeeOther, "/admin/ai-api-logs?saved=deleted")
}

// PurgeOlderThan deletes AI API logs older than days; returns how many.
func PurgeOlderThan(days int) (int64, error) {
	result, err := facades.Orm().Query().Where("created_at < ?", time.Now().UTC().AddDate(0, 0, -days)).Delete(&models.AiApiLog{})
	if err != nil {
		return 0, err
	}
	return result.RowsAffected, nil
}

func valueOf(number *float64) float64 {
	if number == nil {
		return 0
	}
	return *number
}
