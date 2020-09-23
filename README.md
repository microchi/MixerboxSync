    ___  ____               _                 _____                  
    |  \/  (_)             | |               /  ___|                 
    | .  . |___  _____ _ __| |__   _____  __ \  --. _   _ _ __   ___ 
    | |\/| | \ \/ / _ \ '__| '_ \ / _ \ \/ /   --. \ | | | '_ \ / __|
    | |  | | |>  <  __/ |  | |_) | (_) >  <  /\__/ / |_| | | | | (__ 
    \_|  |_/_/_/\_\___|_|  |_.__/ \___/_/\_\ \____/ \__, |_| |_|\___|
                                                    __/ |           
                                                    |___/            

Synchronize MixerBox PlayList Songs To Locale Folder
==================

[![LICENSE MIT](https://img.shields.io/github/license/microchi/MixboxSync)](https://raw.githubusercontent.com/microchi/MixboxSync/master/LICENSE)
[![BUILD](https://github.com/microchi/MixerboxSync/workflows/Go/badge.svg?branch=master)](https://github.com/microchi/MixerboxSync/actions)
[![codecov](https://codecov.io/gh/microchi/MixerboxSync/branch/master/graph/badge.svg)](https://codecov.io/gh/microchi/MixerboxSync)
[![Go Report Card](https://goreportcard.com/badge/github.com/microchi/MixerboxSync)](https://goreportcard.com/report/github.com/microchi/MixerboxSync)

>**[中文說明](README.zh.md)**

This is a command line Tool for synchronizing MixerBox playlist songs to locale folder.

## DOWNLOADS
The latest versions of MixerBoxSync executables are available for download at Release tab

64-bit builds are provided for Windows, Mac OS X, Linux.

To use this tool you will need FFmpeg on your system (see https://ffmpeg.org/download.html).


## USAGE
Visit https://www.mixerbox.com/ to get playlist ID from url.

EX: https://www.mixerbox.com/list/10086761 The ID is 10086761

Or open MixerBox app in you mobile device. use shere button to get playlist ID

From MixerBoxSync Folder Run Command: 
```shel
MixerboxSync 10086761 -sy
```

Flag s will delete file not in the playlist.

Flag y will delete file without confirm.

Flag p can specify folder to sync. EX: -p=yourfolder

Default folder is playlist ID

![Screen](https://microchi.github.io/MixerboxSync/screen.gif)

## BUILDING FROM SOURCE
To build MixerboxSync you will need Go v1.15 installed on your system (see http://golang.org/dl/).

From source folder run command:

```shel
go build
```

## Test Coverage
From source folder run command:
```shel
go test -gcflags=-l -v -cover
```

## LICENSE
This tool is licensed under MIT license. See LICENSE for details.