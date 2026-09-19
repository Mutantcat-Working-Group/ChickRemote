use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::path::Path;

#[derive(Clone, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Settings {
    pub id: String,
    pub server: String,
    pub secret: String,
    pub target: String,
    pub tls: bool,
    pub dashboard_port: u16,
}

impl Default for Settings {
    fn default() -> Self {
        Self {
            id: String::new(),
            server: "127.0.0.1:6154".into(),
            secret: String::new(),
            target: String::new(),
            tls: true,
            dashboard_port: 8080,
        }
    }
}

impl Settings {
    pub fn validate(&self) -> Result<(), String> {
        let valid_id = |s: &str| {
            !s.is_empty()
                && s.len() <= 64
                && s.bytes()
                    .all(|b| b.is_ascii_alphanumeric() || b"-_.".contains(&b))
        };
        if !valid_id(&self.id) {
            return Err("Device ID: use 1-64 letters, numbers, dots, dashes or underscores".into());
        }
        if !self.target.is_empty() && (!valid_id(&self.target) || self.target == self.id) {
            return Err("Target must be a different, valid device ID".into());
        }
        let (host, port) = self
            .server
            .rsplit_once(':')
            .ok_or("Relay address must be host:port")?;
        if host.is_empty()
            || host
                .chars()
                .any(|c| c.is_whitespace() || c == '/' || c == '\\')
            || port.parse::<u16>().unwrap_or(0) == 0
        {
            return Err("Relay address must be host:port with a port between 1 and 65535".into());
        }
        if host.contains(':')
            && !(host.starts_with('[')
                && host.ends_with(']')
                && host[1..host.len() - 1]
                    .parse::<std::net::Ipv6Addr>()
                    .is_ok())
        {
            return Err("IPv6 relay addresses must use [address]:port".into());
        }
        if self.secret.len() < 16 || self.secret.len() > 4096 {
            return Err("Shared secret must contain 16-4096 bytes".into());
        }
        if self.dashboard_port < 1024 {
            return Err("Dashboard port must be between 1024 and 65535".into());
        }
        Ok(())
    }

    pub fn engine_config(&self, dir: &Path) -> Value {
        let rules = if self.target.is_empty() {
            vec![]
        } else {
            vec![
                json!({"name":"desktop", "target":self.target, "type":"vnc", "local_addr":"127.0.0.1", "fps":15}),
            ]
        };
        json!({"id":self.id, "server":self.server, "secret":self.secret,
            "ssl":{"enabled":self.tls,"insecure":false},
            "dashboard":{"enabled":true,"listen":"127.0.0.1","port":self.dashboard_port},
            "link":{"read_timeout":"5s","write_timeout":"5s"},
            "log":{"dir":dir.join("logs"),"size":"10M","rotate":3},
            "codedir":dir.join("code"), "rules":rules})
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn valid() -> Settings {
        Settings {
            id: "local".into(),
            server: "localhost:6154".into(),
            secret: "a sufficiently long secret".into(),
            target: "remote".into(),
            tls: true,
            dashboard_port: 8080,
        }
    }
    #[test]
    fn validates_required_fields_and_ports() {
        assert!(valid().validate().is_ok());
        for server in [
            "localhost",
            "http://localhost:6154",
            "localhost:0",
            "localhost:99999",
        ] {
            let mut config = valid();
            config.server = server.into();
            assert!(config.validate().is_err(), "{server}");
        }
        let mut config = valid();
        config.secret = "short".into();
        assert!(config.validate().is_err());
        config = valid();
        config.target = config.id.clone();
        assert!(config.validate().is_err());
    }
    #[test]
    fn generated_config_quotes_values_and_limits_listeners() {
        let mut config = valid();
        config.secret = "secret\nwith: \"quotes\"".into();
        let value = config.engine_config(std::path::Path::new("/tmp/app data"));
        assert_eq!(value["secret"], config.secret);
        assert_eq!(value["dashboard"]["listen"], "127.0.0.1");
        assert_eq!(value["rules"][0]["local_addr"], "127.0.0.1");
        assert_eq!(value["ssl"]["insecure"], false);
        config.target.clear();
        assert!(config.engine_config(std::path::Path::new("/tmp"))["rules"]
            .as_array()
            .unwrap()
            .is_empty());
    }
}
