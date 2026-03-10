package handler

import (
	"encoding/csv"
	"fmt"
	"net/http"

	"goaltrack/internal/middleware"
	"goaltrack/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ExportHandler struct {
	db *gorm.DB
}

func NewExportHandler(db *gorm.DB) *ExportHandler {
	return &ExportHandler{db: db}
}

func (h *ExportHandler) ExportGoals(c *gin.Context) {
	uid := middleware.GetUserID(c)

	var goals []model.Goal
	h.db.Where("user_id = ?", uid).Find(&goals)

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=goals.csv")
	c.Writer.Write([]byte{0xEF, 0xBB, 0xBF}) // UTF-8 BOM

	w := csv.NewWriter(c.Writer)
	w.Write([]string{"目标名称", "分类", "目标值", "当前值", "单位", "进度%", "状态", "截止日期"})

	for _, g := range goals {
		w.Write([]string{
			g.Title,
			g.Category,
			fmt.Sprintf("%.1f", g.TargetValue),
			fmt.Sprintf("%.1f", g.CurrentValue),
			g.Unit,
			fmt.Sprintf("%.0f", g.Progress()),
			g.Status,
			g.Deadline.Format("2006-01-02"),
		})
	}
	w.Flush()
	c.Status(http.StatusOK)
}

func (h *ExportHandler) ExportHabits(c *gin.Context) {
	uid := middleware.GetUserID(c)

	var checkIns []model.HabitCheckIn
	h.db.Preload("Habit").Where("user_id = ?", uid).Order("check_date DESC").Find(&checkIns)

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=habits.csv")
	c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	w := csv.NewWriter(c.Writer)
	w.Write([]string{"日期", "习惯ID", "备注"})

	for _, ci := range checkIns {
		w.Write([]string{
			ci.CheckDate.Format("2006-01-02"),
			fmt.Sprintf("%d", ci.HabitID),
			ci.Note,
		})
	}
	w.Flush()
	c.Status(http.StatusOK)
}
