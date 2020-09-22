package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	. "github.com/agiledragon/gomonkey"
	"github.com/bogem/id3v2"
	"github.com/eiannone/keyboard"
	"github.com/gookit/color"
	"github.com/gosuri/uiprogress"
	"github.com/kkdai/youtube/v2"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/pflag"
	flag "github.com/spf13/pflag"
)

type mockFile struct{}

func (myFile mockFile) Name() string {
	return "name"
}
func (myFile mockFile) Size() int64        { return 0 }
func (myFile mockFile) Mode() os.FileMode  { return 0 }
func (myFile mockFile) ModTime() time.Time { return time.Now() }
func (myFile mockFile) IsDir() bool        { return false }
func (myFile mockFile) Sys() interface{}   { return nil }

func TestProgressBar(t *testing.T) {
	Convey("uiProgressWriter", t, func() {
		Convey("Write_ShouldSetToBar", func() {
			myProgressBar := uiprogress.AddBar(10)
			myProgressWriter := uiProgressWriter{ProgressBar: myProgressBar}
			_, _ = myProgressWriter.Write([]byte("123"))
			So(myProgressBar.Current(), ShouldEqual, 4)
		})
	})
}

func TestDeletFileNotInList(t *testing.T) {
	Convey("DeletFileNotInList", t, func() {
		Convey("OSRemoveError_ShouldPrint", func() {
			syncPath := ""
			isNoConfirm := true
			myList := playList{}
			myList.Vector.Items = []listItem{
				{ID: ""},
				{ID: "1"},
			}

			myPatches := NewPatches()
			defer myPatches.Reset()

			myPatches.ApplyFunc(os.Remove, func(string) error { return errors.New("OSRemoveError") })

			var myBuffer bytes.Buffer
			color.SetOutput(io.Writer(&myBuffer))
			defer color.SetOutput(os.Stdout)

			deletFileNotInList(&[]os.FileInfo{&mockFile{}}, &myList, &syncPath, &isNoConfirm)

			So(myBuffer.String(), ShouldContainSubstring, "OSRemoveError")
		})

		Convey("OSRemoved_ShouldPrint", func() {
			syncPath := ""
			isNoConfirm := true
			myList := playList{}
			myList.Vector.Items = []listItem{
				{ID: ""},
				{ID: "1"},
			}

			myPatches := NewPatches()
			defer myPatches.Reset()

			myPatches.ApplyFunc(os.Remove, func(string) error { return nil })

			var myBuffer bytes.Buffer
			color.SetOutput(io.Writer(&myBuffer))
			defer color.SetOutput(os.Stdout)

			deletFileNotInList(&[]os.FileInfo{&mockFile{}}, &myList, &syncPath, &isNoConfirm)

			So(myBuffer.String(), ShouldContainSubstring, "Deleted!")
		})

		Convey("KeyboardOpenFail_ShouldPrint", func() {
			syncPath := ""
			isNoConfirm := false
			myList := playList{}
			myList.Vector.Items = []listItem{
				{ID: ""},
				{ID: "1"},
			}

			myPatches := NewPatches()
			defer myPatches.Reset()

			myPatches.ApplyFunc(os.Remove, func(string) error { return nil })
			myPatches.ApplyFunc(keyboard.Open, func() error { return errors.New("KeyboardOpenFail") })

			var myBuffer bytes.Buffer
			color.SetOutput(io.Writer(&myBuffer))
			defer color.SetOutput(os.Stdout)

			deletFileNotInList(&[]os.FileInfo{&mockFile{}}, &myList, &syncPath, &isNoConfirm)

			So(myBuffer.String(), ShouldContainSubstring, "KeyboardOpenFail")
		})

		Convey("KeyboardGetKeyFail_ShouldPrint", func() {
			syncPath := ""
			isNoConfirm := false
			myList := playList{}
			myList.Vector.Items = []listItem{
				{ID: ""},
				{ID: "1"},
			}

			myPatches := NewPatches()
			defer myPatches.Reset()

			myPatches.ApplyFunc(os.Remove, func(string) error { return nil })
			myPatches.ApplyFunc(keyboard.Open, func() error { return nil })
			myPatches.ApplyFunc(keyboard.GetKey, func() (rune, keyboard.Key, error) { return 0, 0, errors.New("KeyboardGetKeyFail") })

			var myBuffer bytes.Buffer
			color.SetOutput(io.Writer(&myBuffer))
			defer color.SetOutput(os.Stdout)

			deletFileNotInList(&[]os.FileInfo{&mockFile{}}, &myList, &syncPath, &isNoConfirm)

			So(myBuffer.String(), ShouldContainSubstring, "KeyboardGetKeyFail")
		})

		Convey("KeyboardGetKey1_ShouldNotRemove", func() {
			actualRemove := false
			syncPath := ""
			isNoConfirm := false
			myList := playList{}
			myList.Vector.Items = []listItem{
				{ID: ""},
				{ID: "1"},
			}

			myPatches := NewPatches()
			defer myPatches.Reset()

			myPatches.ApplyFunc(os.Remove, func(string) error {
				actualRemove = true
				return nil
			})
			myPatches.ApplyFunc(keyboard.Open, func() error { return nil })
			myPatches.ApplyFunc(keyboard.GetKey, func() (rune, keyboard.Key, error) { return 65, 0, nil })

			var myBuffer bytes.Buffer
			color.SetOutput(io.Writer(&myBuffer))
			defer color.SetOutput(os.Stdout)

			deletFileNotInList(&[]os.FileInfo{&mockFile{}}, &myList, &syncPath, &isNoConfirm)

			So(actualRemove, ShouldBeFalse)
		})
	})
}

