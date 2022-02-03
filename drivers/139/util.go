package _39

import (
	"encoding/base64"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

func encodeURIComponent(str string) string {
	r := url.QueryEscape(str)
	r = strings.Replace(r, "+", "%20", -1)
	return r
}

func calSign(body, ts, randStr string) string {
	body = strings.ReplaceAll(body, "\n", "")
	body = strings.ReplaceAll(body, " ", "")
	body = encodeURIComponent(body)
	strs := strings.Split(body, "")
	sort.Strings(strs)
	body = strings.Join(strs, "")
	body = base64.StdEncoding.EncodeToString([]byte(body))
	res := utils.GetMD5Encode(body) + utils.GetMD5Encode(ts+":"+randStr)
	res = strings.ToUpper(utils.GetMD5Encode(res))
	return res
}

func getTime(t string) *time.Time {
	stamp, _ := time.ParseInLocation("20060102150405", t, time.Local)
	return &stamp
}

func isFamily(account *model.Account) bool {
	return account.InternalType == "Family"
}

func unicode(str string) string {
	textQuoted := strconv.QuoteToASCII(str)
	textUnquoted := textQuoted[1 : len(textQuoted)-1]
	return textUnquoted
}

func MergeMap(mObj ...map[string]interface{}) map[string]interface{} {
	newObj := map[string]interface{}{}
	for _, m := range mObj {
		for k, v := range m {
			newObj[k] = v
		}
	}
	return newObj
}

func newJson(data map[string]interface{}, account *model.Account) map[string]interface{} {
	common := map[string]interface{}{
		"catalogType": 3,
		"cloudID":     account.SiteId,
		"cloudType":   1,
		"commonAccountInfo": base.Json{
			"account":     account.Username,
			"accountType": 1,
		},
	}
	return MergeMap(data, common)
}
