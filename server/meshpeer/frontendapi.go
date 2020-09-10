package meshpeer

// FrontendUserDataType holds final frontend-defined data
type FrontendUserDataType interface{}

// FrontendUserData is FrontendUserDataType + update timestamp
type FrontendUserData struct {
	TS   NetworkTime
	Data FrontendUserDataType
}

// FrontEndUpdateObject describes the entire network state known to this peer
type FrontEndUpdateObject struct {
	ThisPeer FrontendUserData
	AllPeers map[NetworkID]FrontendUserData
}

// FrontendAPI allows a peer network code to interact with higher-layer frontend code
type FrontendAPI interface {
	RegisterUserDataUpdateHandler(func(FrontendUserDataType))
	HandleUpdate(FrontEndUpdateObject)
}