func TestConvert(t *testing.T) {
	Convey("Convert", t, func() {
		Convey("Converted_ShouldAddIDToTag", func() {
			mp4FileName := ""
			mp3FileName := ""
			myID := "123"
			myCmd := &exec.Cmd{}
			myTag := id3v2.NewEmptyTag()

			myPatches := NewPatches()
			defer myPatches.Reset()

			myPatches.ApplyFunc(exec.Command, func(name string, arg ...string) *exec.Cmd { return myCmd })
			myPatches.ApplyMethod(reflect.TypeOf(myCmd), "StderrPipe", func(*exec.Cmd) (io.ReadCloser, error) {
				return ioutil.NopCloser(strings.NewReader("")), nil
			})
			myPatches.ApplyMethod(reflect.TypeOf(myCmd), "Start", func(*exec.Cmd) error { return nil })
			myPatches.ApplyFunc(id3v2.Open, func(name string, opts id3v2.Options) (*id3v2.Tag, error) { return myTag, nil })
			myPatches.ApplyMethod(reflect.TypeOf(myTag), "AddTextFrame", func(tag *id3v2.Tag, id string, encoding id3v2.Encoding, text string) {
				So(id, ShouldEqual, "TPUB")
				So(text, ShouldEqual, "123")
			})
			myPatches.ApplyMethod(reflect.TypeOf(myTag), "Save", func(*id3v2.Tag) error { return nil })

			convert(&mp4FileName, &mp3FileName, &myID, nil)
		})

		Convey("Progressing_ShouldSetToBar", func() {
			mp4FileName := ""
			mp3FileName := ""
			myID := ""
			myCmd := &exec.Cmd{}
			myTag := id3v2.NewEmptyTag()
			myProgressBar := uiprogress.AddBar(10)

			myPatches := NewPatches()
			defer myPatches.Reset()

			myPatches.ApplyFunc(exec.Command, func(name string, arg ...string) *exec.Cmd { return myCmd })
			myPatches.ApplyMethod(reflect.TypeOf(myCmd), "StderrPipe", func(*exec.Cmd) (io.ReadCloser, error) {
				return ioutil.NopCloser(strings.NewReader("Duration: 00:01:00.00 time=00:00:30.00")), nil
			})
			myPatches.ApplyMethod(reflect.TypeOf(myCmd), "Start", func(*exec.Cmd) error { return nil })
			myPatches.ApplyFunc(id3v2.Open, func(name string, opts id3v2.Options) (*id3v2.Tag, error) { return myTag, nil })
			myPatches.ApplyMethod(reflect.TypeOf(myTag), "Save", func(*id3v2.Tag) error { return nil })

			convert(&mp4FileName, &mp3FileName, &myID, myProgressBar)

			So(myProgressBar.Current(), ShouldEqual, 31)
		})
	})
}

