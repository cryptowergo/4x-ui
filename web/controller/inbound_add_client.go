package controller

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/mhsanaei/3x-ui/v2/database/model"
)

type AddClientDTO struct {
	Id       int    `form:"id" json:"id"`
	Settings string `form:"settings" json:"settings"`
}

// addInboundClient adds a new client to an existing inbound.
func (a *InboundController) addInboundClient(c *gin.Context) {
	var dto AddClientDTO
	if err := c.ShouldBind(&dto); err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err) // не inboundUpdateSuccess
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

	data := &model.Inbound{Id: dto.Id}
	data.SetSettingsString(dto.Settings)

	needRestart, err := a.inboundService.AddInboundClient(data)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	}
	jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.inboundClientAddSuccess"), nil)

	if needRestart {
		a.xrayService.SetToNeedRestart()
	}
}
