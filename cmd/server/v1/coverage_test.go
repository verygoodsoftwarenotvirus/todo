package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func TestRunMain(t *testing.T) {
	d, err := time.ParseDuration(os.Getenv("RUNTIME_DURATION"))
	if err != nil {
		log.Fatal(err)
	}

	go main()

	sigs := make(chan os.Signal, 1)
	signal.Notify(
		sigs,
		// syscall.SIGABRT,
		// syscall.SIGHUP,
		// syscall.SIGKILL,
		// syscall.SIGQUIT,
		// syscall.SIGSTOP,

		syscall.SIGABRT,
		syscall.SIGALRM,
		syscall.SIGBUS,
		syscall.SIGCHLD,
		syscall.SIGCLD,
		syscall.SIGCONT,
		syscall.SIGFPE,
		syscall.SIGHUP,
		syscall.SIGILL,
		syscall.SIGINT,
		syscall.SIGIO,
		syscall.SIGIOT,
		syscall.SIGKILL,
		syscall.SIGPIPE,
		syscall.SIGPOLL,
		syscall.SIGPROF,
		syscall.SIGPWR,
		syscall.SIGQUIT,
		syscall.SIGSEGV,
		syscall.SIGSTKFLT,
		syscall.SIGSTOP,
		syscall.SIGSYS,
		syscall.SIGTERM,
		syscall.SIGTRAP,
		syscall.SIGTSTP,
		syscall.SIGTTIN,
		syscall.SIGTTOU,
		syscall.SIGUNUSED,
		syscall.SIGURG,
		syscall.SIGUSR1,
		syscall.SIGUSR2,
		syscall.SIGVTALRM,
		syscall.SIGWINCH,
		syscall.SIGXCPU,
		syscall.SIGXFSZ,
	)

	for {
		select {
		case x := <-sigs:
			log.Printf("Received signal: %v\n", x)
			return
		case <-time.After(d):
			log.Println("time passed")
			return
		}
	}
}
