package db

import (
	"github.com/Xhofe/go-cache"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/singleflight"
	"github.com/pkg/errors"
)

var userCache = cache.NewMemCache(cache.WithShards[*model.User](2))
var userG singleflight.Group[*model.User]

func GetAdmin() (*model.User, error) {
	user := model.User{Role: model.ADMIN}
	if err := db.Where(user).Take(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func GetGuest() (*model.User, error) {
	user := model.User{Role: model.GUEST}
	if err := db.Where(user).Take(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByName(username string) (*model.User, error) {
	if username == "" {
		return nil, errors.WithStack(errs.EmptyUsername)
	}
	user, ok := userCache.Get(username)
	if ok {
		return user, nil
	}
	user, err, _ := userG.Do(username, func() (*model.User, error) {
		user := model.User{Username: username}
		if err := db.Where(user).First(&user).Error; err != nil {
			return nil, errors.Wrapf(err, "failed find user")
		}
		userCache.Set(username, &user)
		return &user, nil
	})
	return user, err
}

func GetUserById(id uint) (*model.User, error) {
	var u model.User
	if err := db.First(&u, id).Error; err != nil {
		return nil, errors.Wrapf(err, "failed get old user")
	}
	return &u, nil
}

func CreateUser(u *model.User) error {
	return errors.WithStack(db.Create(u).Error)
}

func UpdateUser(u *model.User) error {
	old, err := GetUserById(u.ID)
	if err != nil {
		return err
	}
	userCache.Del(old.Username)
	return errors.WithStack(db.Save(u).Error)
}

func GetUsers(pageIndex, pageSize int) ([]model.User, int64, error) {
	userDB := db.Model(&model.User{})
	var count int64
	if err := userDB.Count(&count).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get users count")
	}
	var users []model.User
	if err := userDB.Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get find users")
	}
	return users, count, nil
}

func DeleteUserById(id uint) error {
	old, err := GetUserById(id)
	if err != nil {
		return err
	}
	userCache.Del(old.Username)
	return errors.WithStack(db.Delete(&model.User{}, id).Error)
}
