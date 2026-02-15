package controller

import (
	"github.com/mhsanaei/3x-ui/v2/database/model"
	"github.com/mhsanaei/3x-ui/v2/xray"
)

type InboundDTO struct {
	Id                   int                  `json:"id"`
	UserId               int                  `json:"-"` // как и в модели
	Up                   int64                `json:"up"`
	Down                 int64                `json:"down"`
	Total                int64                `json:"total"`
	AllTime              int64                `json:"allTime"`
	Remark               string               `json:"remark"`
	Enable               bool                 `json:"enable"`
	ExpiryTime           int64                `json:"expiryTime"`
	TrafficReset         string               `json:"trafficReset"`
	LastTrafficResetTime int64                `json:"lastTrafficResetTime"`
	ClientStats          []xray.ClientTraffic `json:"clientStats"`

	Listen   string         `json:"listen"`
	Port     int            `json:"port"`
	Protocol model.Protocol `json:"protocol"`

	// 👇 ключевое отличие
	Settings       string `json:"settings"`
	StreamSettings string `json:"streamSettings"`
	Sniffing       string `json:"sniffing"`

	Tag string `json:"tag"`
}

func toInboundDTO(in model.Inbound) InboundDTO {
	return InboundDTO{
		Id:                   in.Id,
		UserId:               in.UserId,
		Up:                   in.Up,
		Down:                 in.Down,
		Total:                in.Total,
		AllTime:              in.AllTime,
		Remark:               in.Remark,
		Enable:               in.Enable,
		ExpiryTime:           in.ExpiryTime,
		TrafficReset:         in.TrafficReset,
		LastTrafficResetTime: in.LastTrafficResetTime,
		ClientStats:          in.ClientStats,
		Listen:               in.Listen,
		Port:                 in.Port,
		Protocol:             in.Protocol,
		Tag:                  in.Tag,
		Settings:             string(in.Settings),
		StreamSettings:       string(in.StreamSettings),
		Sniffing:             string(in.Sniffing),
	}
}

func toInboundDTOPtr(in *model.Inbound) InboundDTO {
	if in == nil {
		return InboundDTO{}
	}
	return toInboundDTO(*in)
}
