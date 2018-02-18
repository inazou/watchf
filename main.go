package main

import (
	"flag"
	"fmt"
	"github.com/go-fsnotify/fsnotify"
	"os"
	"strings"
	"log"
	"os/exec"
	"time"
	"path/filepath"
)

var path, _ = os.Getwd()
// 引数
// 対象のパス
var target_path = flag.String("p", path, "watch file change path")
// 実行するコマンド
var command = flag.String("c", "", "exec command if watching file changed")
var cmd string
var cmd_args []string

var lock = false

func main() {

	flag.Parse()

	var cmds = strings.Split(*command, " ")
	cmd = cmds[0]
	cmd_args = cmds[1:]

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()
	done := make(chan bool)

	go monitor(watcher, done)

	filepath_err := filepath.Walk(*target_path,
		func(path string, info os.FileInfo, err error) error {

			watcher_err := watcher.Add(path)
			if watcher_err != nil {
				panic(watcher_err)
			}
			return nil
		})

	if filepath_err != nil {
		panic(filepath_err)
	}

	log.Println("Watching:", *target_path)
	<-done
}

func monitor(watcher *fsnotify.Watcher, done chan bool) {
	for {
		select {
		case event := <-watcher.Events:
			go notify(event)
		case err := <-watcher.Errors:
			log.Println("error:", err)
			done <- false
			return
		}
	}
}

func notify(event fsnotify.Event) {

	if lock {
		return
	}

	lock = true
	defer func() {
		lock = false
	}()

	if event.Op&fsnotify.Write == fsnotify.Write ||
		event.Op&fsnotify.Create == fsnotify.Create ||
		event.Op&fsnotify.Remove == fsnotify.Remove ||
		event.Op&fsnotify.Rename == fsnotify.Rename ||
		event.Op&fsnotify.Chmod == fsnotify.Chmod {
		log.Println(event.Name, event)
		execCommand()
	}
	return
}

func execCommand() {
	wait := make(chan bool)
	go progress(wait)
	out, _ := exec.Command(cmd, cmd_args...).CombinedOutput()
	wait <- true
	fmt.Println(*command)
	fmt.Print(string(out))
}

func progress(wait chan bool) {
	for {
		select {
		case <-wait:
			return
		case <-time.After(time.Second * 2):
			return
		}
	}
}
