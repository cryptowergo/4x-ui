package controller

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func (a *InboundController) getInbound(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	}

	inbound, err := a.inboundService.GetInbound(id)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.obtain"), err)
		return
	}

	// inbound у тебя скорее всего *model.Inbound
	dto := toInboundDTO(*inbound)

	jsonObj(c, dto, nil)
}
