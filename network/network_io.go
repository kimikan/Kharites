package network

/*
 * Written by kimi kan, 2016-10
 * This file is used for marshal & unmarshal the package of the request & response.
 */

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
)

//NetReader ...
type NetReader struct {
	reader io.Reader

	addr net.Addr
}

//NewReader ...
func NewReader(r io.Reader, addr net.Addr) *NetReader {
	return &NetReader{reader: r, addr: addr}
}

func doRead(r io.Reader, order binary.ByteOrder, args ...interface{}) error {
	var err error

	for _, arg := range args {
		err = binary.Read(r, order, arg)
		if err != nil {
			break
		}
	}

	return err
}

//big endian
func (reader *NetReader) doReadTCPHeader() *TCPHeader {
	var header TCPHeader
	err := doRead(reader.reader, binary.BigEndian, (*uint16)(&header.Flag),
		&header.Option, &header.Size, &header.Crc)
	if err == nil {
		return &header
	}
	return nil
}

//big endian
func (reader *NetReader) doReadPktHeader() *PacketHeader {
	var header PacketHeader
	err := doRead(reader.reader, binary.LittleEndian, &header.ID,
		&header.Type, &header.Len, &header.Version, &header.Result, &header.Token)

	if err == nil {
		return &header
	}
	return nil
}

//ReadMsg ...
func (reader *NetReader) ReadMsg() interface{} {
	h1 := reader.doReadTCPHeader()
	if h1 != nil && h1.Flag == TCPHeadFlag {
		h2 := reader.doReadPktHeader()
		if h2 != nil {
			switch h2.Type {
			case PacketTypeKeepAlive:
				fallthrough
			case PacketTypeKeepAliveAck:
				msg := new(KeepAliveMsg)
				msg.Header = h2
				return msg
			case PacketTypeReadDisk:
				msg := new(ReadDiskMsg)
				msg.Header = h2
				err := doRead(reader.reader, binary.LittleEndian, &msg.DiskID,
					&msg.SectorCount, &msg.SectorOffset)
				if err == nil {
					return msg
				}

			case PacketTypeLogin:
				msg := new(LoginMsg)
				msg.Header = h2
				err := doRead(reader.reader, binary.LittleEndian, &msg.Reason,
					&msg.DiskID, &msg.SnapshotID)
				if err == nil {
					return msg
				}

			case PacketTypeLogout:
				msg := &LogoutMsg{Header: h2}
				return msg
			default:
				fmt.Println("Unknown message recieved: ", h2)
			}
		} else {
			fmt.Println("Pkt header null")
		}
	} else {
		fmt.Println(reader, "not normal message recieved: ", h1)
	}

	fmt.Println(reader, "Err packet: ", h1, reader.addr.Network())
	return nil
}

//NetWriter ...
type NetWriter struct {
	writer io.Writer
	buf    []byte
	len    int
}

const (
	bufCapacity = 128 * 1024
)

//Write implement io.writer
func (writer *NetWriter) Write(p []byte) (n int, err error) {
	remain := bufCapacity - writer.len
	l := len(p)
	if remain > l {
		copy(writer.buf[writer.len:], p)
		writer.len += l

		return l, nil
	}

	log.Fatal("No enough buf.")
	return 0, nil
}

func (writer *NetWriter) reset() {
	writer.len = 0
}

//NewWriter ...
func NewWriter(w io.Writer) *NetWriter {
	return &NetWriter{writer: w, buf: make([]byte, bufCapacity), len: 0}
}

func (writer *NetWriter) doWrite(order binary.ByteOrder, args ...interface{}) error {
	var err error

	for _, arg := range args {
		err = binary.Write(writer, order, arg)
		if err != nil {
			break
		}
	}
	return err
}

func (writer *NetWriter) doWriteTCPHeader(header *TCPHeader) bool {

	err := writer.doWrite(binary.BigEndian, header.Flag,
		header.Option, header.Size, header.Crc)
	return err == nil
}

//big endian
func (writer *NetWriter) doWritePktHeader(header *PacketHeader) bool {

	err := writer.doWrite(binary.LittleEndian, header.ID,
		header.Type, header.Len, header.Version, header.Result, header.Version)
	return err == nil
}

func newTCPHeader(size uint32) *TCPHeader {
	h := &TCPHeader{
		Flag:   TCPHeadFlag,
		Option: 0,
		Size:   size,
		Crc:    size,
	}
	return h
}

//WriteMsg ...
func (writer *NetWriter) WriteMsg(obj interface{}) bool {

	result := false
	writer.reset()
	switch msg := obj.(type) {
	case *KeepAliveMsg:
		if !writer.doWriteTCPHeader(newTCPHeader(msg.Header.Len)) {
			break
		}
		if writer.doWritePktHeader(msg.Header) {
			result = true
		}
		//should be ack only
	case *ReadDiskMsg:
		if !writer.doWriteTCPHeader(newTCPHeader(msg.Header.Len)) {
			break
		}
		if !writer.doWritePktHeader(msg.Header) {
			break
		}

		err := writer.doWrite(binary.LittleEndian, msg.DiskID,
			msg.SectorCount, msg.SectorOffset)
		if err != nil {
			break
		}
		_, err2 := writer.Write(msg.Data)
		if err2 == nil {
			result = true
		}
	case *LogoutMsg:
		if !writer.doWriteTCPHeader(newTCPHeader(msg.Header.Len)) {
			break
		}
		if !writer.doWritePktHeader(msg.Header) {
			break
		}
		result = true
	case *LoginAckMsg:
		//fmt.Println("Ack write: ", msg)
		if !writer.doWriteTCPHeader(newTCPHeader(msg.Header.Len)) {
			fmt.Println("11")
			break
		}
		if !writer.doWritePktHeader(msg.Header) {
			fmt.Println("22")
			break
		}
		err := writer.doWrite(binary.LittleEndian, msg.DiskID,
			msg.SnapshotID, msg.Flags, msg.SectorCount)
		if err != nil {
			break
		}
		result = true
	default:
		fmt.Println("unknown package type: ", msg)
	}

	if result {
		writer.writer.Write(writer.buf[0:writer.len])
	}
	return result
}
