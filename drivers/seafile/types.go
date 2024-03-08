package seafile

import "time"

type AuthTokenResp struct {
	Token string `json:"token"`
}

type RepoItemResp struct {
	Id         string `json:"id"`
	Type       string `json:"type"` // repo, dir, file
	Name       string `json:"name"`
	Size       int64  `json:"size"`
	Modified   int64  `json:"mtime"`
	Permission string `json:"permission"`
}

type LibraryItemResp struct {
	RepoItemResp
	OwnerContactEmail    string `json:"owner_contact_email"`
	OwnerName            string `json:"owner_name"`
	Owner                string `json:"owner"`
	ModifierEmail        string `json:"modifier_email"`
	ModifierContactEmail string `json:"modifier_contact_email"`
	ModifierName         string `json:"modifier_name"`
	Virtual              bool   `json:"virtual"`
	MtimeRelative        string `json:"mtime_relative"`
	Encrypted            bool   `json:"encrypted"`
	Version              int    `json:"version"`
	HeadCommitId         string `json:"head_commit_id"`
	Root                 string `json:"root"`
	Salt                 string `json:"salt"`
	SizeFormatted        string `json:"size_formatted"`
}

type RepoDirItemResp struct {
	RepoItemResp
}

type LibraryInfo struct {
	LibraryItemResp
	decryptedTime    time.Time
	decryptedSuccess bool
}