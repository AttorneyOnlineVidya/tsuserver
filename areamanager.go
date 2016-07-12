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
	"errors"
	"fmt"
	"sync"
	"time"
)

type Area struct {
	Areaid        int
	Name          string
	Background    string
	bglock        bool
	status        string
	clients       []*Client
	lock          sync.Mutex
	hp_def        int
	hp_pro        int
	song_timer    *time.Timer
	taken_charids map[int]*Client
}

func (a *Area) sendRawMessage(msg string) {
	client_list.sendAllRawIf(msg, func(c *Client) bool {
		return c.area == a
	})
}

func (a *Area) sendServerMessageOOC(msg string) {
	a.sendRawMessage("CT#" + config.Reservedname + "#" + msg + "#%")
}

func (a *Area) getCharCount() int {
	a.lock.Lock()
	defer a.lock.Unlock()

	count := 0
	for _, c := range a.clients {
		if isValidCharID(c.charid) {
			count += 1
		}
	}

	return count
}

func (a *Area) addClient(c *Client) {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.clients = append(a.clients, c)
}

func (a *Area) removeClient(c *Client) {
	a.lock.Lock()
	defer a.lock.Unlock()

	for i, v := range a.clients {
		if c == v {
			a.clients = append(a.clients[:i], a.clients[i+1:]...)
			return
		}
	}
}

func (a *Area) initialize() {
	a.hp_def = 10
	a.hp_pro = 10
	a.taken_charids = make(map[int]*Client)
	a.status = "IDLE"
}

func (a *Area) setDefHP(hp int) error {
	if hp >= 0 && hp <= 10 {
		a.hp_def = hp
		return nil
	} else {
		return errors.New("Invalid HP value.")
	}
}

func (a *Area) setProHP(hp int) error {
	if hp >= 0 && hp <= 10 {
		a.hp_pro = hp
		return nil
	} else {
		return errors.New("Invalid HP value.")
	}
}

func (a *Area) playMusic(songname string, charid int, duration int) {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.sendRawMessage(fmt.Sprintf("MC#%s#%d#%%", songname, charid))

	if a.song_timer != nil {
		a.song_timer.Stop()
	}

	if duration == -1 {
		return
	}

	a.song_timer = time.AfterFunc(time.Second*time.Duration(duration), func() {
		a.playMusic(songname, charid, duration)
	})
}

func (a *Area) changeBackground(name string) error {
	// check if said background exists
	bg, err := stringInSlice(name, config.Backgroundlist, false)
	if err != nil {
		return errors.New("This background does not exist.")
	}

	a.lock.Lock()
	defer a.lock.Unlock()

	// change background
	a.Background = bg
	a.sendRawMessage("BN#" + bg + "#%")

	writeServerLog(fmt.Sprintf("Background in Area %d changed to %s.",
		a.Areaid, a.Background))

	return nil
}

func (a *Area) addTakenCharacter(id int, cl *Client) {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.taken_charids[id] = cl
}

func (a *Area) removeTakenCharacter(id int) {
	a.lock.Lock()
	defer a.lock.Unlock()

	delete(a.taken_charids, id)
}

func (a *Area) isCharIDAvailable(charid int) bool {
	a.lock.Lock()
	defer a.lock.Unlock()

	_, ok := a.taken_charids[charid]
	return !ok
}

func (a *Area) randomFreeCharacterID() (int, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	var avail_ids []int

	for i := range config.Charlist {
		if _, ok := a.taken_charids[i]; !ok {
			avail_ids = append(avail_ids, i)
		}
	}

	if len(avail_ids) == 0 {
		return 0, errors.New("No available IDs.")
	}

	randid := rng.Intn(len(avail_ids))
	return randid, nil
}

func (a *Area) getClientByCharName(charname string) *Client {
	for i := range config.Charlist {
		if config.Charlist[i] == charname {
			a.lock.Lock()
			defer a.lock.Unlock()

			if cl, ok := a.taken_charids[i]; ok {
				return cl
			}

			return nil
		}
	}

	return nil
}

func (a *Area) setAreaStatus(cl *Client, status string) {
	a.status = status
	a.sendServerMessageOOC(cl.getCharacterName() + " changed the area status to " + status)
	writeClientLog(cl, "changed the area status to "+status)
}
