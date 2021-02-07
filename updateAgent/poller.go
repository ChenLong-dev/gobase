package upgradeAgent

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ChenLong-dev/gobase/mbase/mutils"
	"github.com/ChenLong-dev/gobase/mlog"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Poller struct {
	Host 				string
	PollPath 			string
	DownloadedPath		string
	UpgradedPath 		string
	NextIntval 			int
	Client 				*http.Client
	getPollRequest 		func() PollRequest
	upgradedCb 			func(p *PackageInfo, result error)
	WorkDir 			string
}
func (this *Poller) rpc(path string, req interface{}, res interface{}) (err error) {
	mlog.Tracef("path=%s,req=%v", path, req)
	defer func() {mlog.Tracef("res=%v,err=%v", res, err)} ()

	reqBody := mutils.JsonPrint(req)
	resp, perr := this.Client.Post("https://"+this.Host+path, "application/json", strings.NewReader(reqBody))
	if perr != nil {
		mlog.Debugf("post https://%s%s %s error:%v", this.Host, path, reqBody, perr)
		resp, perr = this.Client.Post("http://"+this.Host+path, "application/json", strings.NewReader(reqBody))
	}
	if perr != nil {
		return perr
	}

	decoder := json.NewDecoder(resp.Body)
	return decoder.Decode(res)
}
func (this *Poller) poll(req PollRequest) (p *PackageInfo, err error) {
	mlog.Tracef("req=%v", req)
	defer func() {mlog.Tracef("p=%v,err=%v", p, err)} ()

	type PollResponse struct {
		Code 			int				`json:"code"`
		Message 		string			`json:"message"`
		P 				*PackageInfo	`json:"packege,omitempty"`
		NextInterval	int 			`json:"nextInterval"`
	}

	res := &PollResponse{}

	if err := this.rpc(this.PollPath, req, res); err != nil {
		return nil, err
	}
	mlog.Tracef("res=%v", res)

	if res.NextInterval > 0 && res.NextInterval < 60*60*24*7 {
		this.NextIntval = res.NextInterval
	}
	if res.Code != 0 || res.P == nil {
		return nil, nil
	}
	if err = this.checkPackage(res.P); err != nil {
		return nil, err
	}

	return res.P, nil
}
func (this *Poller) checkPackage(p *PackageInfo) error {
	if p == nil {
		return nil
	}
	if p.Ver == "" || p.DownloadUrl == "" || p.SHA256 == "" || p.Size <= 0 {
		return fmt.Errorf("%v params error", p)
	}
	signData := []byte(p.Ver + p.DownloadUrl + p.SHA256 + fmt.Sprint(p.Size) + p.UpgradeCommand)
	if !VerifyWithServer(signData, p.ServerSignature) {
		return fmt.Errorf("server signature error")
	}
	if !VerifyWithDeveloper(signData, p.DeveloperSignature) {
		return fmt.Errorf("developer signature error")
	}
	return nil
}
func (this *Poller) download(p *PackageInfo, packagePath string) (err error) {
	mlog.Infof("now start download package(%v) to %s...", mutils.JsonPrint(p), packagePath)

	var proxy func(*http.Request) (*url.URL, error) = nil
	if p.Proxy != "" {
		proxy = func(r *http.Request) (*url.URL, error) {
			if r.URL.Scheme == "https" {
				return url.Parse("https://" + p.Proxy)
			}
			return url.Parse("http://" + p.Proxy)
		}
	}
	transport := &http.Transport{Proxy: proxy}
	client := &http.Client{Transport: transport, Timeout: time.Hour}

	resp, rerr := client.Get(p.DownloadUrl)
	if rerr != nil && p.Proxy != "" {
		client.Transport = nil
		resp, rerr = client.Get(p.DownloadUrl)
	}
	if rerr != nil {
		return rerr
	}
	defer resp.Body.Close()

	if resp.ContentLength != -1 && resp.ContentLength != int64(p.Size) {
		return fmt.Errorf("resp.ContentLength(%d) not match package size(%d)", resp.ContentLength, p.Size)
	}

	f, ferr := os.Create(packagePath)
	if ferr != nil {
		return ferr
	}
	defer f.Close()

	hash := sha256.New()
	body := io.TeeReader(resp.Body, hash)
	wn, werr := io.CopyN(f, body, int64(p.Size))
	if wn != int64(p.Size) {
		return fmt.Errorf("down load size(%d) not match package size(%d) error:%v", wn, p.Size, werr)
	}
	if h := hex.EncodeToString(hash.Sum(nil)); h != p.SHA256 {
		return fmt.Errorf("calc hash(%s) not match package hash(%s)", h, p.SHA256)
	}

	return nil
}
func (this *Poller) statusReport(pr PollRequest, urlpath string, ver string, err error) {
	mlog.Tracef("pr=%v,urlpath=%s,ver=%s,err=%v", pr, urlpath, ver, err)

	sr := pr.NewStatusReport(ver, err)
	res := &Response{}

	this.rpc(urlpath, sr, res)
}
func (this *Poller) downloadedReport(pr PollRequest, ver string, err error) {
	mlog.Infof("pr=%v,ver=%s,err=%v", pr, ver, err)

	this.statusReport(pr, this.DownloadedPath, ver, err)
}
func (this *Poller) upgradeReport(pr PollRequest, ver string, err error) {
	mlog.Infof("pr=%v,ver=%s,err=%v", pr, ver, err)

	this.statusReport(pr, this.UpgradedPath, ver, err)
}
func (this *Poller) Run() {
	mlog.Tracef("")

	for {
		//	1. poll
		pr := this.getPollRequest()
		p, perr := this.poll(pr)
		mlog.Tracef("p=%v,perr=%v", p, perr)
		if p != nil {
			//	2. download
			upgradeDir := mkUpgradeDir(this.WorkDir, p)
			packagePath := createUpgradePackagePath(upgradeDir)

			derr := this.download(p, packagePath)
			this.downloadedReport(pr, p.Ver, derr)

			//	3.upgrade
			if derr == nil {
				err := ExecUpgrade(packagePath, p.UpgradeCommand)
				if this.upgradedCb != nil {
					this.upgradedCb(p, err)
				}
				this.upgradeReport(pr, p.Ver, err)
			}

			//	4.clean
			rmUpgradeDir(upgradeDir)
		}

		time.Sleep(time.Second* time.Duration(this.NextIntval))
	}
}

