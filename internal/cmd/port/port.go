package port

import (
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/spf13/cobra"
)

func NewCmdPort(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "port <number>",
		Short: "show or kill processes on a port",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			kill, _ := cmd.Flags().GetBool("kill")
			return RunPort(f, args[0], kill)
		},
	}

	cmd.Flags().BoolP("kill", "k", false, "kill the process on this port")

	return cmd
}

func RunPort(f *factory.Factory, portStr string, kill bool) error {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid port: %s", portStr)
	}

	var pid, process, user string

	if runtime.GOOS == "windows" {
		pid, process, user, err = findProcessWindows(port)
	} else {
		pid, process, user, err = findProcessUnix(port)
	}

	if err != nil {
		return err
	}

	if pid == "" {
		fmt.Fprintf(f.IO.Out, "nothing running on port %d\n", port)
		return nil
	}

	fmt.Fprintf(f.IO.Out, "port %d\n", port)
	fmt.Fprintf(f.IO.Out, "  pid     %s\n", pid)
	fmt.Fprintf(f.IO.Out, "  process %s\n", process)
	fmt.Fprintf(f.IO.Out, "  user    %s\n", user)

	if kill {
		if err := killProcess(pid); err != nil {
			return err
		}
		fmt.Fprintf(f.IO.ErrOut, "killed process %s (%s) on port %d\n", pid, process, port)
	}

	return nil
}

func findProcessUnix(port int) (pid, process, user string, err error) {
	cmd := exec.Command("lsof", "-i", fmt.Sprintf(":%d", port))
	output, err := cmd.Output()
	if err != nil {
		return "", "", "", nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 2 {
		return "", "", "", nil
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 9 {
		return "", "", "", nil
	}

	user = fields[0]
	pid = fields[1]
	process = fields[0]

	if len(fields) >= 2 {
		process = fields[0]
	}

	re := regexp.MustCompile(`^([a-zA-Z0-9_-]+)`)
	for i := 8; i < len(fields); i++ {
		if match := re.FindString(fields[i]); match != "" {
			process = match
			break
		}
	}

	return pid, process, user, nil
}

func findProcessWindows(port int) (pid, process, user string, err error) {
	cmd := exec.Command("cmd", "/C", fmt.Sprintf("netstat -ano | findstr :%d", port))
	output, err := cmd.Output()
	if err != nil {
		return "", "", "", nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		pid = fields[4]
		if pid != "" {
			break
		}
	}

	if pid == "" {
		return "", "", "", nil
	}

	treeCmd := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %s", pid), "/V")
	treeOutput, err := treeCmd.Output()
	if err == nil {
		treeLines := strings.Split(strings.TrimSpace(string(treeOutput)), "\n")
		if len(treeLines) > 1 {
			fields := strings.Fields(treeLines[1])
			if len(fields) > 0 {
				process = fields[0]
			}
		}
	}

	if process == "" {
		process = "unknown"
	}

	user = "unknown"

	return pid, process, user, nil
}

func killProcess(pid string) error {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("taskkill", "/F", "/PID", pid)
		return cmd.Run()
	}

	cmd := exec.Command("kill", "-9", pid)
	return cmd.Run()
}
