package service

import (
	"fmt"
	"gocache/internal/service/consistenthash"
	"gocache/utils/logger"
	"net/http"
	"strings"
	"sync"
)

var _ Picker = (*HTTPPool)(nil)

/*
因为还有其他服务可能托管在主机上，所以添加额外的路径是一个好习惯，
大多数网站都有api接口，这些接口通常以api为前缀；
*/
const (
	defaultBasePath = "/_gocache/"
	apiServerAddr   = "127.0.0.1:9999"
)

type HTTPPool struct {
	address      string
	basePath     string
	mu           sync.Mutex
	peers        *consistenthash.ConsistentHash //用于根据特定键选择节点
	httpFetchers map[string]*httpFetcher
}

func NewHTTPPool(address string) *HTTPPool {
	return &HTTPPool{
		address:  address,
		basePath: defaultBasePath,
	}
}

func (hp *HTTPPool) Log(format string, args ...interface{}) {
	logger.LogrusObj.Infof("[Server %s] %s", hp.address, fmt.Sprintf(format, args...))
}

func (hp *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, hp.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}

	//日志记录请求方法和请求路径
	hp.Log("%s %s", r.Method, r.URL.Path)

	//  prefix/group/key
	parts := strings.SplitN(r.URL.Path[len(hp.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request format, expected prefix/group/key", http.StatusBadRequest)
		return
	}

	groupName, key := parts[0], parts[1]

	g := GetGroup(groupName)
	if g == nil {
		http.Error(w, "no such group"+groupName, http.StatusBadRequest)
		return
	}
	view, err := g.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}

/*
Pick  根据具体键，选择节点，返回该节点对应的HTTP客户端。
*/
func (hp *HTTPPool) Pick(key string) (Fetcher, bool) {
	hp.mu.Lock()
	defer hp.mu.Unlock()
	// 获取key的peer地址
	peerAddress := hp.peers.Get(key)
	if peerAddress == hp.address {
		return nil, false
	}
	logger.LogrusObj.Infof("[dispatcher peer %s] pick remote peer: %s", apiServerAddr, peerAddress)
	return hp.httpFetchers[peerAddress], true
}

/*
UpdatePeers 更新 HTTPPool 中的 peer 列表和对应的 HTTP fetcher 映射
*/
func (hp *HTTPPool) UpdatePeers(peers ...string) {
	hp.mu.Lock()
	defer hp.mu.Unlock()
	// 创建新的哈希环
	hp.peers = consistenthash.NewConsistentHash(defaultReplicas, nil)
	hp.peers.Add(peers)
	// 初始化httpFetchers映射
	hp.httpFetchers = make(map[string]*httpFetcher, len(peers))
	// 为每peer创建httpFetcher
	for _, peer := range peers {
		hp.httpFetchers[peer] = &httpFetcher{
			// such "http://10.0.0.1:9999/_gocache/"
			baseURL: peer + hp.basePath,
		}
	}
}

/*
NOTE:
- application/octet-stream 是一种通用的二进制数据类型，用于传输任意类型的二进制数据，没有特定的结构或者格式，可以用于传输图片、音频、视频、压缩文件等任意二进制数据。
- application/json ：用于传输 JSON（Javascript Object Notation）格式的数据，JSON 是一种轻量级的数据交换格式，常用于 Web 应用程序之间的数据传输。
- application/xml：用于传输 XML（eXtensible Markup Language）格式的数据，XML 是一种标记语言，常用于数据的结构化表示和交换。
- text/plain：用于传输纯文本数据，没有特定的格式或者结构，可以用于传输普通文本、日志文件等。
- multipart/form-data：用于在 HTML 表单中上传文件或者二进制数据，允许将表单数据和文件一起传输。
- image/jpeg、image/png、image/gif：用于传输图片数据，分别对应 JPEG、PNG 和 GIF 格式的图片。
- audio/mpeg、audio/wav：用于传输音频数据，分别对应 MPEG 和 WAV 格式的音频
- video/map、video/quicktime：用于传输视频数据，分别对应 MAP4 和 Quicktime 格式的视频。
*/
