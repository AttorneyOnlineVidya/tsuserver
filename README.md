# tsuserver

This is a server to be used with the game Attorney Online.
It is meant to be an alternative to the many servers floating
around, hoping for better performance and more flexibility.

## Features

* Multiple areas per server
* Simple user management

## How to use

Compile and run.

## User Commands

* /area (number) - Changes user to that area, if blank will list current areas.

* /bg (background)- Changes the background to one in the backgrounds list.

* /charselect - Brings up the character select screen (Shrinks client)

* /g (message) - Sends a global message to everyone in the server
	* /global - Toggles Global off/on
	
* /help - Links to this readme.

* /motd - prints the MOTD to chat
	* If logged in as a mod you can change the MOTD. /motd (message)
	
* /need (message) - Sends an advert to everyone in the server
	* /adverts - Toggles adverts off/on
	
* /pm (target) - Sends a PM to the specified character or OOC name.
	* Character name only PMs the target in your current area.
	
* /pos (position) - Moves user to the specified position in court.
	* (wit, def, pro, jud, hlp, hld)
	
* /randomchar - Changes you to a random free character

* /roll (number) - Rolls a dice between 1 and 6 or 1 and number, min is 2, max is 9999

* /status (status) - If blank displays current area status.
	* Statuses are: idle, buildingopen, buildingfull, casingopen, casingfull, recess

* /switch (character folder) - Changes the user to the specified character.
	* Character must be in characters list.	

## Mod Commands

* /announce - Makes a server wide announcement

* /ban (target) - Disconnects the target and adds them to the banlist, must be IP

* /bglock - Toggles the background lock in the area, preventing users from using /bg

* /kick (target) - Disconnects the target from the server, can be IP, Character name or OOC name.
	* Character name only kicks the target in your current area.
	
* /login (password) - Logs client in as moderator.

* /mute (target) - Mutes the target, can be IP, Character name or OOC name.
	* Character name only mutes the target in your current area.
	
* /reloadbans - Reloads the banlist
	
* /unmute (target) - Unmutes the target, can be IP, Character name or OOC name.
	* Character name only mutes the target in your current area.


## License

This server is licensed under the GPLv3 license. See the
[LICENSE](LICENSE.md) file for more information.
