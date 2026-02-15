package controller

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/mhsanaei/3x-ui/v2/database/model"
)

type UpdateClientDTO struct {
	Id       int    `form:"id" json:"id"`
	Settings string `form:"settings" json:"settings"`
}

// updateInboundClient updates a client's configuration in an inbound.
func (a *InboundController) updateInboundClient(c *gin.Context) {
	clientId := c.Param("clientId")

	var dto UpdateClientDTO
	if err := c.ShouldBind(&dto); err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err) // не success
		return
	}

	if dto.Id <= 0 {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), fmt.Errorf("invalid id"))
		return
	}
	if dto.Settings == "" || !json.Valid([]byte(dto.Settings)) {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), fmt.Errorf("invalid settings JSON"))
		return
	}

	inbound := &model.Inbound{Id: dto.Id}
	inbound.SetSettingsString(dto.Settings)

	needRestart, err := a.inboundService.UpdateInboundClient(inbound, clientId)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	}

	jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.inboundClientUpdateSuccess"), nil)
	if needRestart {
		a.xrayService.SetToNeedRestart()
	}
}