func TestDownload(t *testing.T) {
	Convey("Download", t, func() {
		Convey("GetVideoError_ShouldPrint", func() {

			var myBuffer bytes.Buffer
			color.SetOutput(io.Writer(&myBuffer))
			defer color.SetOutput(os.Stdout)

			var myWaitGroup sync.WaitGroup
			myWaitGroup.Add(1)
			download("", "", youtube.Client{}, &myWaitGroup)

			So(myBuffer.String(), ShouldContainSubstring, "the video id must be at least 10 characters long")
		})

		Convey("GetStreamError_ShouldPrint", func() {
			myClinet := &youtube.Client{}

			myPatches := NewPatches()
			defer myPatches.Reset()

			myPatches.ApplyMethod(reflect.TypeOf(myClinet), "GetVideo", func(*youtube.Client, string) (*youtube.Video, error) {
				myFormat := &youtube.Format{}
				myVideo := &youtube.Video{
					Formats: []youtube.Format{*myFormat},
				}
				return myVideo, nil
			})

			var myBuffer bytes.Buffer
			color.SetOutput(io.Writer(&myBuffer))
			defer color.SetOutput(os.Stdout)

			var myWaitGroup sync.WaitGroup
			myWaitGroup.Add(1)
			download("", "", *myClinet, &myWaitGroup)

			So(myBuffer.String(), ShouldContainSubstring, "cipher not found")
		})

		Convey("OSCreateError_ShouldPrint", func() {
			myClinet := &youtube.Client{}

			myPatches := NewPatches()
			defer myPatches.Reset()

			myPatches.ApplyMethod(reflect.TypeOf(myClinet), "GetVideo", func(*youtube.Client, string) (*youtube.Video, error) {
				myFormat := &youtube.Format{}
				myVideo := &youtube.Video{
					Formats: []youtube.Format{*myFormat},
				}
				return myVideo, nil
			})
			myPatches.ApplyMethod(reflect.TypeOf(myClinet), "GetStream", func(*youtube.Client, *youtube.Video, *youtube.Format) (*http.Response, error) {
				myResponse := &http.Response{}
				myResponse.Body = ioutil.NopCloser(strings.NewReader(""))
				return myResponse, nil
			})
			myPatches.ApplyFunc(os.Create, func(name string) (*os.File, error) {
				return nil, errors.New("OSCreateError")
			})

			var myBuffer bytes.Buffer
			color.SetOutput(io.Writer(&myBuffer))
			defer color.SetOutput(os.Stdout)

			var myWaitGroup sync.WaitGroup
			myWaitGroup.Add(1)
			download("", "", *myClinet, &myWaitGroup)

			So(myBuffer.String(), ShouldContainSubstring, "OSCreateError")
		})

		Convey("IOCopyError_ShouldPrint", func() {
			myClinet := &youtube.Client{}

			myPatches := NewPatches()
			defer myPatches.Reset()

			myPatches.ApplyMethod(reflect.TypeOf(myClinet), "GetVideo", func(*youtube.Client, string) (*youtube.Video, error) {
				myFormat := &youtube.Format{}
				myVideo := &youtube.Video{
					Formats: []youtube.Format{*myFormat},
				}
				return myVideo, nil
			})
			myPatches.ApplyMethod(reflect.TypeOf(myClinet), "GetStream", func(*youtube.Client, *youtube.Video, *youtube.Format) (*http.Response, error) {
				myResponse := &http.Response{}
				myResponse.Body = ioutil.NopCloser(strings.NewReader(""))
				return myResponse, nil
			})
			myPatches.ApplyFunc(os.Create, func(name string) (*os.File, error) { return nil, nil })
			myPatches.ApplyFunc(io.Copy, func(dst io.Writer, src io.Reader) (written int64, err error) {
				return 0, errors.New("IOCopyError")
			})

			var myBuffer bytes.Buffer
			color.SetOutput(io.Writer(&myBuffer))
			defer color.SetOutput(os.Stdout)

			var myWaitGroup sync.WaitGroup
			myWaitGroup.Add(1)
			download("", "", *myClinet, &myWaitGroup)

			So(myBuffer.String(), ShouldContainSubstring, "IOCopyError")
		})

		Convey("OSRemoveError_ShouldPrint", func() {
			myClinet := &youtube.Client{}
			myFile := &os.File{}

			myPatches := NewPatches()
			defer myPatches.Reset()

			myPatches.ApplyMethod(reflect.TypeOf(myClinet), "GetVideo", func(*youtube.Client, string) (*youtube.Video, error) {
				myFormat := &youtube.Format{}
				myVideo := &youtube.Video{
					Formats: []youtube.Format{*myFormat},
				}
				return myVideo, nil
			})
			myPatches.ApplyMethod(reflect.TypeOf(myClinet), "GetStream", func(*youtube.Client, *youtube.Video, *youtube.Format) (*http.Response, error) {
				myResponse := &http.Response{}
				myResponse.Body = ioutil.NopCloser(strings.NewReader(""))
				return myResponse, nil
			})
			myPatches.ApplyMethod(reflect.TypeOf(myFile), "Close", func(*os.File) error { return nil })
			myPatches.ApplyFunc(os.Create, func(name string) (*os.File, error) {
				return myFile, nil
			})
			myPatches.ApplyFunc(io.Copy, func(dst io.Writer, src io.Reader) (written int64, err error) { return 0, nil })
			myPatches.ApplyFunc(convert, func(*string, *string, *string, *uiprogress.Bar) {})
			myPatches.ApplyFunc(os.Remove, func(string) error { return errors.New("OSRemoveError") })

			var myBuffer bytes.Buffer
			color.SetOutput(io.Writer(&myBuffer))
			defer color.SetOutput(os.Stdout)

			var myWaitGroup sync.WaitGroup
			myWaitGroup.Add(1)
			download("", "", *myClinet, &myWaitGroup)

			So(myBuffer.String(), ShouldContainSubstring, "OSRemoveError")
		})

		Convey("ProgressBar_ShouldDisplayStep", func() {
			myClinet := &youtube.Client{}
			myFile := &os.File{}
			var ProgressBar *uiprogress.Bar

			myPatches := NewPatches()
			defer myPatches.Reset()

			myPatches.ApplyMethod(reflect.TypeOf(myClinet), "GetVideo", func(*youtube.Client, string) (*youtube.Video, error) {
				myFormat := &youtube.Format{}
				myVideo := &youtube.Video{
					Formats: []youtube.Format{*myFormat},
				}
				return myVideo, nil
			})
			myPatches.ApplyMethod(reflect.TypeOf(myClinet), "GetStream", func(*youtube.Client, *youtube.Video, *youtube.Format) (*http.Response, error) {
				myResponse := &http.Response{}
				myResponse.ContentLength = 100
				myResponse.Body = ioutil.NopCloser(strings.NewReader(""))
				return myResponse, nil
			})
			myPatches.ApplyMethod(reflect.TypeOf(myFile), "Close", func(*os.File) error {
				So(string(ProgressBar.Bytes()), ShouldContainSubstring, "Downloading")
				return nil
			})
			myPatches.ApplyFunc(os.Create, func(name string) (*os.File, error) {
				return myFile, nil
			})
			myPatches.ApplyFunc(io.TeeReader, func(r io.Reader, w io.Writer) io.Reader {
				ProgressBar = (w.(uiProgressWriter)).ProgressBar
				return nil
			})
			myPatches.ApplyFunc(io.Copy, func(dst io.Writer, src io.Reader) (written int64, err error) { return 0, nil })
			myPatches.ApplyFunc(convert, func(*string, *string, *string, *uiprogress.Bar) {})
			myPatches.ApplyFunc(os.Remove, func(string) error { return nil })
			myPatches.ApplyFunc(color.IsLikeInCmd, func() bool { return true })
			myPatches.ApplyFunc(color.Set, func(colors ...color.Color) (int, error) { return 0, nil })

			var myWaitGroup sync.WaitGroup
			myWaitGroup.Add(1)
			download("", "", *myClinet, &myWaitGroup)

			So(string(ProgressBar.Bytes()), ShouldContainSubstring, "Converting")
		})
	})
}

