package main

/*
 * Written by kimi kan, 2016-10
 * due to this demo server involved the syscall.  so it can only be run @linux
 * it needs freecache https://github.com/coocood/freecache
 * & https://github.com/spaolacci/murmur3
 * This file is used for marshal & unmarshal the package of the request & response.
 */

import (
	"Kharites/ioservice"
	"Kharites/network"
	"Kharites/util"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

func initConfig() {

	path, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	if !util.InitConfig(path + "/config.xml") {
		log.Fatal(util.GetConfig(), path)
	}
}

//test 179.3
func main() {
	mgr := ioservice.NewFIOManager()
	defer mgr.Close()
	initConfig()

	funt := func(conn net.Conn) bool {
		r := network.NewReader(conn, conn.RemoteAddr())
		w := network.NewWriter(conn)
		fmt.Println(r, "Connected")

		for true {
			msg := r.ReadMsg()
			if msg == nil {
				fmt.Println(r, "Connect break!")
				break
			}
			switch realMsg := msg.(type) {
			case *network.KeepAliveMsg:
				if realMsg.Header.Type == network.PacketTypeKeepAlive {
					realMsg.Header.Type = network.PacketTypeKeepAliveAck
					if !w.WriteMsg(realMsg) {
						log.Println("Keepalive ACK send failed!")
					}
				}
				//should be ack only
			case *network.ReadDiskMsg:
				if realMsg.Header.Type == network.PacketTypeReadDisk {
					mgr.AddIORequest(realMsg.DiskID, realMsg, w)
				}

			case *network.LoginMsg:
				msg2 := new(network.LoginAckMsg)
				msg2.Header = realMsg.Header
				msg2.Header.Type = network.PacketTypeLoginAck
				msg2.Header.Len = 0x29
				msg2.DiskID = realMsg.DiskID
				msg2.SnapshotID = 1
				msg2.Flags = 0
				msg2.SectorCount = util.GetSectorCount(realMsg.DiskID)
				if !w.WriteMsg(msg2) {
					fmt.Println("Error writing..LoginAckMsg.")
				}
			case *network.LogoutMsg:
				if realMsg.Header.Type == network.PacketTypeLogout {
					realMsg.Header.Type = network.PacketTypeLogoutAck
					if !w.WriteMsg(realMsg) {
						fmt.Println("Error writing...")
					}
				}

			default:
				fmt.Println("Strange, should be here, Nil")
			}
		}
		return true
	}

	for _, addr := range util.GetConfig().Addrs {
		tcp := network.NewTCPServer(addr.URL, funt)
		tcp.Start()
		tcp.Run()
		defer tcp.Stop()
	}

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGUSR2)
	signal.Notify(c, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		fmt.Println("get signal:", s)
		break
	}
}
