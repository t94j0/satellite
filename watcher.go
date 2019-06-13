package main

import (
	"log"

	"github.com/fsnotify/fsnotify"
)

func createWatcher(path string, f func() error) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Remove == fsnotify.Remove {
					log.Println("Reloading.", event.Op, event.Name)
					if err := f(); err != nil {
						log.Println("Error: ", err)
					} else {
						log.Println("Successful reload")
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Error: ", err)
			}
		}
	}()

	err = watcher.Add(path)
	if err != nil {
		return err
	}

	<-done
	return nil
}
