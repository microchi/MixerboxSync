Synchronize MixerBox PlayList Songs To Locale Folder
==================

This is a command line Tool for synchronizing MixerBox playlist songs to locale folder.

## DOWNLOADS
---
The latest versions of MixerBoxSync executables are available for download at Release tab

64-bit builds are provided for Windows, Mac OS X, Linux.

To use this tool you will need FFmpeg on your system (see https://ffmpeg.org/download.html).


## USAGE
---
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

## BUILDING FROM SOURCE
---
To build MixerboxSync you will need Go v1.15 installed on your system (see http://golang.org/dl/).

From source folder run command:

```shel
go build
```

## Test Coverage
---
From source folder run command:
```shel
go test -gcflags=-l -v -cover
```

## LICENSE
---
This tool is licensed under MIT license. See LICENSE for details.