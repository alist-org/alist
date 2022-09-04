package sftp

import (
	"path"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// do others that not defined in Driver interface

func (d *SFTP) initClient() error {
	var auth ssh.AuthMethod
	if d.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(d.PrivateKey))
		if err != nil {
			return err
		}
		auth = ssh.PublicKeys(signer)
	} else {
		auth = ssh.Password(d.Password)
	}
	config := &ssh.ClientConfig{
		User:            d.Username,
		Auth:            []ssh.AuthMethod{auth},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	conn, err := ssh.Dial("tcp", d.Address, config)
	if err != nil {
		return err
	}
	d.client, err = sftp.NewClient(conn)
	return err
}

func (d *SFTP) remove(remotePath string) error {
	f, err := d.client.Stat(remotePath)
	if err != nil {
		return nil
	}
	if f.IsDir() {
		return d.removeDirectory(remotePath)
	} else {
		return d.removeFile(remotePath)
	}
}

func (d *SFTP) removeDirectory(remotePath string) error {
	remoteFiles, err := d.client.ReadDir(remotePath)
	if err != nil {
		return err
	}
	for _, backupDir := range remoteFiles {
		remoteFilePath := path.Join(remotePath, backupDir.Name())
		if backupDir.IsDir() {
			err := d.removeDirectory(remoteFilePath)
			if err != nil {
				return err
			}
		} else {
			err := d.removeFile(remoteFilePath)
			if err != nil {
				return err
			}
		}
	}
	return d.client.RemoveDirectory(remotePath)
}

func (d *SFTP) removeFile(remotePath string) error {
	return d.client.Remove(path.Join(remotePath))
}
