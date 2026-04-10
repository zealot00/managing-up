// Managing-Up MCP SDK for Rust

use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::env;
use std::io::{self, BufRead, Write};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MCPServerConfig {
    pub name: String,
    pub version: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub description: Option<String>,
    pub transport_type: TransportType,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub url: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub command: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub args: Option<Vec<String>>,
}

#[derive(Debug, Clone, Copy, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum TransportType {
    #[serde(rename = "stdio")]
    Stdio,
    #[serde(rename = "http")]
    Http,
    #[serde(rename = "sse")]
    Sse,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Metrics {
    pub requests_total: u64,
    pub requests_success: u64,
    pub requests_error: u64,
    pub latency_seconds: f64,
    pub tool_calls: HashMap<String, u64>,
}

impl Default for Metrics {
    fn default() -> Self {
        Self {
            requests_total: 0,
            requests_success: 0,
            requests_error: 0,
            latency_seconds: 0.0,
            tool_calls: HashMap::new(),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct JsonRpcRequest {
    pub jsonrpc: String,
    pub method: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub params: Option<serde_json::Value>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub id: Option<serde_json::Value>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct JsonRpcResponse {
    pub jsonrpc: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub id: Option<serde_json::Value>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub result: Option<serde_json::Value>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub error: Option<JsonRpcError>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct JsonRpcError {
    pub code: i32,
    pub message: String,
}

pub trait ToolHandler: Send + Sync {
    fn call(&self, args: serde_json::Value) -> Box<dyn Future<Output = String> + Send + '_>;
}

pub struct MCPServer {
    config: MCPServerConfig,
    platform_url: String,
    token: String,
    server_id: Option<String>,
    tools: HashMap<String, Box<dyn ToolHandler>>,
    metrics: Metrics,
}

impl MCPServer {
    pub fn new(config: MCPServerConfig) -> Self {
        Self {
            config,
            platform_url: env::var("MANAGING_UP_PLATFORM_URL").unwrap_or_else(|_| "http://localhost:8080".to_string()),
            token: env::var("MANAGING_UP_TOKEN").unwrap_or_default(),
            server_id: None,
            tools: HashMap::new(),
            metrics: Metrics::default(),
        }
    }

    pub fn from_env() -> Self {
        let name = env::var("MANAGING_UP_NAME").unwrap_or_else(|_| "rust-mcp-server".to_string());
        let version = env::var("MANAGING_UP_VERSION").unwrap_or_else(|_| "1.0.0".to_string());

        let transport_str = env::var("MANAGING_UP_TRANSPORT_TYPE").unwrap_or_else(|_| "http".to_string());
        let transport_type = match transport_str.as_str() {
            "stdio" => TransportType::Stdio,
            "http" => TransportType::Http,
            "sse" => TransportType::Sse,
            _ => TransportType::Http,
        };

        Self::new(MCPServerConfig {
            name,
            version,
            description: env::var("MANAGING_UP_DESCRIPTION").ok(),
            transport_type,
            url: env::var("MANAGING_UP_URL").ok(),
            command: None,
            args: None,
        })
    }

    pub fn add_tool<F, Fut>(&mut self, name: &str, description: &str, _input_schema: serde_json::Value, handler: F)
    where
        F: Fn(serde_json::Value) -> Fut + Send + Sync + 'static,
        Fut: Future<Output = String> + Send + 'static,
    {
        self.tools.insert(name.to_string(), Box::new(handler));
    }

    pub async fn register(&mut self) -> Result<(), Box<dyn std::error::Error>> {
        let url = format!("{}/api/v1/mcp-servers", self.platform_url);

        let payload = serde_json::json!({
            "name": self.config.name,
            "version": self.config.version,
            "description": self.config.description,
            "transport_type": self.config.transport_type,
            "url": self.config.url,
            "status": "pending",
        });

        let client = reqwest::Client::new();
        let mut request = client.post(&url)
            .header("Content-Type", "application/json")
            .json(&payload);

        if !self.token.is_empty() {
            request = request.header("Authorization", format!("Bearer {}", self.token));
        }

        match request.send().await {
            Ok(response) => {
                if response.status() == 201 {
                    let data: serde_json::Value = response.json().await?;
                    self.server_id = data["data"]["id"].as_str().map(String::from);
                    println!("Registered with Managing-Up platform: {:?}", self.server_id);
                } else {
                    println!("Registration failed with status: {}", response.status());
                }
            }
            Err(e) => {
                println!("Registration failed: {}", e);
            }
        }

        Ok(())
    }

    pub fn record_request(&mut self, success: bool, latency_secs: f64) {
        self.metrics.requests_total += 1;
        if success {
            self.metrics.requests_success += 1;
        } else {
            self.metrics.requests_error += 1;
        }
        self.metrics.latency_seconds += latency_secs;
    }

    pub fn record_tool_call(&mut self, tool_name: &str) {
        *self.metrics.tool_calls.entry(tool_name.to_string()).or_insert(0) += 1;
    }

    pub fn get_metrics(&self) -> Metrics {
        self.metrics.clone()
    }

    pub async fn serve_stdio(&mut self) -> Result<(), Box<dyn std::error::Error>> {
        let stdin = io::stdin();
        let mut handle = stdin.lock().lines();

        while let Some(line) = handle.next() {
            if let Ok(line) = line {
                let request: JsonRpcRequest = match serde_json::from_str(&line) {
                    Ok(r) => r,
                    Err(_) => continue,
                };

                let response = self.handle_request(request).await;

                if let Some(response) = response {
                    let output = serde_json::to_string(&response).unwrap_or_default();
                    println!("{}", output);
                    io::stdout().flush().ok();
                }
            }
        }

        Ok(())
    }

    async fn handle_request(&mut self, request: JsonRpcRequest) -> Option<JsonRpcResponse> {
        let start = std::time::Instant::now();
        let method = request.method.as_str();
        let id = request.id;

        let result = match method {
            "tools/list" => {
                let tools: Vec<serde_json::Value> = self.tools.keys()
                    .map(|name| {
                        serde_json::json!({
                            "name": name,
                        })
                    })
                    .collect();
                Some(serde_json::json!({ "tools": tools }))
            }
            "tools/call" => {
                if let Some(params) = request.params {
                    let name = params["name"].as_str().unwrap_or("");
                    let args = params["arguments"].clone();

                    if let Some(tool) = self.tools.get(name) {
                        let result = tool.call(args).await;
                        self.record_tool_call(name);
                        self.record_request(true, start.elapsed().as_secs_f64());
                        Some(serde_json::json!({
                            "content": [{ "type": "text", "text": result }]
                        }))
                    } else {
                        self.record_request(false, start.elapsed().as_secs_f64());
                        return Some(JsonRpcResponse {
                            jsonrpc: "2.0".to_string(),
                            id,
                            result: None,
                            error: Some(JsonRpcError {
                                code: -32601,
                                message: format!("Tool not found: {}", name),
                            }),
                        });
                    }
                } else {
                    None
                }
            }
            _ => {
                return Some(JsonRpcResponse {
                    jsonrpc: "2.0".to_string(),
                    id,
                    result: None,
                    error: Some(JsonRpcError {
                        code: -32601,
                        message: "Method not found".to_string(),
                    }),
                });
            }
        };

        Some(JsonRpcResponse {
            jsonrpc: "2.0".to_string(),
            id,
            result,
            error: None,
        })
    }
}

pub trait Future: Send {
    type Output;
}

pub struct AsyncFn<F, Fut> {
    _f: F,
    _phantom: std::marker::PhantomData<Fut>,
}

impl<F, Fut> ToolHandler for F
where
    F: Fn(serde_json::Value) -> Fut + Send + Sync + 'static,
    Fut: Future<Output = String> + Send + 'static,
{
    fn call(&self, args: serde_json::Value) -> Box<dyn Future<Output = String> + Send + '_> {
        Box::new(self(args))
    }
}
