package config

import "github.com/gin-gonic/gin"

type ControllerRunnable interface {
	RunController(r *gin.Engine)
}

type GinController struct {
	Controllers []ControllerRunnable
}

func (c *GinController) Start() {
	router := gin.Default()
	for _, o := range c.Controllers {
		o.RunController(router)
	}
	_ = router.Run()
}
