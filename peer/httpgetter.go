package peer

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"io"
	"jie_cache/pb"
	"log"
	"net/http"
	"net/url"
)

type HttpGetter struct {
	baseUrl string
}

func NewHttpGetter(host string) *HttpGetter {
	return &HttpGetter{baseUrl: host}
}

func (h *HttpGetter) Get(req *pb.Request, resp *pb.Response) error {
	u, err := url.Parse("http://" + h.baseUrl)
	if err != nil {
		log.Fatal(err)
	}
	q := u.Query()
	q.Set(`group`, req.Group)
	q.Set(`key`, req.Key)
	u.RawQuery = q.Encode()
	log.Printf("[HttpGetter] url: %s \n", u.String())
	res, err := http.Get(u.String())
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %v", err)
	}

	// 解码数据到resp
	if err := proto.Unmarshal(bytes, resp); err != nil {
		return err
	}
	return nil
}

var _ PeerGetter = (*HttpGetter)(nil)
