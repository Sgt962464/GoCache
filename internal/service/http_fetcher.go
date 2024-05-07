package service

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type httpFetcher struct {
	//要访问的远程节点地址，比如 http://example.com/_gocache/
	baseURL string
}

var _ Fetcher = (*httpFetcher)(nil)

/*
Fetch 从远程 HTTP 服务器获取数据
  - 构造URL
  - 发起http Get请求
  - 检查响应码
  - 读取响应体
  - 关闭响应体返回数据
*/
func (h *httpFetcher) Fetch(group string, key string) ([]byte, error) {
	u := fmt.Sprintf("%v%v/%v", h.baseURL, url.QueryEscape(group), url.QueryEscape(key))
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", resp.Status)
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body failed: %v", err)
	}
	return bytes, nil
}
