/*
tsuserver, an Attorney Online server
Copyright (C) 2016 tsukasa84 <tsukasadev84@gmail.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"encoding/json"
	"io/ioutil"
	"sync"
)

var banlist_file string = "storage/banlist.json"
var ban_list *BanList = new(BanList)

type Ban struct {
	Iplist []string
	Hdlist []string
}

type BanList struct {
	Banlist []Ban
	lock    sync.Mutex
}

// returns a pointer to the specific ban if client is banned
// also returns whether the user is IP banned or HDID banned or both
func (bl *BanList) isBanned(cl *Client) (*Ban, bool, bool) {
	bl.lock.Lock()
	defer bl.lock.Unlock()

	ipbanned := false
	hdbanned := false

	for i := range bl.Banlist {
		// check IPs
		for _, ip := range bl.Banlist[i].Iplist {
			if cl.IP.String() == ip {
				ipbanned = true
				break
			}
		}

		// check HDIDs
		for _, hd := range bl.Banlist[i].Hdlist {
			if cl.HDID == hd {
				hdbanned = true
				break
			}
		}

		if ipbanned || hdbanned {
			return &bl.Banlist[i], ipbanned, hdbanned
		}
	}

	return nil, false, false
}

// adds ban to the list
func (bl *BanList) addBan(cl *Client) {
	var ban *Ban

	// check if such a ban already exists
	if b, ipb, hdb := bl.isBanned(cl); b != nil {
		ban = b

		bl.lock.Lock()
		if !ipb {
			ban.Iplist = append(ban.Iplist, cl.IP.String())
		}
		if !hdb {
			ban.Hdlist = append(ban.Hdlist, cl.HDID)
		}
		bl.lock.Unlock()
	} else {
		// add the IP and HDID
		ban = &Ban{}
		ban.Iplist = append(ban.Iplist, cl.IP.String())
		ban.Hdlist = append(ban.Hdlist, cl.HDID)

		// add to banlist
		bl.lock.Lock()
		bl.Banlist = append(bl.Banlist, *ban)
		bl.lock.Unlock()
	}

	// write results to banlist file
	bl.writeBanlist()
}

func (bl *BanList) writeBanlist() {
	bl.lock.Lock()
	bl_json, _ := json.Marshal(bl.Banlist)
	bl.lock.Unlock()

	ioutil.WriteFile(banlist_file, bl_json, 0666)
}

func (bl *BanList) loadBanlist() error {
	if bytes, err := ioutil.ReadFile(banlist_file); err == nil {
		if err2 := json.Unmarshal(bytes, &bl.Banlist); err2 != nil {
			return err2
		}
	} else {
		return err
	}

	return nil
}
