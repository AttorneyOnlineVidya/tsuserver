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
)

type BanList struct {
	UserList
}

var banlist_file string = "storage/banlist.json"
var ban_list *BanList = new(BanList)

// returns a pointer to the specific ban if client is banned
// also returns whether the user is IP banned or HDID banned or both
func (bl *BanList) isBanned(cl *Client) (*IPHDPair, bool, bool) {
	return bl.isInList(cl)
}

// adds ban to the list
func (bl *BanList) addBan(cl *Client, reason string) {
	bl.addUser(cl, reason)

	// write results to banlist file
	bl.writeBanlist()
}

func (bl *BanList) writeBanlist() {
	bl.lock.RLock()
	bl_json, _ := json.MarshalIndent(bl.Userlist, "", "  ")
	bl.lock.RUnlock()

	ioutil.WriteFile(banlist_file, bl_json, 0666)
}

func (bl *BanList) loadBanlist() error {
	bl.lock.Lock()
	defer bl.lock.Unlock()

	if bytes, err := ioutil.ReadFile(banlist_file); err == nil {
		if err2 := json.Unmarshal(bytes, &bl.Userlist); err2 != nil {
			return err2
		}
	} else {
		return err
	}

	return nil
}
