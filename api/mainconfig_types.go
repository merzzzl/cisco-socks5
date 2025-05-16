package api

import (
	cl "cisco-socks5/pkg/controlloop"
)

const (
	VPNConnectedCondition     = "VPNConnected"
	PFDisabledCondition       = "PFDisabled"
	TunnelEnabledCondition    = "TunnelEnabled"

	VPNReasonReady                   = "VPNDisconnected"
	VPNConnectionFailedReason        = "VPNConnectionFailed"
	VPNConnectionStateFailedReason   = "VPNConnectionStateFailed"
	VPNDisconnectFailedReason        = "VPNDisconnectFailed"
	VPNConnectionStopFiled           = "VPNStopFailed"
	PFDisabledFailedReason           = "PFDisabledFailed"
	TunnelInitializationFailedReason = "TunnelInitializationFailed"
	TunnelDisablingFailedReason      = "TunnelDisablingFailed"
)

type MainConfig struct {
	cl.Resource
	Spec MainConfigSpec
}

type MainConfigSpec struct{}

// todo generate
func (c *MainConfig) DeepCopy() *MainConfig {
	return cl.DeepCopyStruct(c).(*MainConfig)
}
