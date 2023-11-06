# better-battlebit-api

This is a very early proof of concept thus far, I plan to incorporate more cool and helpful features/endpoints in the very near future, including building a website ontop of it as well as a discord bot. If you have any suggestions, feel free to open an issue or pull request.

## Open Letter to Oki

This took me a total of 3 hours to make, most of that was spent figuring out your weird ass json.

You are a great game developer. Battlebit is incredibly fun, the most fun I've had playing an fps in a long time. However, your API is, in the nicest way possible, pretty terrible. I'm sure you know this and most of your efforts likely lie within game development, therefore I IMPLORE you to hire a backend api engineer or at least contract one to make a fully fledged api for the data as it stands. Even using this code is an immense speedup, however under the LICENSE you are not allowed to use this code in the current public api unless you go full open source, as you must state changes and disclose source. I'm sorry, but I have to protect my work and closed source web code is cringe anyway. If you wish to use this code or work with me in any way, feel free to contact me at `yoitscore [at] gmail [dot] com`.


## How does it work?

Every minute, the application makes a call to the Public API, then stores that in MongoDB. The api is built ontop of that. The speedup is significant. The ID you are given is actually a MongoDB Id, this is used as a state primarily for pagination, the data is cleared after 24 hours.