package controlPanel

import (
	"base/auth"

	"github.com/gin-gonic/gin"
)

func SetupControlRoutes(r *gin.Engine, Control *Control) {
	ControlGroup := r.Group("/control", auth.LoginRequired)

	ControlGroup.GET("/", Control.GetModels)
	ControlGroup.GET("/:model", Control.GetModelInstances)
	ControlGroup.DELETE("/:model/delete", Control.DeleteModelInstances)
	ControlGroup.GET("/:model/:id", Control.GetEditModelInstance)
	ControlGroup.PUT("/:model/:id", Control.GetEditModelInstance) // Edit instance
	ControlGroup.DELETE("/:model/:id", Control.DeleteModelInstance)
	ControlGroup.GET("/:model/q/:query", Control.QueryModel)
	ControlGroup.PUT("/:model/q/:query", Control.QueryModel)
	ControlGroup.DELETE("/:model/q/:query", Control.QueryModel)
}
