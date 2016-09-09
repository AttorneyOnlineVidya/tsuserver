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
	"log"
	"strconv"
	"strings"

	"github.com/burntsushi/toml"
)

var server_version string = "tsuserver0.1"

type Song struct {
	Name     string
	Duration int
}

type Evidence struct {
	Type  string
	Name  string
	Desc  string
	Image string
}

type Config struct {
	Port           int
	Slots          int
	Charlist       []string
	Musiclist      []Song
	Arealist       []Area
	Evidencelist   []Evidence
	Backgroundlist []string
	Defaultarea    int
	Timeout        int
	Reservedname   string
	Modpass        string
	Guardpass      string
	Advertise      bool
	Masterserver   string
	Servername     string
	Description    string
	MOTD           string
}

var config Config

func loadConfig() {
	// load configs
	if _, err := toml.DecodeFile("./config/config.toml", &config); err != nil {
		log.Fatal(err)
	}

	if _, err := toml.DecodeFile("./config/characters.toml", &config); err != nil {
		log.Fatal(err)
	}

	if _, err := toml.DecodeFile("./config/areas.toml", &config); err != nil {
		log.Fatal(err)
	}

	if _, err := toml.DecodeFile("./config/evidence.toml", &config); err != nil {
		log.Fatal(err)
	}

	if _, err := toml.DecodeFile("./config/backgrounds.toml", &config); err != nil {
		log.Fatal(err)
	}

	var tmpconf struct{ Musiclist []string }

	if _, err := toml.DecodeFile("./config/musiclist.toml", &tmpconf); err != nil {
		log.Fatal(err)
	}

	for _, v := range tmpconf.Musiclist {
		spl := strings.Split(v, "*")
		name := spl[0]
		dur := -1
		if len(spl) == 2 {
			tmpdur, _ := strconv.Atoi(spl[1])
			dur = tmpdur
		}

		s := Song{Name: name, Duration: dur}
		config.Musiclist = append(config.Musiclist, s)
	}

	// set defaults
	for i := range config.Arealist {
		config.Arealist[i].initialize()
	}
}

func reloadMusicConfig() {
	var tmpconf struct{ Musiclist []string }

	if _, err := toml.DecodeFile("./config/musiclist.toml", &tmpconf); err != nil {
		log.Fatal(err)
	}
	config.Musiclist = nil
	for _, v := range tmpconf.Musiclist {
		spl := strings.Split(v, "*")
		name := spl[0]
		dur := -1
		if len(spl) == 2 {
			tmpdur, _ := strconv.Atoi(spl[1])
			dur = tmpdur
		}

		s := Song{Name: name, Duration: dur}
		config.Musiclist = append(config.Musiclist, s)
	}
}

func reloadCharList() {
	config.Charlist = nil

	if _, err := toml.DecodeFile("./config/characters.toml", &config); err != nil {
		log.Fatal(err)
	}
}

func reloadBackgroundlist() {
	config.Backgroundlist = nil

	if _, err := toml.DecodeFile("./config/backgrounds.toml", &config); err != nil {
		log.Fatal(err)
	}
}

func reloadConfig() {
	config.Slots = -1
	config.Timeout = -1
	config.Reservedname = ""
	config.Modpass = ""
	config.Guardpass = ""
	config.Advertise = false
	config.Masterserver = ""
	config.Servername = ""
	config.Description = ""
	config.MOTD = ""
	if _, err := toml.DecodeFile("./config/config.toml", &config); err != nil {
		log.Fatal(err)
	}
}

func reloadEvidence() {
	config.Evidencelist = nil

	if _, err := toml.DecodeFile("./config/evidence.toml", &config); err != nil {
		log.Fatal(err)
	}
}
