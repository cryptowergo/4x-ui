package job

import (
	"encoding/json"

	"github.com/mhsanaei/3x-ui/v2/database/model"
	"github.com/mhsanaei/3x-ui/v2/logger"
	"github.com/mhsanaei/3x-ui/v2/web/service"
	"github.com/mhsanaei/3x-ui/v2/web/websocket"
	"github.com/mhsanaei/3x-ui/v2/xray"

	"github.com/valyala/fasthttp"
)

type InboundDTO struct {
	Id                   int                  `json:"id"`
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

	Listen   string `json:"listen"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Tag      string `json:"tag"`

	// важно: как строки (как в твоём /panel/api/inbounds/list)
	Settings       string `json:"settings"`
	StreamSettings string `json:"streamSettings"`
	Sniffing       string `json:"sniffing"`
}

func toInboundDTO(m model.Inbound) InboundDTO {
	return InboundDTO{
		Id:                   m.Id,
		Up:                   m.Up,
		Down:                 m.Down,
		Total:                m.Total,
		AllTime:              m.AllTime,
		Remark:               m.Remark,
		Enable:               m.Enable,
		ExpiryTime:           m.ExpiryTime,
		TrafficReset:         m.TrafficReset,
		LastTrafficResetTime: m.LastTrafficResetTime,
		ClientStats:          m.ClientStats,

		Listen:   m.Listen,
		Port:     m.Port,
		Protocol: string(m.Protocol),
		Tag:      m.Tag,

		Settings:       string(m.Settings),
		StreamSettings: string(m.StreamSettings),
		Sniffing:       string(m.Sniffing),
	}
}

func toInboundDTOPtr(in *model.Inbound) InboundDTO {
	if in == nil {
		return InboundDTO{}
	}
	return toInboundDTO(*in)
}

// XrayTrafficJob collects and processes traffic statistics from Xray, updating the database and optionally informing external APIs.
type XrayTrafficJob struct {
	settingService  service.SettingService
	xrayService     service.XrayService
	inboundService  service.InboundService
	outboundService service.OutboundService
}

// NewXrayTrafficJob creates a new traffic collection job instance.
func NewXrayTrafficJob() *XrayTrafficJob {
	return new(XrayTrafficJob)
}

// Run collects traffic statistics from Xray and updates the database, triggering restart if needed.
func (j *XrayTrafficJob) Run() {
	if !j.xrayService.IsXrayRunning() {
		return
	}
	traffics, clientTraffics, err := j.xrayService.GetXrayTraffic()
	if err != nil {
		return
	}
	err, needRestart0 := j.inboundService.AddTraffic(traffics, clientTraffics)
	if err != nil {
		logger.Warning("add inbound traffic failed:", err)
	}
	err, needRestart1 := j.outboundService.AddTraffic(traffics, clientTraffics)
	if err != nil {
		logger.Warning("add outbound traffic failed:", err)
	}
	if ExternalTrafficInformEnable, err := j.settingService.GetExternalTrafficInformEnable(); ExternalTrafficInformEnable {
		j.informTrafficToExternalAPI(traffics, clientTraffics)
	} else if err != nil {
		logger.Warning("get ExternalTrafficInformEnable failed:", err)
	}
	if needRestart0 || needRestart1 {
		j.xrayService.SetToNeedRestart()
	}

	// Get online clients and last online map for real-time status updates
	onlineClients := j.inboundService.GetOnlineClients()
	lastOnlineMap, err := j.inboundService.GetClientsLastOnline()
	if err != nil {
		logger.Warning("get clients last online failed:", err)
		lastOnlineMap = make(map[string]int64)
	}

	// Fetch updated inbounds from database with accumulated traffic values
	// This ensures frontend receives the actual total traffic, not just delta values
	updatedInbounds, err := j.inboundService.GetAllInbounds()
	if err != nil {
		logger.Warning("get all inbounds for websocket failed:", err)
	}

	updatedOutbounds, err := j.outboundService.GetOutboundsTraffic()
	if err != nil {
		logger.Warning("get all outbounds for websocket failed:", err)
	}

	// Broadcast traffic update via WebSocket with accumulated values from database
	trafficUpdate := map[string]any{
		"traffics":       traffics,
		"clientTraffics": clientTraffics,
		"onlineClients":  onlineClients,
		"lastOnlineMap":  lastOnlineMap,
	}
	websocket.BroadcastTraffic(trafficUpdate)

	// Broadcast full inbounds update for real-time UI refresh
	if updatedInbounds != nil {
		dtos := make([]InboundDTO, 0, len(updatedInbounds))
		for _, in := range updatedInbounds {
			dtos = append(dtos, toInboundDTOPtr(in))
		}

		websocket.BroadcastInbounds(updatedInbounds)
	}

	if updatedOutbounds != nil {
		websocket.BroadcastOutbounds(updatedOutbounds)
	}
}

func (j *XrayTrafficJob) informTrafficToExternalAPI(inboundTraffics []*xray.Traffic, clientTraffics []*xray.ClientTraffic) {
	informURL, err := j.settingService.GetExternalTrafficInformURI()
	if err != nil {
		logger.Warning("get ExternalTrafficInformURI failed:", err)
		return
	}
	requestBody, err := json.Marshal(map[string]any{"clientTraffics": clientTraffics, "inboundTraffics": inboundTraffics})
	if err != nil {
		logger.Warning("parse client/inbound traffic failed:", err)
		return
	}
	request := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(request)
	request.Header.SetMethod("POST")
	request.Header.SetContentType("application/json; charset=UTF-8")
	request.SetBody([]byte(requestBody))
	request.SetRequestURI(informURL)
	response := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(response)
	if err := fasthttp.Do(request, response); err != nil {
		logger.Warning("POST ExternalTrafficInformURI failed:", err)
	}
}
