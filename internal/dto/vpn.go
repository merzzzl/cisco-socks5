package dto

const (
	VPNStateConnected    VPNState = "Connected"
	VPNStateDisconnected VPNState = "Disconnected"
	VPNStateUnknown      VPNState = "Unknown"

	VPNNoticeReadyForConnect VPNNotice = "ReadyForConnect"
	VPNNoticeUnknown         VPNNotice = "Unknown"
)

type VPNState string
type VPNNotice string
