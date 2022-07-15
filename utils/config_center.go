package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type RsaEncrypter struct {
	PublicKey string
	block     *pem.Block
	publick   *rsa.PublicKey
}

func NewRsaEncrypter(publicKey string) (*RsaEncrypter, error) {
	re := &RsaEncrypter{}
	if len(publicKey) > 0 {
		if err := re.parsePublicKey(publicKey); err != nil {
			return nil, err
		}
	}
	return re, nil
}

func (d *RsaEncrypter) ParsePublicKeyByBytes(bbs []byte) error {
	pubInterface, err := x509.ParsePKIXPublicKey(bbs)
	if err != nil {
		return err
	}
	d.publick = pubInterface.(*rsa.PublicKey)
	return nil
}

//parse from pem file
func (d *RsaEncrypter) parsePublicKey(publicKey string) error {
	block, _ := pem.Decode([]byte(publicKey))
	if block == nil {
		return errors.New("public key error")
	}
	// 解析公钥
	return d.ParsePublicKeyByBytes(block.Bytes)
}

func (d *RsaEncrypter) Encrypt(data []byte) ([]byte, error) {
	return rsa.EncryptPKCS1v15(rand.Reader, d.publick, data)
}

func (d *RsaEncrypter) EncryptJson(o interface{}) ([]byte, error) {
	b, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}
	return rsa.EncryptPKCS1v15(rand.Reader, d.publick, b)
}

type CenterService struct {
	CenterAddr string //接口地址
	AppId      string //AppId
	Username   string //用户名
	Password   string //密码
	encrypter  *RsaEncrypter
}

func NewCenterService(centerAddr string, appId string, username string, //
	password string, publicKey string) (*CenterService, error) {
	encrypter := &RsaEncrypter{}
	pk, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return nil, err
	}
	err = encrypter.ParsePublicKeyByBytes(pk)
	if err != nil {
		return nil, err
	}
	service := &CenterService{
		CenterAddr: centerAddr,
		AppId:      appId,
		Username:   username,
		Password:   password,
	}
	service.encrypter = encrypter
	return service, nil
}

func (s *CenterService) genAuthHeader() (map[string]string, error) {
	authProp := map[string]interface{}{
		"name":      s.Username,
		"pwd":       s.Password,
		"timestamp": time.Now().UnixNano() / 1000 / 1000,
	}
	authData, err := json.Marshal(authProp)
	if err != nil {
		return nil, err
	}
	encryptedData, err := s.encrypter.Encrypt(authData)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"token":  base64.StdEncoding.EncodeToString(encryptedData),
		"app_id": s.AppId,
	}, nil
}
func (s *CenterService) PushConfig() (error, *ConfigResponse) {
	authHeader, err := s.genAuthHeader()
	if err != nil {
		return err, nil
	}
	pushRegions := []string{"备用", "companion-dev"}
	pushReq := &PushReq{
		Project:     "test",
		Group:       "test",
		Service:     "test",
		Version:     "test",
		PushRegions: pushRegions,
		ConfigInfos: []*ConfigInfoReq{&ConfigInfoReq{
			FileName: "squeezer_log.xml",
			Content:  []byte("hello word"),
		}},
	}
	pushResp := &ConfigResponse{}
	resp := HttpPost(fmt.Sprintf("%s/config/push", s.CenterAddr), authHeader, pushReq, pushResp)
	if resp.GetError() != nil {
		return resp.GetError(), nil
	}
	return nil, pushResp
}

//删除配置
func (s *CenterService) DeleteConfig() (error, *ConfigResponse) {
	authHeader, err := s.genAuthHeader()
	if err != nil {
		return err, nil
	}
	deleteRegions := []string{"备用", "companion-dev"}
	deleteReq := &DeleteReq{
		Project:     "test",
		Group:       "test",
		Service:     "test",
		Version:     "test",
		PushRegions: deleteRegions,
		ConfigName:  "squeezer_log.xml",
	}
	deleteResp := &ConfigResponse{}
	resp := HttpPost(fmt.Sprintf("%s/config/delete", s.CenterAddr), authHeader, deleteReq, deleteResp)
	if resp.GetError() != nil {
		return resp.GetError(), nil
	}
	return nil, deleteResp
}

func (s *CenterService) GetAndPushConfig(proj, cluster, service, version, cfgFileName string, regions []string) error {
	authHeader, err := s.genAuthHeader()
	if err != nil {
		return err
	}
	downResp := &ConfigResponse{}
	reqUrl := fmt.Sprintf("%s/config/download?project=%s&cluster=%s&service=%s&version=%s&configName=%s",
		s.CenterAddr,
		proj,        //项目名称
		cluster,     //集群名称
		service,     //服务名称
		version,     //服务版本
		cfgFileName, //配置文件名称
	)
	resp := HttpGet(reqUrl, authHeader, downResp)
	if resp.GetError() != nil {
		return resp.GetError()
	}
	if downResp.Data == nil {
		fmt.Println("Not found this, ", cfgFileName)
		return errors.New(fmt.Sprintf("Not found this, cfg %s ", cfgFileName))
	}
	data := downResp.Data.(map[string]interface{})
	var content string
	if data == nil {
		fmt.Println("Not found this, ", cfgFileName)
		return errors.New(fmt.Sprintf("Not found this, cfg %s ", cfgFileName))

	}
	content = data["content"].(string)
	pushReq := &PushReq{
		Project:     proj,
		Group:       cluster,
		Service:     service,
		Version:     version,
		PushRegions: regions,
		ConfigInfos: []*ConfigInfoReq{&ConfigInfoReq{
			FileName: cfgFileName,
			Content:  []byte(content),
		}},
	}
	pushResp := &ConfigResponse{}
	presp := HttpPost(fmt.Sprintf("%s/config/push", s.CenterAddr), authHeader, pushReq, pushResp)
	if presp.GetError() != nil {
		return presp.GetError()
	}
	if pushResp.Code == 0 {
		return nil
	}
	return errors.New(pushResp.Message)

}

func (s *CenterService) DownLoadFile() (error, *ConfigResponse) {
	authHeader, err := s.genAuthHeader()
	if err != nil {
		return err, nil
	}
	pushResp := &ConfigResponse{}
	reqUrl := fmt.Sprintf("%s/config/download?project=%s&cluster=%s&service=%s&version=%s&configName=%s",
		s.CenterAddr,
		"guiderAllService", //项目名称
		"gas",              //集群名称
		"seve",             //服务名称
		"1.0.0",            //服务版本
		"seve.toml",        //配置文件名称
	)
	resp := HttpGet(reqUrl, authHeader, pushResp)
	if resp.GetError() != nil {
		return resp.GetError(), nil
	}

	return nil, pushResp
}

func (s *CenterService) IsAlive() bool {
	authHeader, err := s.genAuthHeader()
	if err != nil {
		return false
	}
	reqUrl := fmt.Sprintf("%s/%s",
		s.CenterAddr, "health",
	)
	statusResp := &StatusResponse{}
	resp := HttpGet(reqUrl, authHeader, statusResp)
	if strings.ToLower(statusResp.Status) == "up" && resp.StatusCode == http.StatusOK {
		return true
	}
	return false
}
