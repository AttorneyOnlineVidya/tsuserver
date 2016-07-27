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
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"sync"
	"time"
)

var polls_file string = "storage/polls.json"
var poll_list *PollList = new(PollList)

type PollList struct {
	Polls map[string]*Poll
	lock  sync.RWMutex
}

type Poll struct {
	UserList
	Name        string    `json:"Name"`
	Description string    `json:"Description"`
	TimeStarted time.Time `json:"time_started"`
	Closed      bool
}

func (pl *PollList) getPollResults(name string) (string, error) {
	pl.lock.RLock()
	defer pl.lock.RUnlock()

	p, ok := pl.Polls[name]
	if !ok {
		return "", errors.New("Poll not found.")
	}

	yes := 0
	no := 0
	yes_percent := 0

	for _, v := range p.Userlist {
		if v.Data == "yes" {
			yes++
		} else if v.Data == "no" {
			no++
		}
	}

	total := yes + no

	if total > 0 {
		yes_percent = int(math.Floor((float64(yes)/float64(total))*100 + .5))
	}

	ret := fmt.Sprintf("Poll: %s. A total of %d votes have been registered. Yes: %d, No: %d -> %d<percent>.",
		p.Name, total, yes, no, yes_percent)

	fmt.Println(ret)

	return ret, nil
}

func (pl *PollList) getPollDescription(name string) (string, error) {
	pl.lock.RLock()
	defer pl.lock.RUnlock()

	p, ok := pl.Polls[name]
	if !ok {
		return "", errors.New("Poll not found.")
	}

	ret := "Poll title: " + p.Name + "\r\nPoll Description: " + p.Description

	return ret, nil

}

func (pl *PollList) newPoll(name string, description string) error {
	pl.lock.Lock()

	if _, ok := pl.Polls[name]; ok {
		pl.lock.Unlock()
		return errors.New("A poll with this name already exists.")
	}

	p := Poll{}
	p.Name = name
	p.Description = description
	p.Closed = false
	p.TimeStarted = time.Now()

	pl.Polls[name] = &p
	pl.lock.Unlock()

	pl.writePolls()

	return nil
}

func (pl *PollList) vote(cl *Client, name string, yes bool) error {
	pl.lock.Lock()

	p, ok := pl.Polls[name]

	if !ok {
		pl.lock.Unlock()
		return errors.New("Poll not found.")
	}

	if p.Closed {
		pl.lock.Unlock()
		return errors.New("This poll is closed.")
	}

	val := "yes"
	if !yes {
		val = "no"
	}

	p.addUser(cl, val)
	pl.lock.Unlock()

	pl.writePolls()

	return nil
}

func (pl *PollList) writePolls() {
	pl.lock.RLock()
	pl_json, _ := json.MarshalIndent(pl.Polls, "", "  ")
	pl.lock.RUnlock()

	ioutil.WriteFile(polls_file, pl_json, 0666)
}

func (pl *PollList) loadPolls() error {
	pl.lock.Lock()
	defer pl.lock.Unlock()

	var tmp map[string]*Poll

	if bytes, err := ioutil.ReadFile(polls_file); err == nil {
		if err2 := json.Unmarshal(bytes, &tmp); err2 != nil {
			return err2
		} else {
			pl.Polls = tmp
		}
	} else {
		return err
	}

	return nil
}

func (pl *PollList) closePoll(name string) error {
	pl.lock.Lock()
	p, ok := pl.Polls[name]

	if !ok {
		pl.lock.Unlock()
		return errors.New("Poll not found.")
	}

	if p.Closed {
		pl.lock.Unlock()
		return errors.New("This poll is already closed.")
	} else {
		p.Closed = true
	}

	pl.lock.Unlock()

	pl.writePolls()
	return nil
}

func (pl *PollList) getPollList() []string {
	pl.lock.RLock()
	defer pl.lock.RUnlock()

	var ret []string

	for _, v := range pl.Polls {
		if !v.Closed {
			ret = append(ret, v.Name)
		}
	}

	return ret
}
