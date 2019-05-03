package main

import (
	logs "auth_service_template/logger"
	"auth_service_template/models"
	"auth_service_template/server"
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

var (
	logger   *logs.Log
	Env      map[string]string
	db       *models.DB
	instance *server.Instance
)

func init() {
	Env, _ = godotenv.Read()
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// [Systemd] stop.
	// create a "returnCode" channel which will be the return code of the application
	var returnCode = make(chan int)
	// finishUP channel signals the application to finish up
	var finishUP = make(chan struct{})
	// done channel signals the signal handler that the application has completed
	var done = make(chan struct{})
	// gracefulStop is a channel of os.Signals that we will watch for -SIGTERM
	var gracefulStop = make(chan os.Signal)
	// watch for SIGTERM and SIGINT from the operating system, and notify the app on
	// the gracefulStop channel
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	signal.Notify(gracefulStop, syscall.SIGKILL)
	signal.Notify(gracefulStop, syscall.SIGSTOP)

	// launch a worker whose job it is to always watch for gracefulStop signals
	go func() {
		// wait for our os signal to stop the app
		// on the graceful stop channel
		// this goroutine will block until we get an OS signal
		sig := <-gracefulStop
		fmt.Printf("caught sig: %+v", sig)
		// send message on "finish up" channel to tell the app to
		// gracefully shutdown
		finishUP <- struct{}{}
		// wait for word back if we finished or not
		select {
		case <-time.After(30 * time.Second):
			// timeout after 30 seconds waiting for app to finish,
			// our application should Exit(1)
			returnCode <- (0xf)
		case <-done:
			// if we got a message on done, we finished, so end app
			// our application should Exit(0)
			returnCode <- 0
		}
	}()
	// [Systemd] stop.

	logger = logs.NewLogger()
	ctx_log, cancel_log := context.WithCancel(context.Background())
	go logger.Subscriber(ctx_log)

	// [Server instance block]
	instance = server.NewInstance(logger, &Env)
	go instance.Start()
	// [Server instance block]

	for {
		select {
		case <-finishUP:
			fmt.Println("Stoped")
			instance.Shutdown()
			<-time.After(time.Duration(7) * time.Second)
			cancel_log()
			done <- struct{}{}
			os.Exit(<-returnCode)
		}
	}
}
