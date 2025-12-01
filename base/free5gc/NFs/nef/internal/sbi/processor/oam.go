package processor

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (p *Processor) GetOamIndex(c *gin.Context) {
	c.JSON(http.StatusOK, nil)
}
