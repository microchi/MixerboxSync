package main

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/ahmetb/go-linq/v3"
	"github.com/bogem/id3v2"
	"github.com/eiannone/keyboard"
	"github.com/gookit/color"
	"github.com/gosuri/uiprogress"
	"github.com/kkdai/youtube/v2"
	flag "github.com/spf13/pflag"
)

var (
	ffmpeg      = ""
	runtimeGOOS = runtime.GOOS
)

const logo = `
___  ____               _                 _____                  
|  \/  (_)             | |               /  ___|                 
| .  . |___  _____ _ __| |__   _____  __ \  --. _   _ _ __   ___ 
| |\/| | \ \/ / _ \ '__| '_ \ / _ \ \/ /   --. \ | | | '_ \ / __|
| |  | | |>  <  __/ |  | |_) | (_) >  <  /\__/ / |_| | | | | (__ 
\_|  |_/_/_/\_\___|_|  |_.__/ \___/_/\_\ \____/ \__, |_| |_|\___|
                                                 __/ |           
                                                |___/            
`

type listItem struct {
	ID    string `json:"f"`
	Title string `json:"tt"`
}

type playList struct {
	Vector struct {
		Items []listItem `json:"items"`
	} `json:"getVector"`
}

type uiProgressWriter struct {
	ProgressBar *uiprogress.Bar
}

func (myWriter uiProgressWriter) Write(data []byte) (int, error) {
	myLen := len(data)

	myWriter.ProgressBar.Incr()
	_ = myWriter.ProgressBar.Set(myWriter.ProgressBar.Current() + myLen)

	return myLen, nil
}

func deletFileNotInList(myFiles *[]os.FileInfo, myList *playList, syncPath *string, isNoConfirm *bool) {
	linq.From(*myFiles).WhereT(func(myFile os.FileInfo) bool {
		return !linq.From(myList.Vector.Items).WhereT(func(myItem listItem) bool {
			return myItem.ID != ""
		}).AnyWithT(func(myItem listItem) bool {
			return hasID(*syncPath+myFile.Name(), myItem.ID)
		})
	}).ForEachT(func(myFile os.FileInfo) {
		if !*isNoConfirm {
			color.LightYellow.Print("\nDelete ", myFile.Name(), " [Y/n]")

			if err := keyboard.Open(); err != nil {
				color.LightRed.Println(err)

				return
			}
			defer keyboard.Close()

			char, _, err := keyboard.GetKey()
			if err != nil {
				color.LightRed.Println(err)

				return
			}

			color.LightYellow.Print(string(char))

			// not in y Y space or enter
			if char != 89 && char != 121 && char != 0 {
				return
			}
		}

		if err := os.Remove(*syncPath + myFile.Name()); err != nil {
			color.LightRed.Println("\n", err)
		} else {
			color.LightYellow.Println("\n", myFile.Name(), "Deleted!")
		}
	})
}

func convert(mp4FileName *string, mp3FileName *string, myID *string, myProgressBar *uiprogress.Bar) {
	cmd := exec.Command(ffmpeg, "-y", "-i", *mp4FileName, "-vn", *mp3FileName) //nolint:gosec

	stderr, _ := cmd.StderrPipe()
	_ = cmd.Start()

	scanner := bufio.NewScanner(stderr)
	scanner.Split(bufio.ScanWords)

	isDuration := false

	for scanner.Scan() {
		myText := scanner.Text()

		switch {
		case myText == "Duration:":
			isDuration = true
		case isDuration:
			isDuration = false

			myDurations := strings.Split(strings.TrimRight(myText, ","), ":")
			h, _ := strconv.ParseFloat(myDurations[0], 32)
			m, _ := strconv.ParseFloat(myDurations[1], 32)
			s, _ := strconv.ParseFloat(myDurations[2], 32)

			myProgressBar.Total = int((h * 60 * 60) + (m * 60) + s)
			_ = myProgressBar.Set(0)
			myProgressBar.Incr()
		case strings.HasPrefix(myText, "time="):
			myDurations := strings.Split(strings.TrimLeft(myText, "time="), ":")
			h, _ := strconv.ParseFloat(myDurations[0], 32)
			m, _ := strconv.ParseFloat(myDurations[1], 32)
			s, _ := strconv.ParseFloat(myDurations[2], 32)

			_ = myProgressBar.Set(int((h * 60 * 60) + (m * 60) + s))
			myProgressBar.Incr()
		}
	}

	_ = cmd.Wait()

	tag, _ := id3v2.Open(*mp3FileName, id3v2.Options{Parse: true})
	defer tag.Close()
	tag.AddTextFrame(tag.CommonID("Publisher"), tag.DefaultEncoding(), *myID)
	_ = tag.Save()
}

