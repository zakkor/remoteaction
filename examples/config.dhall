let RemoteAction = https://raw.githubusercontent.com/zakkor/remoteaction/b0c8ce37cf005522518c3c5144828f67edd0e328/remoteaction.dhall sha256:201c3b2f74b7f7a90cb3beb2f328aed5b62414c67a46acd587b46a481a9cad46
let url = RemoteAction.url
let regex = RemoteAction.regex

let ytdl-script = \(args: { path : Text, format : Text }) ->
''
cd ${args.path}
youtube-dl $YTDL_ARGS --ignore-errors --write-info-json --geo-bypass --no-playlist -f ${args.format} "$LINK"
''

let ytdl-audio = ytdl-script { path = "$HOME/Downloads/music", format = "bestaudio" }
let ytdl-video = ytdl-script { path = "$HOME/Downloads/videos", format = "best" }
let github-clone = ''
cd $HOME/tmp/github;
git clone --recursive "$LINK"
''
let dl-torrent = "echo download torrent @ $LINK"

in RemoteAction.Config::{
    executors = 5,
    menus = [
        RemoteAction.Link {
          patterns = [url "*://www.youtube.com/watch*", url "*://youtu.be/*"],
          actions = [
            { name = "Download YouTube audio", script = ytdl-audio },
            { name = "Download YouTube video", script = ytdl-video }
          ]
        },

        RemoteAction.Link {
          patterns = [url "*://www.youtube.com/channel/*", url "*://www.youtube.com/user/*"],
          actions = [
            { name = "Download YouTube Channel as audio", script = ytdl-audio },
            { name = "Download YouTubes Channel videos", script = ytdl-video }
          ]
        },

        RemoteAction.Link {
          patterns = [url "*://*.bandcamp.com/*/*"],
          actions = [ { name = "Download BandCamp track(s)", script = ytdl-audio } ]
        },

        RemoteAction.Link {
          patterns = [url "*://soundcloud.com/*"],
          actions = [ { name = "Download SoundCloud track(s)", script = ytdl-audio } ]
        },

        RemoteAction.Link {
          patterns = [url "https://www.reddit.com/r/*/comments/*/*"],
          actions = [ { name = "Download Reddit video", script = ytdl-video } ]
        },

        RemoteAction.Link {
          patterns = [url "https://github.com/*/*"],
          actions = [ { name = "Clone GitHub repo", script = github-clone } ]
        },

        RemoteAction.Selection {
          regexes = ["[0-9A-F]{40}"],
	  actions = [ { name = "Download torrent hash", script = dl-torrent } ]
        }
    ]
}
