package template

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"os"
	"path"
	"sync"
)

var clientsMap = struct {
	sync.Mutex
	clients map[string]*Client
}{clients: make(map[string]*Client)}

func GetClient(account *model.Account) (*Client, error) {
	clientsMap.Lock()
	defer clientsMap.Unlock()
	if v, ok := clientsMap.clients[account.Name]; ok {
		return v, nil
	}
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", account.SiteUrl, account.Limit), &ssh.ClientConfig{
		User:            account.Username,
		Auth:            []ssh.AuthMethod{ssh.Password(account.Password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		return nil, err
	}
	client, err := sftp.NewClient(conn)
	if err != nil {
		return nil, err
	}
	c := &Client{client}
	clientsMap.clients[account.Name] = c
	return c, nil
}

type Client struct {
	*sftp.Client
}

func (client *Client) Files(remotePath string) ([]os.FileInfo, error) {
	return client.ReadDir(remotePath)
}

func (client *Client) remove(remotePath string) error {
	f, err := client.Stat(remotePath)
	if err != nil {
		return nil
	}
	if f.IsDir() {
		return client.removeDirectory(remotePath)
	} else {
		return client.removeFile(remotePath)
	}
}

func (client *Client) removeDirectory(remotePath string) error {
	//打不开,说明要么文件路径错误了,要么是第一次部署
	remoteFiles, err := client.ReadDir(remotePath)
	if err != nil {
		return err
	}
	for _, backupDir := range remoteFiles {
		remoteFilePath := path.Join(remotePath, backupDir.Name())
		if backupDir.IsDir() {
			err := client.removeDirectory(remoteFilePath)
			if err != nil {
				return err
			}
		} else {
			err := client.Remove(path.Join(remoteFilePath))
			if err != nil {
				return err
			}
		}
	}
	return client.RemoveDirectory(remotePath)
}

func (client *Client) removeFile(remotePath string) error {
	return client.Remove(utils.Join(remotePath))
}

func (driver SFTP) formatFile(f os.FileInfo) model.File {
	t := f.ModTime()
	file := model.File{
		//Id:        f.Id,
		Name:      f.Name(),
		Size:      f.Size(),
		Driver:    driver.Config().Name,
		UpdatedAt: &t,
	}
	if f.IsDir() {
		file.Type = conf.FOLDER
	} else {
		file.Type = utils.GetFileType(path.Ext(f.Name()))
	}
	return file
}

func init() {
	base.RegisterDriver(&SFTP{})
}
