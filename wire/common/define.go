package common

import "time"

// AlgorithmHashSlots is the algorithm in routing
const (
	AlgorithmHashSlots = "hashslots"
)

// Command defined data type between client and server
const (
	CommandLoginSignIn  = "login.signin"
	CommandLoginSignOut = "login.signout"

	CommandChatUserTalk  = "chat.user.talk"
	CommandChatGroupTalk = "chat.group.talk"
	CommandChatTalkAck   = "chat.talk.ack"

	CommandOfflineIndex   = "chat.offline.index"
	CommandOfflineContent = "chat.offline.content"

	CommandGroupCreate  = "chat.group.create"
	CommandGroupJoin    = "chat.group.join"
	CommandGroupQuit    = "chat.group.quit"
	CommandGroupMembers = "chat.group.members" //下面的获取detail已经包括了members
	CommandGroupDetail  = "chat.group.detail"
)

// Meta Key of a packet
const (
	// MetaDestServer 由ServerDispatcher注入消息包: 消息将要送达的网关的serviceName
	MetaDestServer = "dest.logic"
	// MetaDestChannels 消息将要送达的channels，即一条消息可推送给多个用户
	MetaDestChannels = "dest.channels"
)

// Protocol Protocol
type Protocol string

// Protocol
const (
	ProtocolTCP       Protocol = "tcp"
	ProtocolWebsocket Protocol = "websocket"
)

// Service Name 定义统一的服务名
const (
	SNWGateway = "wgateway"
	SNTGateway = "tgateway"
	SNLogin    = "login"  //login
	SNChat     = "chat"   //chat
	SNService  = "occult" //rpc service
	SNRouter   = "router" //router
)

type ServiceID string

type SessionID string

type Magic [4]byte

var (
	MagicLogicPkt = Magic{0xc3, 0x11, 0xa3, 0x65}
	MagicBasicPkt = Magic{0xc3, 0x15, 0xa7, 0x65}
)

const (
	OfflineReadIndexExpiresIn = time.Hour * 24 * 30 // 读索引在缓存中的过期时间
	OfflineSyncIndexCount     = 2000                //单次同步消息索引的数量
	OfflineMessageExpiresIn   = 15                  // 离线消息过期时间
	MessageMaxCountPerPage    = 200                 // 同步消息内容时每页的最大数据
)

// MessageType 消息类型，暂未支持
const (
	MessageTypeText  = 1
	MessageTypeImage = 2
	MessageTypeVoice = 3
	MessageTypeVideo = 4
)
