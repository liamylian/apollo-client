package apollo

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
	"time"
)

func TestLocalIp(t *testing.T) {
	ip := getLocalIP()
	if ip == "" {
		t.FailNow()
	}
}

func TestNotificationURL(t *testing.T) {
	target := notificationURL(
		&Conf{
			IP:       "127.0.0.1:8080",
			MetaAddr: "http://127.0.0.1:8080",
			AppID:    "SampleApp",
			Cluster:  "default",
		}, "")
	_, err := url.Parse(target)
	if err != nil {
		t.Error(err)
	}
}

func TestConfigURL(t *testing.T) {
	target := configURL(
		&Conf{
			MetaAddr: "127.0.0.1:8080",
			IP:       "127.0.0.1:8080",
			AppID:    "SampleApp",
			Cluster:  "default",
		}, "application", "")
	_, err := url.Parse(target)
	if err != nil {
		t.Error(err)
	}
}

func TestCopyStruct(t *testing.T) {
	type st struct {
		Foo string
		Bar time.Time
		foo int
	}

	t.Log("copy struct")
	origin := st{"foo", time.Now(), 11}
	copied := copyStruct(origin).(st)
	assert.Equal(t, origin.Foo, copied.Foo)
	assert.Equal(t, origin.Bar, copied.Bar)

	t.Log("copy struct ptr")
	originP := &st{"foo", time.Now(), 11}
	copiedP := copyStruct(originP).(*st)
	assert.Equal(t, originP.Foo, copiedP.Foo)
	assert.Equal(t, originP.Bar, copiedP.Bar)
}
