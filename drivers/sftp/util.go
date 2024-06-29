package sftp

import (
	"path"

	"github.com/pkg/sftp"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

// do others that not defined in Driver interface

func (d *SFTP) initClient() error {
	var auth ssh.AuthMethod
	if len(d.PrivateKey) > 0 {
		var err error
		var signer ssh.Signer
		if len(d.Passphrase) > 0 {
			signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(d.PrivateKey), []byte(d.Passphrase))
		} else {
			signer, err = ssh.ParsePrivateKey([]byte(d.PrivateKey))
		}
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
	if err == nil {
		d.clientConnectionError = nil
		go func(d *SFTP) {
			d.clientConnectionError = d.client.Wait()
		}(d)
	}
	return err
}

func (d *SFTP) clientReconnectOnConnectionError() error {
	err := d.clientConnectionError
	if err == nil {
		return nil
	}
	log.Debugf("[sftp] discarding closed sftp connection: %v", err)
	_ = d.client.Close()
	err = d.initClient()
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
