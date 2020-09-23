    ___  ____               _                 _____                  
    |  \/  (_)             | |               /  ___|                 
    | .  . |___  _____ _ __| |__   _____  __ \  --. _   _ _ __   ___ 
    | |\/| | \ \/ / _ \ '__| '_ \ / _ \ \/ /   --. \ | | | '_ \ / __|
    | |  | | |>  <  __/ |  | |_) | (_) >  <  /\__/ / |_| | | | | (__ 
    \_|  |_/_/_/\_\___|_|  |_.__/ \___/_/\_\ \____/ \__, |_| |_|\___|
                                                    __/ |           
                                                    |___/   
同步 MixerBox 歌單上的歌曲到本地資料夾
==================

[![LICENSE MIT](https://img.shields.io/github/license/microchi/MixboxSync)](https://raw.githubusercontent.com/microchi/MixboxSync/master/LICENSE)
[![BUILD](https://github.com/microchi/MixerboxSync/workflows/Go/badge.svg?branch=master)](https://github.com/microchi/MixerboxSync/actions)
[![codecov](https://codecov.io/gh/microchi/MixerboxSync/branch/master/graph/badge.svg)](https://codecov.io/gh/microchi/MixerboxSync)
[![Go Report Card](https://goreportcard.com/badge/github.com/microchi/MixerboxSync)](https://goreportcard.com/report/github.com/microchi/MixerboxSync)

這是個命令列工具用來同步 MixerBox 歌單上的歌曲到本地資料夾

## 下載
最新版的 MixerBoxSync 執行檔可在 Release 分頁下載

有64位元 Windows, Mac OS X, Linux 版

要使用本工具 你需要有 FFmpeg 在你的系統上 (參考 https://ffmpeg.org/download.html).

## 使用
到 https://www.mixerbox.com/ 由網址取得 ID

例如: https://www.mixerbox.com/list/10086761 ID 就是 10086761

或 在移動裝置開啟 MixerBox APP. 點選分享歌單也可取得 ID

在 MixerBoxSync 的資料夾 執行: 
```shel
MixerboxSync 10086761 -sy
```

標籤 s 會刪除不在歌單上的檔案

標籤 y 會刪除檔案時不再確認

標籤 p 可指定同步資料夾 例如: -p=yourfolder

預設資料夾為歌單ID

![Screen](https://microchi.github.io/MixerboxSync/screen.gif)

## 從原始碼編譯
要編譯 MixerboxSync 你需要在你的系統上安裝 GO v1.15 (參考 http://golang.org/dl/)

在 原始碼 的資料夾  執行: 

```shel
go build
```

## 測試 覆蓋率
在 原始碼 的資料夾  執行:
```shel
go test -gcflags=-l -v -cover
```

## 授權
本工具為 MIT 授權 細節參考 LICENSE