func TestGetFiles(t *testing.T) {
	Convey("GetFiles", t, func() {
		Convey("PathNoExist_ShouldCreate", func() {
			actualCalled := false

			myPatches := NewPatches()
			defer myPatches.Reset()

			myPatches.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
				return nil, os.ErrNotExist
			})

			myPatches.ApplyFunc(os.Mkdir, func(name string, perm os.FileMode) error {
				actualCalled = true
				So(name, ShouldEqual, "MyPath")
				return nil
			})

			myPatches.ApplyFunc(ioutil.ReadDir, func(dirname string) ([]os.FileInfo, error) {
				return nil, nil
			})

			getFiles("MyPath")

			So(actualCalled, ShouldBeTrue)
		})

		Convey("IOUtilReadDirError_ShouldPrint", func() {
			myPatches := NewPatches()
			defer myPatches.Reset()

			myPatches.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
				return nil, nil
			})

			myPatches.ApplyFunc(ioutil.ReadDir, func(dirname string) ([]os.FileInfo, error) {
				return nil, errors.New("IOUtilReadDirError")
			})

			var myBuffer bytes.Buffer
			color.SetOutput(io.Writer(&myBuffer))
			defer color.SetOutput(os.Stdout)

			getFiles("")

			So(myBuffer.String(), ShouldContainSubstring, "IOUtilReadDirError")
		})

		Convey("AllRight_ShouldReturnReadDir", func() {
			myPatches := NewPatches()
			defer myPatches.Reset()

			myPatches.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
				return nil, nil
			})

			myFiles := []os.FileInfo{}

			myPatches.ApplyFunc(ioutil.ReadDir, func(dirname string) ([]os.FileInfo, error) {
				return myFiles, nil
			})

			myResult := getFiles("MyPath")

			So(myResult, ShouldResemble, &myFiles)
		})
	})
}

