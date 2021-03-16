package utils

import "github.com/Xhofe/alist/conf"

func GetDriveByName(name string) *conf.Drive {
	for i, drive := range conf.Conf.AliDrive.Drives{
		if drive.Name == name {
			return &conf.Conf.AliDrive.Drives[i]
		}
	}
	return nil
}

func GetNames() []string {
	res := make([]string, 0)
	for _, drive := range conf.Conf.AliDrive.Drives{
		if !drive.Hide {
			res = append(res, drive.Name)
		}
	}
	return res
}