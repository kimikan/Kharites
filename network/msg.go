package network

const (
	flagData    = 0xfdfd
	flagControl = 0xfefe

	optionNone     = 0x0000
	optionZip      = 0x0001
	optionCheckCrc = 0x0002
	optionEncrypt  = 0x0010

	encryptKey = 0x7f
)

const (
	//GameServerPort ...
	GameServerPort = 7740
	//TCPHeadFlag ...
	TCPHeadFlag uint16 = 0xFDFD
	//PacketVersion ...
	PacketVersion = 0

	//PacketTypeKeepAlive ...
	PacketTypeKeepAlive = 0x1
	//PacketTypeKeepAliveAck ...
	PacketTypeKeepAliveAck = 0x80000001

	//PacketTypeLogin ...
	PacketTypeLogin = 0x4
	//PacketTypeLoginAck ...
	PacketTypeLoginAck = 0x80000004

	//PacketTypeLogout ...
	PacketTypeLogout = 0x8
	//PacketTypeLogoutAck ...
	PacketTypeLogoutAck = 0x80000008

	//PacketTypeReadDisk ...
	PacketTypeReadDisk = 0x5
	//PacketTypeReadDiskAck ...
	PacketTypeReadDiskAck = 0x80000005
)

//TCPHeader ...
type TCPHeader struct {
	Flag   uint16
	Option uint16

	Size uint32
	Crc  uint32
}

//PacketHeader ...
type PacketHeader struct {
	ID      uint32 //00，  包流水号
	Type    uint32 //04，  包类型
	Len     uint32 //08，  大小			//	ntohl(所有长度)
	Version uint32 //0C，  协议版本号
	Result  uint32 //10，  返回值 0=成功，其他是FR（错误编号）
	Token   uint32 //14，  令牌
}

//LoginMsg ...
type LoginMsg struct {
	Header     *PacketHeader
	Reason     uint32 //登录原因()
	DiskID     uint32 //磁盘Id
	SnapshotID uint32 //还原点Id，重连时使用
}

//LoginAckMsg ...
type LoginAckMsg struct {
	Header     *PacketHeader
	DiskID     uint32 //磁盘Id
	SnapshotID uint32 //还原点Id
	Flags      uint8  //第0位1表示是整个磁盘， 0表示是分区(游戏虚拟盘)
	//第1位1表示全新会话，   0表示复用会话
	//第7位1表示支持UDP分包，0表示不支持
	SectorCount uint64 //扇区数(64b)
}

//LogoutMsg ...
type LogoutMsg struct {
	Header *PacketHeader
}

//KeepAliveMsg ...
type KeepAliveMsg struct {
	Header *PacketHeader
}

//ReadDiskMsg ...
type ReadDiskMsg struct {
	Header       *PacketHeader
	DiskID       uint32
	SectorCount  uint16
	SectorOffset uint64
	Data         []byte
}