func TestGetPlayList(t *testing.T) {
	Convey("GetPlayList", t, func() {
		Convey("NewRequestError_ShouldPrint", func() {
			myPatches := NewPatches()
			defer myPatches.Reset()

			myPatches.ApplyFunc(http.NewRequest, func(method, url string, body io.Reader) (*http.Request, error) {
				return nil, errors.New("NewRequestError")
			})

			var myBuffer bytes.Buffer
			color.SetOutput(io.Writer(&myBuffer))
			defer color.SetOutput(os.Stdout)

			getPlayList("")

			So(myBuffer.String(), ShouldContainSubstring, "NewRequestError")
		})

		Convey("DoRequestError_ShouldPrint", func() {

			myPatches := NewPatches()
			defer myPatches.Reset()

			myTestServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))

			myPatches.ApplyFunc(http.NewRequest, func(method, url string, body io.Reader) (*http.Request, error) {
				myRequest, _ := http.NewRequestWithContext(context.Background(), method, myTestServer.URL, nil)
				myRequest.URL = nil
				return myRequest, nil
			})

			var myBuffer bytes.Buffer
			color.SetOutput(io.Writer(&myBuffer))
			defer color.SetOutput(os.Stdout)

			getPlayList("")

			So(myBuffer.String(), ShouldContainSubstring, "http: nil Request.URL")
		})

		Convey("IOUtilReadAllError_ShouldPrint", func() {

			myPatches := NewPatches()
			defer myPatches.Reset()

			myTestServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))

			myPatches.ApplyFunc(http.NewRequest, func(method, url string, body io.Reader) (*http.Request, error) {
				return http.NewRequestWithContext(context.Background(), method, myTestServer.URL, nil)
			})

			myPatches.ApplyFunc(ioutil.ReadAll, func(r io.Reader) ([]byte, error) { return nil, errors.New("IOUtilReadAllError") })

			var myBuffer bytes.Buffer
			color.SetOutput(io.Writer(&myBuffer))
			defer color.SetOutput(os.Stdout)

			getPlayList("")

			So(myBuffer.String(), ShouldContainSubstring, "IOUtilReadAllError")
		})

		Convey("JSONUnmarshalError_ShouldPrint", func() {

			myPatches := NewPatches()
			defer myPatches.Reset()

			myTestServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))

			myPatches.ApplyFunc(http.NewRequest, func(method, url string, body io.Reader) (*http.Request, error) {
				return http.NewRequestWithContext(context.Background(), method, myTestServer.URL, nil)
			})

			myPatches.ApplyFunc(json.Unmarshal, func(data []byte, v interface{}) error { return errors.New("JSONUnmarshalError") })

			var myBuffer bytes.Buffer
			color.SetOutput(io.Writer(&myBuffer))
			defer color.SetOutput(os.Stdout)

			getPlayList("")

			So(myBuffer.String(), ShouldContainSubstring, "JSONUnmarshalError")
		})

		Convey("AllRight_ShouldGetList", func() {

			myPatches := NewPatches()
			defer myPatches.Reset()

			myTestServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				_, _ = w.Write([]byte(`{"getVector":{"items":[{"f":"123","tt":"abc"}]}}`))
			}))

			myPatches.ApplyFunc(http.NewRequest, func(method, url string, body io.Reader) (*http.Request, error) {
				So(url, ShouldEndWith, "456")
				return http.NewRequestWithContext(context.Background(), method, myTestServer.URL, nil)
			})

			myList := getPlayList("456")

			So(myList.Vector.Items[0].ID, ShouldEqual, "123")
			So(myList.Vector.Items[0].Title, ShouldEqual, "abc")
		})
	})
}

