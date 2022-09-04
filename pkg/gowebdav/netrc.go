package gowebdav

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
)

func parseLine(s string) (login, pass string) {
	fields := strings.Fields(s)
	for i, f := range fields {
		if f == "login" {
			login = fields[i+1]
		}
		if f == "password" {
			pass = fields[i+1]
		}
	}
	return login, pass
}

// ReadConfig reads login and password configuration from ~/.netrc
// machine foo.com login username password 123456
func ReadConfig(uri, netrc string) (string, string) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", ""
	}

	file, err := os.Open(netrc)
	if err != nil {
		return "", ""
	}
	defer file.Close()

	re := fmt.Sprintf(`^.*machine %s.*$`, u.Host)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		s := scanner.Text()

		matched, err := regexp.MatchString(re, s)
		if err != nil {
			return "", ""
		}
		if matched {
			return parseLine(s)
		}
	}

	return "", ""
}
