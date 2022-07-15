package utils

// 配置中心响应,获取配置文件列表
type ConfigResponse struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

type StatusResponse struct {
	Status string `json:"status"`
}

type PushReq struct {
	Project     string           `json:"project"`     //项目名称
	Group       string           `json:"group"`       //组名
	Service     string           `json:"service"`     //服务名称
	Version     string           `json:"version"`     //版本名称
	PushRegions []string         `json:"pushRegions"` //推送区域
	ConfigInfos []*ConfigInfoReq `json:"configInfos"` //配置信息
}

type DeleteReq struct {
	Project     string   `json:"project"`     //项目名称
	Group       string   `json:"group"`       //组名
	Service     string   `json:"service"`     //服务名称
	Version     string   `json:"version"`     //版本名称
	PushRegions []string `json:"pushRegions"` //推送区域
	ConfigName  string   `json:"configName"`  //配置信息
}
type ConfigInfoReq struct {
	FileName string `json:"fileName"` //文件名称
	Content  []byte `json:"content"`  //配置内容
}