type InitPollerParams struct {
	ServerHost 				string		//	升级服务器地址，默认 upgrade.sase.sangfor.com.cn
	PollPath				string		//	轮询包的URL路径，如 /pop/vac/poll
	DownloadedPath 			string		//	下载结束的URL报告路径，如 /pop/vac/downloaded
	UpgradedPath 			string		//	升级结束的URL报告路径, 如 /pop/vac/upgraded
	DefaultPollIntval 		int			//	默认轮询包的时间间隔，单位秒，该值会受服务端控制改变，这里只是设置初始值
	EnableVerifyCert		bool 		//	是否校验服务端证书
	HttpTimeout 			int			//	服务端响应的超时时间
	GetPollRequestCb 		func() PollRequest
	UpgradedCb 				func(p *PackageInfo, result error)
	WorkDir 				string
}
func NewPoller(params *InitPollerParams) *Poller {
	poller := &Poller{
		Host: params.ServerHost,
		PollPath: params.PollPath,
		DownloadedPath: params.DownloadedPath,
		UpgradedPath: params.UpgradedPath,
		NextIntval: params.DefaultPollIntval,
		getPollRequest: params.GetPollRequestCb,
		upgradedCb: params.UpgradedCb,
		WorkDir: params.WorkDir,
	}
	if poller.Host == "" {
		poller.Host = "upgrade.sase.sangfor.com.cn"
	}
	if poller.NextIntval <= 0 {
		poller.NextIntval = 3*60
	}
	if poller.WorkDir == "" {
		poller.WorkDir = "/tmp/upgradeAgent/"
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !params.EnableVerifyCert},
	}
	poller.Client = &http.Client{Timeout: time.Second*time.Duration(params.HttpTimeout), Transport: tr}

	return poller
}
/*
func NewPoller() *Poller {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Timeout: time.Second*20, Transport: tr}
	return &Poller{
		Host: "upgrade.isspsec.com",
		PollPath: "/pop/vac/poll",
		DownloadedPath: "/pop/vac/downloaded",
		UpgradedPath: "/pop/vac/upgraded",
		NextIntval: gConfigure.DefaultPollIntval,
		Client: client,
	}
}*/

//var gPoller = NewPoller()
	