package controller

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/mhsanaei/3x-ui/v2/database/model"
	"github.com/mhsanaei/3x-ui/v2/pkg/x/xs"
	"github.com/mhsanaei/3x-ui/v2/web/session"
	"github.com/mhsanaei/3x-ui/v2/web/websocket"
)

type InboundCreateDTO struct {
	Up                   int64          `form:"up" json:"up"`
	Down                 int64          `form:"down" json:"down"`
	Total                int64          `form:"total" json:"total"`
	Remark               string         `form:"remark" json:"remark"`
	Enable               bool           `form:"enable" json:"enable"`
	ExpiryTime           int64          `form:"expiryTime" json:"expiryTime"`
	TrafficReset         string         `form:"trafficReset" json:"trafficReset"`
	LastTrafficResetTime int64          `form:"lastTrafficResetTime" json:"lastTrafficResetTime"`
	Listen               string         `form:"listen" json:"listen"`
	Port                 int            `form:"port" json:"port"`
	Protocol             model.Protocol `form:"protocol" json:"protocol"`

	Settings       string `form:"settings" json:"settings"`
	StreamSettings string `form:"streamSettings" json:"streamSettings"`
	Sniffing       string `form:"sniffing" json:"sniffing"`
}

// addInbound creates a new inbound configuration.
func (a *InboundController) addInbound(c *gin.Context) {
	var dto InboundCreateDTO

	// у тебя сейчас form-data, значит так:
	if err := c.ShouldBind(&dto); err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err) // не success
		return
	}

	// валидируем JSON-поля (иначе можно записать мусор в jsonb)
	if dto.Settings != "" && !json.Valid([]byte(dto.Settings)) {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), fmt.Errorf("invalid settings JSON"))
		return
	}
	if dto.StreamSettings != "" && !json.Valid([]byte(dto.StreamSettings)) {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), fmt.Errorf("invalid streamSettings JSON"))
		return
	}
	if dto.Sniffing != "" && !json.Valid([]byte(dto.Sniffing)) {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), fmt.Errorf("invalid sniffing JSON"))
		return
	}

	user := session.GetLoginUser(c)

	inbound := &model.Inbound{
		UserId:               user.Id,
		Up:                   dto.Up,
		Down:                 dto.Down,
		Total:                dto.Total,
		AllTime:              0,
		Remark:               dto.Remark,
		Enable:               dto.Enable,
		ExpiryTime:           dto.ExpiryTime,
		TrafficReset:         dto.TrafficReset,
		LastTrafficResetTime: dto.LastTrafficResetTime,
		ClientStats:          nil,
		Listen:               dto.Listen,
		Port:                 dto.Port,
		Protocol:             dto.Protocol,
		Tag:                  "",
	}

	inbound.SetSettingsString(dto.Settings)
	inbound.SetStreamSettingsString(dto.StreamSettings)
	inbound.SetSniffingString(dto.Sniffing)

	// tag
	if inbound.Listen == "" || inbound.Listen == "0.0.0.0" || inbound.Listen == "::" || inbound.Listen == "::0" {
		inbound.Tag = fmt.Sprintf("inbound-%v", inbound.Port)
	} else {
		inbound.Tag = fmt.Sprintf("inbound-%v:%v", inbound.Listen, inbound.Port)
	}

	created, needRestart, err := a.inboundService.AddInbound(inbound)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	}

	jsonMsgObj(c, I18nWeb(c, "pages.inbounds.toasts.inboundCreateSuccess"), toInboundDTOPtr(created), nil)
	if needRestart {
		a.xrayService.SetToNeedRestart()
	}

	inbounds, _ := a.inboundService.GetInbounds(user.Id)
	websocket.BroadcastInbounds(xs.Map(inbounds, func(item *model.Inbound, index int) InboundDTO {
		return toInboundDTOPtr(item)
	}))
}