func uiprogressFactory(myContentLength int, myTitle string, isDownloadStep *bool) *uiprogress.Bar {
	return uiprogress.
		AddBar(myContentLength).
		AppendCompleted().PrependElapsed().
		AppendFunc(func(myBar *uiprogress.Bar) string { return color.LightCyan.Render(myTitle) }).
		PrependFunc(func(myBar *uiprogress.Bar) string {
			if color.IsLikeInCmd() {
				_, _ = color.Set(color.LightCyan)
			}

			if *isDownloadStep {
				return color.LightMagenta.Render("Downloading")
			}

			return color.LightMagenta.Render("Converting ")
		})
}

func download(myID string, myPath string, myClient youtube.Client, myWaitGroup *sync.WaitGroup) {
	defer myWaitGroup.Done()

	video, err := myClient.GetVideo(myID)
	if err != nil {
		color.LightRed.Println(err)

		return
	}

	myResponse, err := myClient.GetStream(video, &video.Formats[0])
	if err != nil {
		color.LightRed.Println(err)

		return
	}
	defer myResponse.Body.Close()

	video.Title = strings.TrimSpace(regexp.MustCompile(`[<>:"\/\|?*]`).ReplaceAllString(video.Title, ""))

	mp4FileName := myPath + video.Title + ".mp4"
	mp3FileName := myPath + video.Title + ".mp3"

	file, err := os.Create(mp4FileName)
	if err != nil {
		color.LightRed.Println(err)

		return
	}

	isDownloadStep := true

	myProgressBar := uiprogressFactory(int(myResponse.ContentLength), video.Title, &isDownloadStep)

	myReader := io.TeeReader(myResponse.Body, uiProgressWriter{ProgressBar: myProgressBar})

	if _, err = io.Copy(file, myReader); err != nil {
		color.LightRed.Println(err)

		return
	}

	file.Close()

	isDownloadStep = false

	convert(&mp4FileName, &mp3FileName, &myID, myProgressBar)

	if err = os.Remove(mp4FileName); err != nil {
		color.LightRed.Println(err)
	}
}

func getFiles(myPath string) *[]os.FileInfo {
	if _, err := os.Stat(myPath); os.IsNotExist(err) {
		_ = os.Mkdir(myPath, os.ModePerm)
	}

	files, err := ioutil.ReadDir(myPath)
	if err != nil {
		color.LightRed.Println(err)

		return &[]os.FileInfo{}
	}

	return &files
}

func getPlayList(myID string) *playList {
	myList := new(playList)

	myURL := "https://www.mixerbox.com/api/0/com.mixerbox.www/0/en/getVector?type=playlist&vectorId=" + myID

	myRequest, err := http.NewRequest(http.MethodGet, myURL, nil)
	myRequest = myRequest.WithContext(context.Background())

	if err != nil {
		color.LightRed.Println(err)

		return nil
	}

	myRequest.Header.Add("referer", "https://www.mixerbox.com/")

	myResponse, err := http.DefaultClient.Do(myRequest)
	if err != nil {
		color.LightRed.Println(err)

		return nil
	}
	defer myResponse.Body.Close()

	myBody, err := ioutil.ReadAll(myResponse.Body)
	if err != nil {
		color.LightRed.Println(err)

		return nil
	}

	err = json.Unmarshal(myBody, myList)

	if err != nil {
		color.LightRed.Println(err)

		return nil
	}

	return myList
}

func hasID(myFile string, myID string) bool {
	if strings.ToLower(filepath.Ext(myFile)) != ".mp3" {
		return false
	}

	myTag, err := id3v2.Open(myFile, id3v2.Options{Parse: true})
	if err != nil {
		return false
	}
	defer myTag.Close()

	return myTag.GetTextFrame(myTag.CommonID("Publisher")).Text == myID
}

