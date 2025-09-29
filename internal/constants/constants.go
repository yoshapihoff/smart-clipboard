package constants

// SyncPort is the fixed UDP port for clipboard synchronization
const SyncPort = 9999

// SyncBroadcastPort is the UDP port used for server discovery broadcasts
const SyncBroadcastPort = 9998

// SyncMagicHeader is the magic header used to identify sync packets
const SyncMagicHeader = "SMART_CLIPBOARD_SYNC"

// DiscoveryMagicHeader is the magic header used for server discovery
const DiscoveryMagicHeader = "SMART_CLIPBOARD_DISCOVER"
