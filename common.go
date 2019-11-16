package apollo

import (
	"fmt"
	"net"
	"net/url"
	"reflect"
	"strings"
)

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, a := range addrs {
		if ip4 := toIP4(a); ip4 != nil {
			return ip4.String()
		}
	}
	return ""
}

func toIP4(addr net.Addr) net.IP {
	if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
		return ipnet.IP.To4()
	}
	return nil
}

func httpurl(ipOrAddr string) string {
	if strings.HasPrefix(ipOrAddr, "http://") || strings.HasPrefix(ipOrAddr, "https://") {
		return ipOrAddr
	}

	return fmt.Sprintf("http://%s", ipOrAddr)
}

func notificationURL(conf *Conf, notifications string) string {
	var addr = conf.IP
	if conf.MetaAddr != "" {
		addr = conf.MetaAddr
	}
	return fmt.Sprintf("%s/notifications/v2?appId=%s&cluster=%s&notifications=%s",
		httpurl(addr),
		url.QueryEscape(conf.AppID),
		url.QueryEscape(conf.Cluster),
		url.QueryEscape(notifications))
}

func configURL(conf *Conf, namespace, releaseKey string) string {
	var addr = conf.IP
	if conf.MetaAddr != "" {
		addr = conf.MetaAddr
	}
	return fmt.Sprintf("%s/configs/%s/%s/%s?releaseKey=%s&ip=%s",
		httpurl(addr),
		url.QueryEscape(conf.AppID),
		url.QueryEscape(conf.Cluster),
		url.QueryEscape(namespace),
		releaseKey,
		getLocalIP())
}

func copyStruct(obj interface{}) interface{} {
	if obj == nil {
		return nil
	}

	var isPtr bool
	objTyp := reflect.TypeOf(obj)
	objElemTyp := objTyp
	if objTyp.Kind() == reflect.Struct {
		isPtr = false
	} else if objTyp.Kind() == reflect.Ptr && objTyp.Elem().Kind() == reflect.Struct {
		isPtr = true
		objElemTyp = objTyp.Elem()
	} else {
		return nil
	}

	objElemVal := reflect.ValueOf(obj)
	newObjElemVal := reflect.New(objElemTyp).Elem()
	if isPtr {
		objElemVal = objElemVal.Elem()
	}

	for i := 0; i < objElemTyp.NumField(); i++ {
		inField := objElemVal.Field(i)
		outField := newObjElemVal.Field(i)
		if outField.CanSet() {
			outField.Set(inField)
		}
	}

	if isPtr {
		return newObjElemVal.Addr().Interface()
	} else {
		return newObjElemVal.Interface()
	}
}

func elem(obj interface{}) interface{} {
	val := reflect.ValueOf(obj)
	if val.Kind() != reflect.Ptr {
		return obj
	}

	return val.Elem().Interface()
}

func isLower(s string) bool {
	if len(s) == 0 {
		return false
	}

	c := s[0]
	return 'a' <= c && c <= 'z'
}
