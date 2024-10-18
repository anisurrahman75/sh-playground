package sh

import (
	"errors"
	"fmt"
	_ "fmt"
	"io"
	"os"
	"strings"
	_ "strings"
	"syscall"
	"time"
)

var ErrExecTimeout = errors.New("execute timeout")

// unmarshal shell output to decode json
//func (s *Session) UnmarshalJSON(data interface{}) (err error) {
//	bufrw := bytes.NewBuffer(nil)
//	s.Stdout = bufrw
//	if err = s.Run(); err != nil {
//		return
//	}
//	return json.NewDecoder(bufrw).Decode(data)
//}
//
//// unmarshal command output into xml
//func (s *Session) UnmarshalXML(data interface{}) (err error) {
//	bufrw := bytes.NewBuffer(nil)
//	s.Stdout = bufrw
//	if err = s.Run(); err != nil {
//		return
//	}
//	return xml.NewDecoder(bufrw).Decode(data)
//}

// start command
//func (s *Session) Start() (err error) {
//	s.started = true
//	var rd *io.PipeReader
//	var wr *io.PipeWriter
//	var length = len(s.cmds)
//	if s.ShowCMD {
//		var cmds = make([]string, 0, 4)
//		for _, cmd := range s.cmds {
//			cmds = append(cmds, strings.Join(cmd.Cmd.Args, " "))
//		}
//		s.writePrompt(strings.Join(cmds, " | "))
//	}
//	//var reader io.PipeReader
//	//s.start(0, &reader)
//
//	for index, cmd := range s.cmds {
//		if index == 0 {
//			cmd.Cmd.Stdin = s.Stdin
//		} else {
//			cmd.Cmd.Stdin = rd
//		}
//
//		rd, wr = io.Pipe() // create pipe
//		cmd.Cmd.Stdout = wr
//		if s.PipeStdErrors {
//			cmd.Cmd.Stderr = s.Stderr
//		} else {
//			cmd.Cmd.Stderr = os.Stderr
//		}
//
//		if index == length-1 {
//			cmd.Cmd.Stdout = s.Stdout
//			cmd.Cmd.Stderr = s.Stderr
//		}
//		err = cmd.Cmd.Start()
//		if err != nil {
//			return
//		}
//
//	}
//	return
//}

func (s *Session) Start() error {
	if s.ShowCMD {
		var cmds []string
		for _, cmd := range s.cmds {
			cmds = append(cmds, strings.Join(cmd.Cmd.Args, " "))
		}
		s.writePrompt(strings.Join(cmds, " | "))
	}

	return s.startRecursive(0, nil)
}

func (s *Session) startRecursive(index int, stdin *io.PipeReader) error {
	if index >= len(s.cmds) {
		return nil
	}

	fmt.Println("----Index:", index)

	cmd := s.cmds[index]
	if index == 0 {
		cmd.Cmd.Stdin = s.Stdin
	} else {
		cmd.Cmd.Stdin = stdin
	}

	var (
		allWriters []io.Writer
		allReaders []*io.PipeReader
	)

	for range cmd.ChildCmds {
		pipeReader, pipeWriter := io.Pipe()
		allReaders = append(allReaders, pipeReader)
		allWriters = append(allWriters, pipeWriter)
	}

	mainReader, mainWriter := io.Pipe()
	allReaders = append(allReaders, mainReader)
	allWriters = append(allWriters, mainWriter)
	multiWriter := io.MultiWriter(allWriters...)

	if index == len(s.cmds)-1 { // Last command in the pipeline.
		cmd.Cmd.Stdout = s.Stdout
		cmd.Cmd.Stderr = s.Stderr
	} else {
		cmd.Cmd.Stdout = multiWriter
		if s.PipeStdErrors {
			cmd.Cmd.Stderr = s.Stderr
		} else {
			cmd.Cmd.Stderr = os.Stderr
		}
	}

	if err := cmd.Cmd.Start(); err != nil {
		// Close all writers on error to prevent leaks.
		//for _, writer := range allWriters {
		//	writer.Close()
		//}
		return err
	}

	return s.startRecursive(index+1, allReaders[0])
}

func (s *Session) start(index int, oldReader *io.PipeReader) { // index==0 s.stdin
	length := len(s.cmds)
	cmd := s.cmds[index]
	cmd.Cmd.Stdin = oldReader
	newReader, newWritter := io.Pipe() // create pipe

	cmd.Cmd.Stdout = newWritter
	if s.PipeStdErrors {
		cmd.Cmd.Stderr = s.Stderr
	} else {
		cmd.Cmd.Stderr = os.Stderr
	}
	if index == length-1 {
		cmd.Cmd.Stdout = s.Stdout
		cmd.Cmd.Stderr = s.Stderr
	}

	err := cmd.Cmd.Start()
	if err != nil {
		return
	}
	if index == length-1 {
		return
	}

	s.start(index+1, newReader)
}

// Should be call after Start()
// only catch the last command error
func (s *Session) Wait() error {
	var pipeErr, lastErr error
	for _, cmd := range s.cmds {
		if lastErr = cmd.Cmd.Wait(); lastErr != nil {
			pipeErr = lastErr
		}
		wr, ok := cmd.Cmd.Stdout.(*io.PipeWriter)
		if ok {
			wr.Close()
		}
	}
	if s.PipeFail {
		return pipeErr
	}
	return lastErr
}

func (s *Session) Kill(sig os.Signal) {
	for _, cmd := range s.cmds {
		if cmd.Cmd.Process != nil {
			cmd.Cmd.Process.Signal(sig)
		}
	}
}

func (s *Session) WaitTimeout(timeout time.Duration) (err error) {
	select {
	case <-time.After(timeout):
		s.Kill(syscall.SIGKILL)
		return ErrExecTimeout
	case err = <-Go(s.Wait):
		return err
	}
}

func Go(f func() error) chan error {
	ch := make(chan error, 1)
	go func() {
		ch <- f()
	}()
	return ch
}

func (s *Session) Run() (err error) {
	if err = s.Start(); err != nil {
		return
	}
	if s.timeout != time.Duration(0) {
		return s.WaitTimeout(s.timeout)
	}
	return s.Wait()
}

//func (s *Session) Output() (out []byte, err error) {
//	oldout := s.Stdout
//	defer func() {
//		s.Stdout = oldout
//	}()
//	stdout := bytes.NewBuffer(nil)
//	s.Stdout = stdout
//	err = s.Run()
//	out = stdout.Bytes()
//	return
//}

//func (s *Session) WriteStdout(f string) error {
//	oldout := s.Stdout
//	defer func() {
//		s.Stdout = oldout
//	}()
//
//	out, err := os.Create(f)
//	if err != nil {
//		return err
//	}
//	defer out.Close()
//	s.Stdout = out
//	return s.Run()
//}

//func (s *Session) AppendStdout(f string) error {
//	oldout := s.Stdout
//	defer func() {
//		s.Stdout = oldout
//	}()
//
//	out, err := os.OpenFile(f, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
//	if err != nil {
//		return err
//	}
//	defer out.Close()
//	s.Stdout = out
//	return s.Run()
//}

//func (s *Session) CombinedOutput() (out []byte, err error) {
//	oldout := s.Stdout
//	olderr := s.Stderr
//	defer func() {
//		s.Stdout = oldout
//		s.Stderr = olderr
//	}()
//	stdout := bytes.NewBuffer(nil)
//	s.Stdout = stdout
//	s.Stderr = stdout
//
//	err = s.Run()
//	out = stdout.Bytes()
//	return
//}
