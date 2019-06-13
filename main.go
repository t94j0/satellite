package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/fsnotify/fsnotify"
)

const ServerPath string = "/Users/maxh/Programming/DEITYSHADOW/base"

var paths *Paths

func matchHandler(w http.ResponseWriter, req *http.Request, path Path) {
	data, err := ioutil.ReadFile(path.FullPath)
	if err != nil {
		log.Println(err)
		return
	}
	io.WriteString(w, string(data))
	if path.Once {
		paths.Remove(path)
	}
	return
}

func noMatchHandler(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "Errored\n")
}

func doesNotExistHandler(w http.ResponseWriter, req *http.Request) {
	// TODO: Write a 404 handler
	io.WriteString(w, "404\n")
	return
}

func handler(w http.ResponseWriter, req *http.Request) {
	file, exists := paths.Match(req.URL.Path)

	if !exists {
		doesNotExistHandler(w, req)
	} else if file.ShouldHost(req) {
		matchHandler(w, req, file)
	} else {
		noMatchHandler(w, req)
	}
}

func createWatcher(path string) {
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
				if event.Op&fsnotify.Create == fsnotify.Create {
					log.Println("Reloading.", event.Name)
					if err := paths.Reload(); err != nil {
						log.Println("error:", err)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(ServerPath)
	if err != nil {
		log.Fatal(err)
	}

	<-done
}

func main() {
	var err error
	paths, err = NewPaths(ServerPath)
	if err != nil {
		log.Fatal(err)
	}

	go createWatcher(ServerPath)

	http.HandleFunc("/", handler)
	log.Print("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
