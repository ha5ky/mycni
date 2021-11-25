/**
 * @Author: zhangc
 * @Date: 2021/11/22 17:01
 * @LastEditors: zhangc
 * @FilePath: /main.go
 * @Description:
 * @Contactme: zhangchun34582@hundsun.com
**/

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/containernetworking/cni/pkg/skel"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/containernetworking/plugins/pkg/utils/buildversion"
	"github.com/vishvananda/netlink"
	"syscall"
)

type NetConf struct {
	Bridge string `json:"bridge"`
	IP     string `json:"ip"`
}

func loadConf(conf []byte) (*NetConf, error) {
	n := &NetConf{}
	if err := json.Unmarshal(conf, n); err != nil {
		return nil, err
	}
	return n, nil
}

func setupBridge(n *NetConf) (*netlink.Bridge, error) {
	// initial bridge obj
	br := &netlink.Bridge{
		LinkAttrs: netlink.LinkAttrs{
			Name:   n.Bridge,
			MTU:    1500,
			TxQLen: -1,
		},
	}
	var err error

	// create a bridge device
	if err = netlink.LinkAdd(br); err != nil && err != syscall.EEXIST {
		return br, err
	}

	// check if bridge created successfully
	br, err = bridgeByName(n.Bridge)
	if err != nil {
		return nil, err
	}

	// setup bridge
	err = netlink.LinkSetUp(br)
	if err != nil {
		return nil, err
	}
	return br, nil
}

func bridgeByName(name string) (*netlink.Bridge, error) {
	l, err := netlink.LinkByName(name)
	if err != nil {
		return nil, err
	}
	br, ok := l.(*netlink.Bridge)
	if !ok {
		return nil, errors.New(fmt.Sprintf("%s is not a bridge dev", name))
	}
	return br, nil
}

func setupVeth(netNS ns.NetNS, br *netlink.Bridge, ifName string, hwAddr string, ipAddr string) error {
	hostIface := &current.Interface{}

	err := netNS.Do(func(hostNS ns.NetNS) error {
		hostVeth, containerVeth, err := ip.SetupVeth(ifName, 1500, hwAddr, hostNS)
		if err != nil {
			return err
		}
		hostIface.Name = hostVeth.Name
		return nil
	})
	if err != nil {
		return err
	}
	hIf, err := netlink.LinkByName(hostIface.Name)
	if err != nil {
		return err
	}

	err = netlink.LinkSetMaster(hIf, br)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, version.All, buildversion.BuildString("my-bridge"))
}

func cmdAdd(args *skel.CmdArgs) error {

}

func cmdCheck(args *skel.CmdArgs) error {
	return nil
}

func cmdDel(args *skel.CmdArgs) error {
	return nil
}
