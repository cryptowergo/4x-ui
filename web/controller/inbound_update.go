package controller

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/mhsanaei/3x-ui/v2/database/model"
)

type InboundUpdateDTO struct {
	Up                   int64          `json:"up" form:"up"`
	Down                 int64          `json:"down" form:"down"`
	Total                int64          `json:"total" form:"total"`
	Remark               string         `json:"remark" form:"remark"`
	Enable               bool           `json:"enable" form:"enable"`
	ExpiryTime           int64          `json:"expiryTime" form:"expiryTime"`
	TrafficReset         string         `json:"trafficReset" form:"trafficReset"`
	LastTrafficResetTime int64          `json:"lastTrafficResetTime" form:"lastTrafficResetTime"`
	Listen               string         `json:"listen" form:"listen"`
	Port                 int            `json:"port" form:"port"`
	Protocol             model.Protocol `json:"protocol" form:"protocol"`

	// ВАЖНО: в DTO это строки, потому что form-data присылает текст
	Settings       string `json:"settings" form:"settings"`
	StreamSettings string `json:"streamSettings" form:"streamSettings"`
	Sniffing       string `json:"sniffing" form:"sniffing"`
}

func bindInboundUpdate(c *gin.Context, id int) (*model.Inbound, error) {
	var dto InboundUpdateDTO

	ct := c.GetHeader("Content-Type")
	if strings.HasPrefix(ct, "application/json") {
		if err := c.ShouldBindJSON(&dto); err != nil {
			return nil, err
		}
	} else {
		if err := c.ShouldBind(&dto); err != nil {
			return nil, err
		}
	}

	// Валидация JSON строк (чтобы не записать мусор в jsonb)
	if dto.Settings != "" && !json.Valid([]byte(dto.Settings)) {
		return nil, fmt.Errorf("invalid settings JSON")
	}
	if dto.StreamSettings != "" && !json.Valid([]byte(dto.StreamSettings)) {
		return nil, fmt.Errorf("invalid streamSettings JSON")
	}
	if dto.Sniffing != "" && !json.Valid([]byte(dto.Sniffing)) {
		return nil, fmt.Errorf("invalid sniffing JSON")
	}

	inb := &model.Inbound{
		Id:                   id,
		Up:                   dto.Up,
		Down:                 dto.Down,
		Total:                dto.Total,
		Remark:               dto.Remark,
		Enable:               dto.Enable,
		ExpiryTime:           dto.ExpiryTime,
		TrafficReset:         dto.TrafficReset,
		LastTrafficResetTime: dto.LastTrafficResetTime,
		Listen:               dto.Listen,
		Port:                 dto.Port,
		Protocol:             dto.Protocol,
	}

	inb.SetSettingsString(dto.Settings)
	inb.SetStreamSettingsString(dto.StreamSettings)
	inb.SetSniffingString(dto.Sniffing)

	return inb, nil
}

// updateInbound updates an existing inbound configuration.
func (a *InboundController) updateInbound(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	}

	inbound, err := bindInboundUpdate(c, id)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	}

	updated, needRestart, err := a.inboundService.UpdateInbound(inbound)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	}

	jsonMsgObj(c, I18nWeb(c, "pages.inbounds.toasts.inboundUpdateSuccess"), toInboundDTOPtr(updated), nil)

	if needRestart {
		a.xrayService.SetToNeedRestart()
	}
}
