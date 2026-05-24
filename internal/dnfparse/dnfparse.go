// Package dnfparse provides a trivial parser for DNF/Yum repository files (.repo).
// It supports: id, name, baseurl, gpgcheck, repo_gpgcheck, gpgkey, sslcacert, sslclientcert, sslclientkey, sslverify.
package dnfparse

import (
	"bufio"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// Repo holds the parsed repository configuration for a single [section].
type Repo struct {
	Id            string
	Name          string
	BaseURLs      []string
	CheckGPG      *bool
	CheckRepoGPG  *bool
	GPGKeys       []string
	SSLCACert     string
	SSLClientCert string
	SSLClientKey  string
	// IgnoreSSL corresponds to sslverify=0 (true = ignore SSL verification). sslverify=1 means false.
	IgnoreSSL *bool
}

// ParseFile reads a DNF repo file and returns all repository sections
// that have at least one baseurl. Sections without baseurl are skipped.
func ParseFile(path string) ([]Repo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open repo file: %w", err)
	}
	defer f.Close()
	return Parse(f, path)
}

// Parse reads from r (e.g. an open .repo file) and returns all repository
// sections that have at least one baseurl. The name argument is used in errors.
func Parse(r io.Reader, name string) ([]Repo, error) {
	var repos []Repo
	var cur *Repo
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Comment or empty
		if line == "" || line[0] == '#' || line[0] == ';' {
			continue
		}
		if len(line) >= 2 && line[0] == '[' && line[len(line)-1] == ']' {
			// Flush current section if it has baseurls
			if cur != nil && len(cur.BaseURLs) > 0 {
				repos = append(repos, *cur)
			}
			id := strings.TrimSpace(line[1 : len(line)-1])
			cur = &Repo{Id: id}
			continue
		}
		eq := strings.Index(line, "=")
		if eq <= 0 || cur == nil {
			continue
		}
		key := strings.TrimSpace(strings.ToLower(line[:eq]))
		val := strings.TrimSpace(line[eq+1:])
		// Remove optional quotes around value
		if len(val) >= 2 && (val[0] == '"' && val[len(val)-1] == '"' || val[0] == '\'' && val[len(val)-1] == '\'') {
			val = val[1 : len(val)-1]
		}
		switch key {
		case "name":
			cur.Name = val
		case "baseurl":
			cur.BaseURLs = append(cur.BaseURLs, val)
		case "gpgcheck":
			b := parseBool(val)
			cur.CheckGPG = &b
		case "repo_gpgcheck":
			b := parseBool(val)
			cur.CheckRepoGPG = &b
		case "gpgkey":
			cur.GPGKeys = append(cur.GPGKeys, val)
		case "sslcacert":
			cur.SSLCACert = val
		case "sslclientcert":
			cur.SSLClientCert = val
		case "sslclientkey":
			cur.SSLClientKey = val
		case "sslverify":
			// sslverify=1 -> verify (IgnoreSSL=false), sslverify=0 -> ignore (IgnoreSSL=true)
			verify := parseBool(val)
			ignoreSSL := !verify
			cur.IgnoreSSL = &ignoreSSL
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", name, err)
	}
	if cur != nil && len(cur.BaseURLs) > 0 {
		repos = append(repos, *cur)
	}
	return repos, nil
}

func parseBool(s string) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "1" || s == "yes" || s == "true" {
		return true
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err == nil && n != 0 {
		return true
	}
	return false
}

// GPGKeyContents resolves GPG key URIs and returns the concatenated contents
// of the key files. Only file:// URIs are supported; any other scheme returns an error.
func (r *Repo) GPGKeyContents() ([]byte, error) {
	var out []byte
	for i, uri := range r.GPGKeys {
		u, err := url.Parse(uri)
		if err != nil {
			return nil, fmt.Errorf("gpgkey[%d] invalid URI %q: %w", i, uri, err)
		}
		if u.Scheme != "file" {
			return nil, fmt.Errorf("gpgkey[%d] only file:// URIs are supported, got %q", i, uri)
		}
		path := u.Path
		if path == "" {
			return nil, fmt.Errorf("gpgkey[%d] file URI has no path: %q", i, uri)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("gpgkey[%d] read %q: %w", i, path, err)
		}
		if len(out) > 0 {
			out = append(out, '\n')
		}
		out = append(out, data...)
	}
	return out, nil
}
