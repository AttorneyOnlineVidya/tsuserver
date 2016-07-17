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

// A self-updating user list.

package main

import "sync"

type IPHDPair struct {
	Iplist []string `json:"IP"`
	Hdlist []string `json:"HD"`
	Data   string   `json:"data"`
}

type UserList struct {
	Userlist []IPHDPair `json:"Users"`
	lock     sync.RWMutex
}

// returns a pointer to the specific IPHDPair if client is in the list
// also returns whether the user's IP or HD are present
func (ul *UserList) isInList(cl *Client) (*IPHDPair, bool, bool) {
	ul.lock.RLock()
	defer ul.lock.RUnlock()

	ip_present := false
	hd_present := false

	for i := range ul.Userlist {
		// check IPs
		for _, ip := range ul.Userlist[i].Iplist {
			if cl.IP.String() == ip {
				ip_present = true
				break
			}
		}

		// check HDIDs
		for _, hd := range ul.Userlist[i].Hdlist {
			if cl.HDID == hd {
				hd_present = true
				break
			}
		}

		if ip_present || hd_present {
			return &ul.Userlist[i], ip_present, hd_present
		}
	}

	return nil, false, false
}

// adds user to the list
func (ul *UserList) addUser(cl *Client, data string) {
	var iphd *IPHDPair

	// check if such a user already exists
	if p, ipp, hdp := ul.isInList(cl); p != nil {
		iphd = p

		ul.lock.Lock()
		if len(data) > 0 {
			p.Data = data
		}
		if !ipp {
			iphd.Iplist = append(iphd.Iplist, cl.IP.String())
		}
		if !hdp {
			iphd.Hdlist = append(iphd.Hdlist, cl.HDID)
		}
		ul.lock.Unlock()
	} else {
		// add the IP and HDID
		iphd = &IPHDPair{}
		iphd.Iplist = append(iphd.Iplist, cl.IP.String())
		iphd.Hdlist = append(iphd.Hdlist, cl.HDID)
		iphd.Data = data

		// add to userlist
		ul.lock.Lock()
		ul.Userlist = append(ul.Userlist, *iphd)
		ul.lock.Unlock()
	}
}