func TestHasID(t *testing.T) {
	Convey("HasID", t, func() {
		Convey("NotMP3_ShouldBeFalse", func() {
			So(hasID("aaa", ""), ShouldBeFalse)
		})

		Convey("OpenTagFail_ShouldBeFalse", func() {
			myPatches := NewPatches()
			defer myPatches.Reset()

			myPatches.ApplyFunc(id3v2.Open, func(name string, opts id3v2.Options) (*id3v2.Tag, error) { return nil, errors.New("") })

			So(hasID("1.mp3", ""), ShouldBeFalse)
		})

		Convey("GetIDEqual_ShouldBeTrue", func() {
			myPatches := NewPatches()
			defer myPatches.Reset()

			myTag := id3v2.NewEmptyTag()

			myPatches.ApplyFunc(id3v2.Open, func(name string, opts id3v2.Options) (*id3v2.Tag, error) { return myTag, nil })
			myPatches.ApplyMethod(reflect.TypeOf(myTag), "GetTextFrame", func(tag *id3v2.Tag, id string) id3v2.TextFrame {
				So(id, ShouldEqual, "TPUB")
				return id3v2.TextFrame{Text: "123"}
			})

			So(hasID("1.mp3", "123"), ShouldBeTrue)
		})
	})
}

func TestCheckFFMpeg(t *testing.T) {
	Convey("CheckFFMpeg", t, func() {
		Convey("FFMpegNotFound_ShouldPrint", func() {

			myPatches := NewPatches()
			defer myPatches.Reset()

			myPatches.ApplyFunc(exec.LookPath, func(file string) (string, error) { return "", nil })

			var myBuffer bytes.Buffer

			color.SetOutput(io.Writer(&myBuffer))
			defer color.SetOutput(os.Stdout)

			checkFFMpeg()

			So(myBuffer.String(), ShouldContainSubstring, "FFMpeg Not Found!")
		})

		Convey("Windows_ShouldPrintWindowsInfomation", func() {

			myPatches := NewPatches()
			defer myPatches.Reset()

			myPatches.ApplyFunc(exec.LookPath, func(file string) (string, error) { return "", nil })
			myPatches.ApplyGlobalVar(&runtimeGOOS, "windows")

			var myBuffer bytes.Buffer

			color.SetOutput(io.Writer(&myBuffer))
			defer color.SetOutput(os.Stdout)

			checkFFMpeg()

			So(myBuffer.String(), ShouldContainSubstring, "Windows")
		})

		Convey("MacOS_ShouldPrintMacOSInfomation", func() {

			myPatches := NewPatches()
			defer myPatches.Reset()

			myPatches.ApplyFunc(exec.LookPath, func(file string) (string, error) { return "", nil })
			myPatches.ApplyGlobalVar(&runtimeGOOS, "darwin")

			var myBuffer bytes.Buffer

			color.SetOutput(io.Writer(&myBuffer))
			defer color.SetOutput(os.Stdout)

			checkFFMpeg()

			So(myBuffer.String(), ShouldContainSubstring, "MacOS")
		})

		Convey("Linux_ShouldPrintLinuxInfomation", func() {

			myPatches := NewPatches()
			defer myPatches.Reset()

			myPatches.ApplyFunc(exec.LookPath, func(file string) (string, error) { return "", nil })
			myPatches.ApplyGlobalVar(&runtimeGOOS, "linux")

			var myBuffer bytes.Buffer

			color.SetOutput(io.Writer(&myBuffer))
			defer color.SetOutput(os.Stdout)

			checkFFMpeg()

			So(myBuffer.String(), ShouldContainSubstring, "Linux")
		})
	})
}

