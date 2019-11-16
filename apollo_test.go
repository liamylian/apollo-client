package apollo

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type testConfig struct {
	DbHost     string `apollo_key:"DB_HOST" apollo_callback:"OnDb"`
	DbPort     uint   `apollo_callback:"OnDb" apollo_default:"80"`
	onDbCalled bool
	oldDbHost  string
	oldDbPort  uint
}

func (c *testConfig) OnDb(old testConfig) {
	c.onDbCalled = true
	c.oldDbHost = old.DbHost
	c.oldDbPort = old.DbPort
}

func TestConfigUpdater_ParseConfig(t *testing.T) {
	_, err := newConfigUpdater(testConfig{})
	assert.NotNil(t, err)

	config := &testConfig{}
	updater, err := newConfigUpdater(config)
	assert.Nil(t, err)
	assert.Equal(t, map[string]fieldMeta{
		"DB_HOST": {"DbHost", "DB_HOST", "OnDb", ""},
		"DbPort":  {"DbPort", "DbPort", "OnDb", "80"},
	}, updater.fieldsMeta)
}

func TestConfigUpdater_SetValue(t *testing.T) {
	config := &testConfig{}
	updater, err := newConfigUpdater(config)
	assert.Nil(t, err)

	err = updater.setValue("DbPort", "8080")
	assert.Nil(t, err)
	assert.Equal(t, uint(8080), config.DbPort)
}

func TestConfigUpdater_CallMethod(t *testing.T) {
	config := &testConfig{}
	updater, err := newConfigUpdater(config)
	assert.Nil(t, err)

	err = updater.callMethod("OnDb", *config)
	assert.Nil(t, err)
	assert.True(t, config.onDbCalled)
}

func TestConfigUpdater_Update(t *testing.T) {
	config := &testConfig{}
	updater, err := newConfigUpdater(config)
	assert.Nil(t, err)

	config.DbHost = "127.0.0.1"
	config.DbPort = 8000
	err = updater.Update(map[string]string{
		"DB_HOST": "localhost",
		"DbPort":  "8080",
		"Foo":     "bar",
	})
	assert.Nil(t, err)
	assert.True(t, config.onDbCalled)
	assert.Equal(t, "localhost", config.DbHost)
	assert.Equal(t, uint(8080), config.DbPort)
	assert.Equal(t, "127.0.0.1", config.oldDbHost)
	assert.Equal(t, uint(8000), config.oldDbPort)

}
