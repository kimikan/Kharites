package util

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
)

/*
<game>
    <addrs>
        <addr>xx</addr>
    </addrs>

    <disks>
        <disk>
            <key></key>
            <value></value>
        </disk>
    </disks>
	<cache></cache>
</game>
*/

//Disk ...
type Disk struct {
	Key   uint32 `xml:"key"`
	Value string `xml:"value"`
}

//Addr ...
type Addr struct {
	URL string `xml:",chardata"`
	//ServerName string `xml:"serverName"`
}

//Config ...
type Config struct {
	Addrs     []Addr `xml:"addrs>addr"`
	Disks     []Disk `xml:"disks>disk"`
	CacheSize int    `xml:"cache"`
}

var (
	config *Config
)

//InitConfig ...
func InitConfig(file string) bool {
	body, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println(err)
		return false
	}
	config = new(Config)
	err = xml.Unmarshal(body, config)
	fmt.Println(err, config)
	return err == nil
}

//GetConfig ...
func GetConfig() *Config {
	return config
}
