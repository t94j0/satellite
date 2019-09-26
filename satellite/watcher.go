package main

import (
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func createWatcher(path string, durationStr string, f func() error) error {
	maxDuration, err := time.ParseDuration(durationStr)
	if err != nil {
		return err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.Wrap(err, "unable to initialize watcher")
	}
	defer watcher.Close()

	done := make(chan bool)

	go func() {
		last := time.Now()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Remove == fsnotify.Remove {
					if time.Now().Sub(last) > maxDuration {
						log.Info("Reloading.")
						log.Debug("Reloading Event: ", event.Op, event.Name)
						if err := f(); err != nil {
							log.Error("Error: ", err)
						} else {
							log.Debug("Successful reload")
						}
						last = time.Now()
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Error("Error: ", err)
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