func TestParseFlag(t *testing.T) {
	Convey("ParseFlag", t, func() {
		Convey("ParseFlag_ShouldParseFormArgs", func() {
			os.Args = []string{
				"",
				"123",
				"-sy",
				"-p",
				"MyPath",
			}

			pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)

			playListID, isHelp, syncPath, isSync, isNoConfirm := parseFlag()

			So(playListID, ShouldEqual, 123)
			So(*syncPath, ShouldEqual, "MyPath"+string(os.PathSeparator))
			So(*isHelp, ShouldBeFalse)
			So(*isSync, ShouldBeTrue)
			So(*isNoConfirm, ShouldBeTrue)

		})

		Convey("NonePath_ShouldAssignID", func() {
			os.Args = []string{
				"",
				"123",
			}

			pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)

			_, _, syncPath, _, _ := parseFlag()

			So(*syncPath, ShouldEqual, "123"+string(os.PathSeparator))

		})
	})
}

func TestMain(t *testing.T) {
	Convey("Main", t, func() {

		var myBuffer bytes.Buffer
		color.SetOutput(io.Writer(&myBuffer))
		defer color.SetOutput(os.Stdout)

		Convey("Call_ShouldParseFlag", func() {
			actualCalled := false

			myPatches := NewPatches()
			defer myPatches.Reset()

			myPatches.ApplyFunc(parseFlag, func() (playListID int, isHelp *bool, syncPath *string, isSync *bool, isNoConfirm *bool) {
				actualCalled = true
				return
			})
			myPatches.ApplyFunc(flag.PrintDefaults, func() {})
			myPatches.ApplyFunc(os.Exit, func(code int) {})

			main()

			So(actualCalled, ShouldBeTrue)
		})

		Convey("ZeroPlaylistID_ShouldPrintUsageAndExit", func() {
			actualPrintUsageCalled := false
			actualExitCode := -1

			myPatches := NewPatches()
			defer myPatches.Reset()

			myPatches.ApplyFunc(parseFlag, func() (playListID int, isHelp *bool, syncPath *string, isSync *bool, isNoConfirm *bool) {
				return
			})

			myPatches.ApplyFunc(printUsage, func() {
				actualPrintUsageCalled = true
			})

			myPatches.ApplyFunc(os.Exit, func(code int) {
				actualExitCode = code
			})

			main()

			So(actualPrintUsageCalled, ShouldBeTrue)
			So(actualExitCode, ShouldBeZeroValue)
		})

		Convey("CheckFFMpegFail_ShouldExit", func() {
			actualExitCode := -1

			myPatches := NewPatches()
			defer myPatches.Reset()

			myPatches.ApplyFunc(parseFlag, func() (playListID int, isHelp *bool, syncPath *string, isSync *bool, isNoConfirm *bool) {
				playListID = 1
				isHelp = new(bool)
				*isHelp = false
				return
			})

			myPatches.ApplyFunc(checkFFMpeg, func() bool { return false })

			myPatches.ApplyFunc(os.Exit, func(code int) {
				actualExitCode = code
			})

			main()

			So(actualExitCode, ShouldBeZeroValue)
		})

		Convey("ItemId_ShouldCheckhasID", func() {
			expected := "1"
			actualID := ""

			myPatches := NewPatches()
			defer myPatches.Reset()

			myPatches.ApplyFunc(parseFlag, func() (playListID int, isHelp *bool, syncPath *string, isSync *bool, isNoConfirm *bool) {
				playListID = 1
				isHelp = new(bool)
				isSync = new(bool)
				syncPath = new(string)
				*isHelp = false
				*isSync = false
				*syncPath = ""
				return
			})

			myPatches.ApplyFunc(checkFFMpeg, func() bool { return true })

			myPatches.ApplyFunc(getPlayList, func(myID string) *playList {
				myList := playList{}
				myList.Vector.Items = []listItem{
					{ID: ""},
					{ID: expected},
				}
				return &myList
			})

			myPatches.ApplyFunc(getFiles, func(myPath string) *[]os.FileInfo {
				return &[]os.FileInfo{&mockFile{}}
			})

			myPatches.ApplyFunc(hasID, func(myFile string, myID string) bool {
				actualID = myID
				return true
			})

			main()

			So(actualID, ShouldEqual, expected)
		})

		Convey("NotHasId_ShouldDownload", func() {
			expected := "1"
			actualID := ""

			myPatches := NewPatches()
			defer myPatches.Reset()

			myPatches.ApplyFunc(parseFlag, func() (playListID int, isHelp *bool, syncPath *string, isSync *bool, isNoConfirm *bool) {
				playListID = 1
				isHelp = new(bool)
				isSync = new(bool)
				syncPath = new(string)
				*isHelp = false
				*isSync = false
				*syncPath = ""
				return
			})

			myPatches.ApplyFunc(checkFFMpeg, func() bool { return true })

			myPatches.ApplyFunc(getPlayList, func(myID string) *playList {
				myList := playList{}
				myList.Vector.Items = []listItem{{ID: expected}}
				return &myList
			})

			myPatches.ApplyFunc(getFiles, func(myPath string) *[]os.FileInfo {
				return &[]os.FileInfo{&mockFile{}}
			})

			myPatches.ApplyFunc(hasID, func(myFile string, myID string) bool { return false })

			myPatches.ApplyFunc(download, func(myID string, myPath string, myClient youtube.Client, myWaitGroup *sync.WaitGroup) {
				actualID = myID
				myWaitGroup.Done()
			})

			main()

			So(actualID, ShouldEqual, expected)
		})

		Convey("IsSync_ShouldDeletFileNotInList", func() {
			actualCalled := false

			myPatches := NewPatches()
			defer myPatches.Reset()

			myPatches.ApplyFunc(parseFlag, func() (playListID int, isHelp *bool, syncPath *string, isSync *bool, isNoConfirm *bool) {
				playListID = 1
				isHelp = new(bool)
				isSync = new(bool)
				syncPath = new(string)
				*isHelp = false
				*isSync = true
				*syncPath = ""
				return
			})

			myPatches.ApplyFunc(checkFFMpeg, func() bool { return true })

			myPatches.ApplyFunc(getPlayList, func(myID string) *playList {
				myList := playList{}
				myList.Vector.Items = []listItem{}
				return &myList
			})

			myPatches.ApplyFunc(getFiles, func(myPath string) *[]os.FileInfo { return &[]os.FileInfo{} })

			myPatches.ApplyFunc(hasID, func(myFile string, myID string) bool { return false })

			myPatches.ApplyFunc(download, func(myID string, myPath string, myClient youtube.Client, myWaitGroup *sync.WaitGroup) {
				myWaitGroup.Done()
			})

			myPatches.ApplyFunc(deletFileNotInList, func(myFiles *[]os.FileInfo, myList *playList, syncPath *string, isNoConfirm *bool) {
				actualCalled = true
			})

			main()

			So(actualCalled, ShouldBeTrue)
		})
	})
}
