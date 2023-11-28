package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"os/user"
	"strings"
)

func getWrfPids() []string {
	pids := make([]string, 0)
	currusr, err := user.Current()
	if err != nil {
		fmt.Printf("Cannot get current user: %s \n", err.Error())
		return nil
	}

	pscmd := exec.Command(
		"ps",
		"aux",
	)
	var psout bytes.Buffer
	pscmd.Stdout = &psout
	err = pscmd.Run()
	if err != nil {
		fmt.Printf("Cannot kill previous wrf.exe processes. ps command: %s\n", err.Error())
		return nil
	}

	scanner := bufio.NewScanner(&psout)
	for scanner.Scan() {
		p := strings.Fields(scanner.Text())
		usr := p[0]
		pid := p[1]
		exe := strings.Join(p[10:], " ")
		if strings.Contains(exe, "wrf.exe") {
			if usr == currusr.Username {
				pids = append(pids, pid)
			} else {
				fmt.Printf("WARNING: cannot kill process `%s` (pid %s) owned by user %s\n", exe, pid, usr)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Cannot scan ps output: %s\n", err.Error())
		return nil
	}

	return pids
}

func main() {
	killPreviousWrfProcesses()
}

func killPreviousWrfProcesses() {
	pids := getWrfPids()

	if pids == nil {
		// an error occurred
		return
	}

	if len(pids) == 0 {
		// no processes to kill
		fmt.Printf("No previous wrf.exe processes found.\n")
		return
	}

	args := []string{"-9"}
	args = append(args, pids...)
	cmd := exec.Command("kill", args...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Cannot kill previous wrf.exe processes: %s\nCommand output:\n%s\n", err.Error(), string(out))
		return
	}

	fmt.Printf("Previous wrf.exe processes killed: %v.\n", pids)

}
