package controllers

import (
	"cisco-socks5/api"
	"cisco-socks5/internal/dto"
	"cisco-socks5/internal/services/fw"
	"cisco-socks5/internal/services/tunnel"
	"cisco-socks5/internal/services/vpn"
	cl "cisco-socks5/pkg/controlloop"
	"cisco-socks5/pkg/log"
	"context"
	"time"
)

func NewMainReconcile(
	conditionsChan chan []cl.Condition,
	vpnService *vpn.Service,
	fwService *fw.Service,
	tunnelService *tunnel.Service,

) *MainReconcile {
	return &MainReconcile{
		conditionsChan: conditionsChan,
		vpnService:     vpnService,
		fwService:      fwService,
		tunnelService:  tunnelService,
	}
}

type MainReconcile struct {
	conditionsChan chan []cl.Condition
	vpnService     *vpn.Service
	fwService      *fw.Service
	tunnelService  *tunnel.Service
}

func (r *MainReconcile) Reconcile(_ context.Context, object *api.MainConfig) (cl.Result, error) {
	config := object

	defer func() {
		r.conditionsChan <- config.GetConditions()
	}()

	if config.GetKillTimestamp() != "" {
		return r.reconcileKill(config)
	}

	return r.reconcileNormal(config)
}

func (r *MainReconcile) reconcileNormal(mc *api.MainConfig) (cl.Result, error) {
	vpnState, _, err := r.vpnService.GetState()
	if err != nil {
		mc.MarkFalse(api.VPNConnectedCondition, api.VPNConnectionStateFailedReason, err.Error())
		return cl.Result{}, err
	}
	if vpnState != dto.VPNStateConnected {
		log.Info().Msg("Main", "Connecting to VPN...")
		err = r.vpnService.Connect()
		if err != nil {
			mc.MarkFalse(api.VPNConnectedCondition, api.VPNConnectionFailedReason, err.Error())
			return cl.Result{}, err
		}
		log.Info().Msg("Main", "Connecting to VPN success!")
		return cl.Result{RequeueAfter: time.Second}, nil
	}
	mc.MarkTrue(api.VPNConnectedCondition)
	err = r.fwService.Disable()
	if err != nil {
		mc.MarkFalse(api.PFDisabledCondition, api.PFDisabledFailedReason, err.Error())
		return cl.Result{}, err
	}
	mc.MarkTrue(api.PFDisabledCondition)

	err = r.tunnelService.StartTunnel()
	if err != nil {
		log.Info().Err(err)
		mc.MarkFalse(api.TunnelEnabledCondition, api.TunnelInitializationFailedReason, err.Error())
		return cl.Result{}, err
	}
	mc.MarkTrue(api.TunnelEnabledCondition)
	return cl.Result{RequeueAfter: time.Second * 20}, nil
}

func (r *MainReconcile) reconcileKill(mc *api.MainConfig) (cl.Result, error) {
	log.Info().Msg("Main", "Reconcile Kill")

	pid, ok, err := r.tunnelService.GetTunnelPID()
	if err != nil {
		mc.MarkFalse(api.TunnelEnabledCondition, api.TunnelDisablingFailedReason, err.Error())
		return cl.Result{}, err
	}
	log.Info().Msg("Main", "Reconcile Kill PID", pid)

	if ok {
		err = r.tunnelService.StopTunnel(pid)
		if err != nil {
			mc.MarkFalse(api.TunnelEnabledCondition, api.TunnelDisablingFailedReason, err.Error())
			return cl.Result{}, err
		}
		log.Info().Msg("Main", "Reconcile Kill STOP TUNNEL", pid)
	}

	vpnState, _, err := r.vpnService.GetState()
	if err != nil {
		return cl.Result{}, err
	}

	if vpnState == dto.VPNStateConnected {
		err = r.vpnService.Disconnect()
		if err != nil {
			mc.MarkFalse(api.VPNConnectedCondition, api.VPNDisconnectFailedReason, err.Error())
			return cl.Result{}, err
		}
		log.Info().Msg("Main", "Reconcile Kill STOP VPN")
	}

	log.Info().Msg("Main", "Reconcile Kill STOPPED")

	return cl.Result{}, nil
}