func checkFFMpeg() bool {
	ffmpeg, _ = exec.LookPath("ffmpeg")

	if ffmpeg == "" {
		ffmpeg, _ = exec.LookPath("./ffmpeg")
	}

	if ffmpeg == "" {
		color.LightRed.Println("FFMpeg Not Found!\n")
		color.LightRed.Println("Please Visit https://ffmpeg.org/download.html For Install Instructions!\n")
		color.LightBlue.Println("Or Download Link As Below")

		switch runtimeGOOS {
		case "windows":
			color.LightBlue.Println(
				"Windows 64bit https://ffmpeg.zeranoe.com/builds/win64/static/ffmpeg-4.3.1-win64-static-lgpl.zip")
			color.LightBlue.Println(
				"Windows 32bit https://ffmpeg.zeranoe.com/builds/win32/static/ffmpeg-4.3.1-win32-static-lgpl.zip")
		case "darwin":
			color.LightBlue.Println(
				"MacOS 64bit https://ffmpeg.zeranoe.com/builds/macos64/static/ffmpeg-4.3.1-macos64-static-lgpl.zip")
		case "linux":
			color.LightBlue.Println("Linux amd64 https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-amd64-static.tar.xz")
			color.LightBlue.Println("Linux i686  https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-i686-static.tar.xz")
			color.LightBlue.Println("Linux arm64 https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-arm64-static.tar.xz")
			color.LightBlue.Println("Linux armhf https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-armhf-static.tar.xz")
			color.LightBlue.Println("Linux armel https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-armel-static.tar.xz")
		}
	}

	return ffmpeg != ""
}

func printUsage() {
	color.LightBlue.Println("Usage   : MixerboxSync ID [-p=PATH] [-s] [-y]")
	color.LightBlue.Println("Example : MixerboxSync 10086761 -sy")
	color.LightBlue.Println("Playlist ID Cloud Be Found in https://www.mixerbox.com/\n")
	flag.PrintDefaults()
}

func parseFlag() (playListID int, isHelp *bool, syncPath *string, isSync *bool, isNoConfirm *bool) {
	flag.Usage = printUsage

	isHelp = flag.BoolP("help", "h", false, "Show This Discription.")
	syncPath = flag.StringP("path", "p", "", "Path To Sync. Default is Playlist ID.")
	isSync = flag.BoolP("sync", "s", false, "Delete File Not In the List.")
	isNoConfirm = flag.BoolP("yes", "y", false, "Delete File Without Confirm.")

	flag.Parse()

	if len(flag.Args()) == 1 {
		playListID, _ = strconv.Atoi(flag.Arg(0))
	}

	if *syncPath == "" {
		*syncPath = strconv.Itoa(playListID)
	}

	if !strings.HasSuffix(*syncPath, string(os.PathSeparator)) {
		*syncPath += string(os.PathSeparator)
	}

	return
}

func main() {
	color.LightCyan.Println(logo)

	playListID, isHelp, syncPath, isSync, isNoConfirm := parseFlag()

	if playListID == 0 || *isHelp {
		printUsage()
		os.Exit(0)

		return
	}

	if !checkFFMpeg() {
		os.Exit(0)

		return
	}

	myList := getPlayList(strconv.Itoa(playListID))

	myFiles := getFiles(*syncPath)

	var myWaitGroup sync.WaitGroup

	uiprogress.Start()

	linq.From(myList.Vector.Items).WhereT(func(myItem listItem) bool {
		return myItem.ID != ""
	}).ForEachT(func(myItem listItem) {
		if linq.From(*myFiles).AnyWithT(func(myFile os.FileInfo) bool {
			return hasID(*syncPath+myFile.Name(), myItem.ID)
		}) {
			color.LightGreen.Print("  File Exist!  ")
			color.LightBlue.Println(myItem.Title)
		} else {
			color.LightRed.Print("Downloading... ")
			color.LightBlue.Println(myItem.Title)
			myWaitGroup.Add(1)
			go download(myItem.ID, *syncPath, youtube.Client{}, &myWaitGroup)
		}
	})

	myWaitGroup.Wait()

	if *isSync {
		deletFileNotInList(myFiles, myList, syncPath, isNoConfirm)
	}

	color.LightBlue.Println("\nSync Done !!!\n")
}
