package main

import (
	"fmt"
	"net/http"
)

var logger = CreateLogger("main", &LoggerOptions{
	Level: Debug,
})

func httpHandleRoot(w http.ResponseWriter, r *http.Request) {
	logger.Debug("accessing /")

	user, err := HttpToUser(w, r)
	if err != nil {
		logger.Warn(fmt.Sprintf("failed creating user, err: %s", err.Error()))
		return
	}

	logger.Info(fmt.Sprintf("new user created %s", user.Id))
}

func main() {
	logger.Info("SKEWRTC SFU & Signaling server is up!")

	mux := http.DefaultServeMux
	mux.HandleFunc("/", httpHandleRoot)

	if err := http.ListenAndServe("0.0.0.0:8888", mux); err != nil {
		panic(err)
	}
}
