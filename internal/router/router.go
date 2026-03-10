package router

import (
	"goaltrack/internal/handler"
	"goaltrack/internal/middleware"
	"goaltrack/internal/service"
	"goaltrack/pkg/jwt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Setup(r *gin.Engine, db *gorm.DB, jm *jwt.Manager, uploadSvc *service.UploadService) {
	authSvc := service.NewAuthService(db)
	goalSvc := service.NewGoalService(db)
	habitSvc := service.NewHabitService(db)
	todoSvc := service.NewTodoService(db)
	familySvc := service.NewFamilyService(db)
	dashSvc := service.NewDashboardService(db)
	projectSvc := service.NewProjectService(db)

	authH := handler.NewAuthHandler(authSvc, jm)
	goalH := handler.NewGoalHandler(goalSvc, dashSvc)
	habitH := handler.NewHabitHandler(habitSvc, dashSvc)
	todoH := handler.NewTodoHandler(todoSvc)
	familyH := handler.NewFamilyHandler(familySvc)
	dashH := handler.NewDashboardHandler(dashSvc)
	projectH := handler.NewProjectHandler(projectSvc, dashSvc)

	var uploadH *handler.UploadHandler
	if uploadSvc != nil {
		uploadH = handler.NewUploadHandler(uploadSvc)
	}

	api := r.Group("/api")

	// Public routes
	auth := api.Group("/auth")
	{
		auth.POST("/register", authH.Register)
		auth.POST("/login", authH.Login)
		auth.POST("/refresh", authH.RefreshToken)
	}

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.JWTAuth(jm))
	{
		// Profile
		protected.GET("/auth/profile", authH.GetProfile)
		protected.PUT("/auth/profile", authH.UpdateProfile)

		// Goals
		goals := protected.Group("/goals")
		{
			goals.GET("", goalH.List)
			goals.POST("", goalH.Create)
			goals.GET("/:id", goalH.GetByID)
			goals.PUT("/:id", goalH.Update)
			goals.DELETE("/:id", goalH.Archive)
			goals.GET("/:id/logs", goalH.GetLogs)
			goals.POST("/:id/logs", goalH.AddLog)
			goals.GET("/:id/milestones", goalH.GetMilestones)
			goals.POST("/:id/milestones", goalH.CreateMilestone)
			goals.GET("/:id/trend", goalH.GetTrend)
			goals.GET("/:id/predict", goalH.Predict)
		}

		// Habits
		habits := protected.Group("/habits")
		{
			habits.GET("", habitH.List)
			habits.POST("", habitH.Create)
			habits.GET("/today", habitH.TodayHabits)
			habits.PUT("/:id", habitH.Update)
			habits.DELETE("/:id", habitH.Archive)
			habits.POST("/:id/checkin", habitH.CheckIn)
			habits.DELETE("/:id/checkin", habitH.UndoCheckIn)
			habits.GET("/:id/calendar", habitH.GetCalendar)
		}

		// Todos
		todos := protected.Group("/todos")
		{
			todos.GET("", todoH.List)
			todos.POST("", todoH.Create)
			todos.PUT("/:id", todoH.Update)
			todos.DELETE("/:id", todoH.Delete)
			todos.PUT("/reorder", todoH.Reorder)
		}

		// Family
		family := protected.Group("/family")
		{
			family.POST("", familyH.Create)
			family.GET("", familyH.Get)
			family.POST("/join", familyH.Join)
			family.GET("/members", familyH.GetMembers)
			family.DELETE("/members/:id", familyH.RemoveMember)
			family.POST("/invite-code", familyH.RegenerateInviteCode)
		}

		// Projects
		projects := protected.Group("/projects")
		{
			projects.GET("", projectH.List)
			projects.POST("", projectH.Create)
			projects.GET("/:id", projectH.GetByID)
			projects.PUT("/:id", projectH.Update)
			projects.DELETE("/:id", projectH.Archive)
			projects.POST("/:id/phases", projectH.CreatePhase)
			projects.PUT("/:id/phases/:phaseId", projectH.UpdatePhase)
			projects.DELETE("/:id/phases/:phaseId", projectH.DeletePhase)
			projects.POST("/:id/tasks", projectH.CreateTask)
			projects.PUT("/:id/tasks/:taskId", projectH.UpdateTask)
			projects.DELETE("/:id/tasks/:taskId", projectH.DeleteTask)
			projects.PUT("/:id/tasks/:taskId/move", projectH.MoveTask)
			projects.PUT("/:id/tasks/reorder", projectH.ReorderTasks)
			projects.GET("/overdue-tasks", projectH.GetOverdueTasks)
		}

		// Dashboard & Activity
		protected.GET("/dashboard", dashH.GetDashboard)
		protected.GET("/activity", dashH.GetActivities)

		// Upload
		if uploadH != nil {
			protected.POST("/upload", uploadH.Upload)
		}

		// Export
		exportH := handler.NewExportHandler(db)
		export := protected.Group("/export")
		{
			export.GET("/goals", exportH.ExportGoals)
			export.GET("/habits", exportH.ExportHabits)
		}
	}
}
