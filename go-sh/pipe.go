package sh

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
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

func (s *Session) UnmarshalJSON(data interface{}) (err error) {
	bufrw := bytes.NewBuffer(nil)
	s.Stdout = bufrw
	if err = s.Run(); err != nil {
		return
	}
	return json.NewDecoder(bufrw).Decode(data)
}

// // unmarshal command output into xml
func (s *Session) UnmarshalXML(data interface{}) (err error) {
	bufrw := bytes.NewBuffer(nil)
	s.Stdout = bufrw
	if err = s.Run(); err != nil {
		return
	}
	return xml.NewDecoder(bufrw).Decode(data)
}
func (s *Session) displayCommandChain() {
	var cmds []string
	for _, cmd := range s.cmds {
		cmds = append(cmds, strings.Join(cmd.Args, " "))
	}
	s.writePrompt(strings.Join(cmds, " | "))
}

func (s *Session) Start() error {
	if s.ShowCMD {
		s.displayCommandChain()
	}
	return s.executeCommandChain(0, nil)
}

func (s *Session) executeCommandChain(index int, stdin *io.PipeReader) error {
	if index >= len(s.cmds) {
		return nil
	}
	pipeCount := s.determinePipeCount(index)
	pipeReaders, pipeWriters := createPipes(pipeCount)
	var writers []io.Writer
	for _, writer := range pipeWriters {
		writers = append(writers, writer)
		s.PipeWriters = append(s.PipeWriters, writer)
	}

	multiWriter := io.MultiWriter(writers...)
	cmd := s.cmds[index]

	if index == 0 {
		cmd.Stdin = s.Stdin
	} else {
		cmd.Stdin = stdin
	}

	if s.isLastCommand(index) && len(s.backupCmds) == 0 {
		cmd.Stdout = s.Stdout
		cmd.Stderr = s.Stderr
	} else {
		cmd.Stdout = multiWriter
		cmd.Stderr = s.selectStderr()
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	if s.isLastCommand(index) && len(s.backupCmds) != 0 {
		return s.executeBackupCommands(pipeReaders)
	}
	return s.executeCommandChain(index+1, pipeReaders[0])
}

func (s *Session) executeBackupCommands(readers []*io.PipeReader) error {
	for idx, reader := range readers {
		cmd := s.backupCmds[idx]
		cmdOutput := bytes.Buffer{}
		cmd.Stdin = reader
		cmd.Stdout = io.MultiWriter(s.Stdout, &cmdOutput)
		cmd.Stderr = s.selectStderr()

		if err := cmd.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Session) determinePipeCount(index int) int {
	if s.isLastCommand(index) && len(s.backupCmds) != 0 {
		return len(s.backupCmds)
	}
	return 1
}

func (s *Session) isLastCommand(index int) bool {
	return index == len(s.cmds)-1
}

func (s *Session) selectStderr() io.Writer {
	if s.PipeStdErrors {
		return s.Stderr
	}
	return os.Stderr
}

func createPipes(count int) ([]*io.PipeReader, []*io.PipeWriter) {
	readers := make([]*io.PipeReader, count)
	writers := make([]*io.PipeWriter, count)

	for i := 0; i < count; i++ {
		r, w := io.Pipe()
		readers[i] = r
		writers[i] = w
	}

	return readers, writers
}

// Should be call after Start()
// only catch the last command error
func (s *Session) Wait() error {
	var pipeErr, lastErr error
	for idx, writter := range s.PipeWriters {
		if idx < len(s.cmds) {
			cmd := s.cmds[idx]
			if lastErr = cmd.Wait(); lastErr != nil {
				pipeErr = lastErr
			}
		}
		writter.Close()
	}
	var pipeErrs []error
	for _, cmd := range s.backupCmds {
		if err := cmd.Wait(); err != nil {
			pipeErrs = append(pipeErrs, err)
		}
	}

	if s.PipeFail {
		return pipeErr
	}

	pipeErrs = append([]error{pipeErr}, pipeErrs...)
	return errors.Join(pipeErrs...)
}

func (s *Session) Kill(sig os.Signal) {
	for _, cmd := range s.cmds {
		if cmd.Process != nil {
			cmd.Process.Signal(sig)
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

func (s *Session) Output() (out []byte, err error) {
	oldout := s.Stdout
	defer func() {
		s.Stdout = oldout
	}()
	stdout := bytes.NewBuffer(nil)
	s.Stdout = stdout
	err = s.Run()
	out = stdout.Bytes()
	return
}

func (s *Session) WriteStdout(f string) error {
	oldout := s.Stdout
	defer func() {
		s.Stdout = oldout
	}()

	out, err := os.Create(f)
	if err != nil {
		return err
	}
	defer out.Close()
	s.Stdout = out
	return s.Run()
}

func (s *Session) AppendStdout(f string) error {
	oldout := s.Stdout
	defer func() {
		s.Stdout = oldout
	}()

	out, err := os.OpenFile(f, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer out.Close()
	s.Stdout = out
	return s.Run()
}

func (s *Session) CombinedOutput() (out []byte, err error) {
	oldout := s.Stdout
	olderr := s.Stderr
	defer func() {
		s.Stdout = oldout
		s.Stderr = olderr
	}()
	stdout := bytes.NewBuffer(nil)
	s.Stdout = stdout
	s.Stderr = stdout

	err = s.Run()
	out = stdout.Bytes()
	return
}
