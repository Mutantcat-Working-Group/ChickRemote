package global

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDesktopJSONConfiguration(t *testing.T) {
	dir := t.TempDir()
	secret := "long secret\nwith: \"quotes\""
	data, err := json.Marshal(map[string]interface{}{
		"id": "desktop-test", "server": "localhost:6154", "secret": secret,
		"ssl":       map[string]interface{}{"enabled": true, "insecure": false},
		"log":       map[string]interface{}{"dir": dir, "size": "10M", "rotate": 3},
		"codedir":   dir,
		"dashboard": map[string]interface{}{"enabled": true, "listen": "127.0.0.1", "port": 8080},
		"rules":     []map[string]interface{}{{"name": "desktop", "type": "vnc", "target": "remote", "local_addr": "127.0.0.1", "fps": 15}},
	})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "client.yaml")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	config := LoadConf(path)
	if config.ID != "desktop-test" || !config.UseSSL || config.SSLInsecure || config.DashboardListen != "127.0.0.1" || config.Rules[0].Target != "remote" {
		t.Fatalf("desktop configuration did not round trip: %+v", config)
	}
}
