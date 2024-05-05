package ssh

import "github.com/ajrudzitis/ssh-resume/pty"

// SSHApplication is an interface that defines the methods that an SSH application must implement
type SSHApplication interface {
	Run(pty *pty.Pty)
}
