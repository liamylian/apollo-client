package apollo_client

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type CallBack struct {
	SuccessCallBack   func([]byte) (interface{}, error)
	NotModifyCallBack func() error
}

type ConnectConfig struct {
	//设置到http.client中timeout字段
	Timeout time.Duration
	//连接接口的uri
	Uri string
}

func request(requestUrl string, connectionConfig *ConnectConfig, callBack *CallBack) (interface{}, error) {
	client := &http.Client{}
	//如有设置自定义超时时间即使用
	if connectionConfig != nil && connectionConfig.Timeout != 0 {
		client.Timeout = connectionConfig.Timeout
	} else {
		client.Timeout = connectTimeout
	}

	retry := 0
	var responseBody []byte
	var err error
	var res *http.Response
	for {
		retry++

		if retry > maxRetries {
			break
		}

		res, err = client.Get(requestUrl)

		if res == nil || err != nil {
			logger.Error("connect apollo server fail,url:%s,error:%s", requestUrl, err)
			continue
		}

		//not modified break
		switch res.StatusCode {
		case http.StatusOK:
			responseBody, err = ioutil.ReadAll(res.Body)
			if err != nil {
				logger.Error("connect apollo server fail,url:%s,error:", requestUrl, err)
				continue
			}

			if callBack != nil && callBack.SuccessCallBack != nil {
				return callBack.SuccessCallBack(responseBody)
			} else {
				return nil, nil
			}
		case http.StatusNotModified:
			logger.Info("config not modified:", err)
			if callBack != nil && callBack.NotModifyCallBack != nil {
				return nil, callBack.NotModifyCallBack()
			} else {
				return nil, nil
			}
		default:
			logger.Error("connect apollo server fail,url:%s,error:%s", requestUrl, err)
			if res != nil {
				logger.Error("connect apollo server fail,url:%s,statusCode:%s", requestUrl, res.StatusCode)
			}
			err = errors.New("connect apollo server fail")
			// if error then sleep
			time.Sleep(onErrorRetryInterval)
			continue
		}
	}

	logger.Error("over max retry still error,error:", err)
	if err != nil {
		err = errors.New("over max retry still error")
	}
	return nil, err
}

func requestRecovery(appConfig *AppConfig,
	connectConfig *ConnectConfig,
	callBack *CallBack) (interface{}, error) {
	format := "%s%s"
	var err error
	var response interface{}

	for {
		host := appConfig.selectHost()
		if host == "" {
			return nil, err
		}

		requestUrl := fmt.Sprintf(format, host, connectConfig.Uri)
		response, err = request(requestUrl, connectConfig, callBack)
		if err == nil {
			return response, err
		}

		setDownNode(host)
	}

	return nil, errors.New("try all nodes still error")
}
