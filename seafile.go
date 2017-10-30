package seafile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Client struct {
	baseURL string
	repoID  string
	token   string
}

func NewClient(baseURL, repoID string) *Client {
	return &Client{baseURL: baseURL, repoID: repoID}
}

func (s *Client) Request(meth, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(meth, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Token "+s.token)
	return req, nil
}

type TokenResponse struct{ Token string }

func (s *Client) FetchToken(username, password string) error {
	data := url.Values{}
	data.Add("username", username)
	data.Add("password", password)

	resp, err := http.PostForm(s.baseURL+"/api2/auth-token/", data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var tok TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return err
	}

	s.token = tok.Token

	return nil
}

type UploadResponseItem struct {
	Name string `json:"name"`
	ID   string `json:"id"`
	Size uint64 `json:"size"`
}

// UploadFile creates an upload link first and then posts a file to it.
func (s *Client) UploadFile(file string) ([]*UploadResponseItem, error) {
	req, err := s.Request("GET", fmt.Sprintf("%s/api2/repos/%s/upload-link/?p=/", s.baseURL, s.repoID), nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return nil, err
	}

	uploadURL := strings.Trim(buf.String(), `"`)

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buf.Reset()
	w := multipart.NewWriter(&buf)

	part, err := w.CreateFormFile("file", file)
	if err != nil {
		return nil, err
	}

	if _, err = io.Copy(part, f); err != nil {
		return nil, err
	}
	if err := w.WriteField("parent_dir", "/"); err != nil {
		return nil, err
	}
	w.SetBoundary(mime.TypeByExtension(filepath.Ext(file)))
	if err := w.Close(); err != nil {
		return nil, err
	}

	req, err = s.Request("POST", uploadURL+"?ret-json=1", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	if err != nil {
		return nil, err
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var jsonResp []*UploadResponseItem
	return jsonResp, json.NewDecoder(resp.Body).Decode(&jsonResp)
}

type ShareFileResponse struct {
	Username   string      `json:"username"`
	ViewCnt    uint64      `json:"view_cnt"`
	Ctime      string      `json:"ctime"`
	Token      string      `json:"token"`
	RepoID     string      `json:"repo_id"`
	Link       string      `json:"link"`
	ExpireDate interface{} `json:"expire_date"`
	Path       string      `json:"path"`
	IsExpired  bool        `json:"is_expired"`
}

func (s *Client) ShareFile(filename string) (*ShareFileResponse, error) {
	data := url.Values{}
	data.Add("repo_id", s.repoID)
	data.Add("path", filename)

	req, err := s.Request("POST", s.baseURL+"/api/v2.1/share-links/", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var share ShareFileResponse
	return &share, json.NewDecoder(resp.Body).Decode(&share)
}

type ListDirectoryResponseItem struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
	Size uint64 `json:"size"`
}

func (s *Client) ListDirectoryEntries() ([]*ListDirectoryResponseItem, error) {
	data := url.Values{}
	data.Add("t", "f") // type file

	req, err := s.Request("GET", fmt.Sprintf("%s/api2/repos/%s/dir/?%s", s.baseURL, s.repoID, data.Encode()), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result []*ListDirectoryResponseItem
	return result, json.NewDecoder(resp.Body).Decode(&result)
}

type FileDetail struct {
	ID    string `json:"id"`
	Mtime uint64 `json:"mtime"`
	Type  string `json:"type"`
	Name  string `json:"name"`
	Size  uint64 `json:"size"`
}

func (s *Client) FileDetail(name string) (*FileDetail, error) {
	req, err := s.Request("GET", fmt.Sprintf("%s/api2/repos/%s/file/detail/?p=/%s", s.baseURL, s.repoID, name), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var detail FileDetail
	return &detail, json.NewDecoder(resp.Body).Decode(&detail)
}

func (s *Client) DeleteFile(name string) error {
	url := fmt.Sprintf("%s/api2/repos/%s/file/?p=/%s", s.baseURL, s.repoID, name)
	req, err := s.Request("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected response code for DELETE %s to be 200, was %d", url, resp.StatusCode)
	}

	return nil
}
