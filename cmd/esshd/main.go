package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"

	"github.com/creack/pty"
	"github.com/gliderlabs/ssh"
)

const tip string = `[?] TIP: ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -o Port=%s %s`

func setWinsize(f *os.File, w, h int) {
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(h), uint16(w), 0, 0})))
}

func main() {
	log.SetFlags(0)
	log.Println("[+] esshd v0.0.0 (2020-03-05)")
	if len(os.Args) == 2 {
		log.Fatal("[!] Missing argument #2 (host:port).")
	}
	if len(os.Args) == 1 {
		log.Fatal("[!] Missing argument #1 (shell path).")
	}
	command := os.Args[1]
	port := os.Args[2]
	var commandStr string
	var argsSlice []string
	commandSlice := strings.Fields(command)
	commandBin = commandSlice[0]
	if len(commandSlice) > 1 {
		argsSlice = commandSlice[1:]
	}
	ssh.Handle(func(s ssh.Session) {
		_, err := os.Stat("/esshd.txt")
		if err == nil {
			b, _ := os.ReadFile("/esshd.txt")
			s.Write(b)
		}
		cmd := exec.Command(commandBin, argsSlice...)
		cmd.Env = append(os.Environ(), "HOME=/")
		ptyReq, winCh, isPty := s.Pty()
		if isPty {
			log.Println(fmt.Sprintf("[+] Connected SSH client (%s@%s).", s.User(), s.RemoteAddr()))
			cmd.Env = append(cmd.Env, fmt.Sprintf("TERM=%s", ptyReq.Term))
			f, err := pty.Start(cmd)
			defer func() {
				_ = f.Close()
				log.Println(fmt.Sprintf("[+] Disconnected SSH client (%s@%s).", s.User(), s.RemoteAddr()))
			}()
			if err != nil {
				panic(err)
			}
			go func() {
				for win := range winCh {
					setWinsize(f, win.Width, win.Height)
				}
			}()
			go func() {
				io.Copy(f, s) // stdin
			}()
			io.Copy(s, f) // stdout
			cmd.Wait()
		} else {
			io.WriteString(s, "No PTY requested.\n")
			s.Exit(1)
		}
	})
	hp := strings.Split(port, ":")
	if len(hp) == 1 {
		log.Fatal("[!] Argument #2 should be in the format `host:port` e.g. `127.0.0.1:2222`")
	}
	if len(hp[0]) == 0 {
		hp[0] = "0.0.0.0"
	}
	log.Println("[+] Starting ephemeral sshd...")
	log.Println(fmt.Sprintf("[+] Listening on %s:%s...", hp[0], hp[1]))
	log.Println(fmt.Sprintf(tip, hp[1], hp[0]))
	log.Fatal(ssh.ListenAndServe(port, nil))
}
