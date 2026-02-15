package controller

import (
	"github.com/gin-gonic/gin"

	"github.com/mhsanaei/3x-ui/v2/web/session"
)

// getInbounds retrieves the list of inbounds for the logged-in user.
func (a *InboundController) getInbounds(c *gin.Context) {
	user := session.GetLoginUser(c)

	inbounds, err := a.inboundService.GetInbounds(user.Id)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.obtain"), err)
		return
	}

	resp := make([]InboundDTO, 0, len(inbounds))
	for _, inb := range inbounds {
		resp = append(resp, toInboundDTO(*inb))
	}

	jsonObj(c, resp, nil)
}
