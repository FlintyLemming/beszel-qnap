package main

import (
	"fmt"
	"net/http/cgi"
	"os"

	"beszel-qnap/internal/agentctl"
	"beszel-qnap/internal/auth"
	"beszel-qnap/internal/updater"
	"beszel-qnap/internal/web"
)

func main() {
	paths, err := agentctl.ResolvePaths()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed resolving QPKG paths: %v\n", err)
		os.Exit(1)
	}
	mgr := agentctl.NewManager(paths)

	// CGI mode when invoked by QTS Apache (GATEWAY_INTERFACE is set).
	if os.Getenv("GATEWAY_INTERFACE") != "" && (len(os.Args) < 2 || os.Args[1] != "ctl") {
		srv := web.NewServer(mgr, auth.NewQTSVerifier())
		if err := cgi.Serve(srv); err != nil {
			fmt.Fprintf(os.Stderr, "cgi error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// CLI control mode for BeszelAgent.sh and detached background jobs.
	if len(os.Args) < 3 || os.Args[1] != "ctl" {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[2] {
	case "start":
		if err := mgr.Start(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "start-if-configured":
		if err := mgr.StartIfConfigured(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "stop":
		if err := mgr.Stop(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "restart":
		if err := mgr.Restart(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "status":
		pid := mgr.FindRunningPID()
		if pid > 0 {
			fmt.Printf("beszel-agent is running (pid=%d)\n", pid)
			os.Exit(0)
		}
		fmt.Println("beszel-agent is stopped")
		os.Exit(1)
	case "download-job":
		if len(os.Args) < 5 {
			fmt.Fprintln(os.Stderr, "usage: ctl download-job <reqPath> <jobID>")
			os.Exit(1)
		}
		if err := updater.RunDownloadJob(mgr, os.Args[3], os.Args[4]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage: api.cgi ctl {start|start-if-configured|stop|restart|status|download-job}")
}
