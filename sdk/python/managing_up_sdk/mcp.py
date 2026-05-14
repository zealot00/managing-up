"""
Managing-Up MCP SDK for Python
"""

import os
import json
import logging
from typing import Any, Callable, Optional
from dataclasses import dataclass, field
from enum import Enum

logger = logging.getLogger(__name__)


class TransportType(str, Enum):
    STDIO = "stdio"
    HTTP = "http"
    SSE = "sse"


@dataclass
class MCPServerConfig:
    name: str
    version: str
    description: Optional[str] = None
    transport_type: TransportType = TransportType.HTTP
    url: Optional[str] = None
    command: Optional[str] = None
    args: list = field(default_factory=list)
    env: dict = field(default_factory=dict)


@dataclass
class Metrics:
    requests_total: int = 0
    requests_success: int = 0
    requests_error: int = 0
    latency_seconds: float = 0.0
    tool_calls: dict = field(default_factory=dict)

    def to_dict(self) -> dict:
        return {
            "requests_total": self.requests_total,
            "requests_success": self.requests_success,
            "requests_error": self.requests_error,
            "latency_seconds": self.latency_seconds,
            "tool_calls": self.tool_calls,
        }


class MCPServer:
    def __init__(self, config: MCPServerConfig):
        self.config = config
        self.platform_url = os.getenv(
            "MANAGING_UP_PLATFORM_URL", "http://localhost:8080"
        )
        self.token = os.getenv("MANAGING_UP_TOKEN", "")
        self.server_id: Optional[str] = None

        self._tools: dict[str, tuple[Callable, str, dict]] = {}
        self._resources: dict[str, tuple[Callable, str, str]] = {}  # uri -> (handler, name, mime_type)
        self._prompts: dict[str, tuple[Callable, str, list]] = {}  # name -> (handler, description, arguments)
        self._metrics = Metrics()
        self._server_impl = None

    @classmethod
    def from_env(cls) -> "MCPServer":
        config = MCPServerConfig(
            name=os.getenv("MANAGING_UP_NAME", "python-mcp-server"),
            version=os.getenv("MANAGING_UP_VERSION", "1.0.0"),
            description=os.getenv("MANAGING_UP_DESCRIPTION", ""),
            transport_type=TransportType(
                os.getenv("MANAGING_UP_TRANSPORT_TYPE", "http")
            ),
            url=os.getenv("MANAGING_UP_URL"),
        )
        return cls(config)

    def add_tool(
        self, name: str, description: str, input_schema: dict, handler: Callable
    ):
        self._tools[name] = (handler, description, input_schema)

    def add_resource(self, uri: str, name: str, mime_type: str, handler: Callable):
        """Add a resource to the MCP server.

        Args:
            uri: The URI of the resource.
            name: Human-readable name for the resource.
            mime_type: MIME type of the resource content.
            handler: Async callable that takes no args and returns the resource content.
        """
        self._resources[uri] = (handler, name, mime_type)

    def add_prompt(self, name: str, description: str, arguments: list, handler: Callable):
        """Add a prompt template to the MCP server.

        Args:
            name: Name of the prompt.
            description: Human-readable description.
            arguments: List of argument dicts with 'name', 'description', 'required' keys.
            handler: Async callable that takes arguments dict and returns prompt messages.
        """
        self._prompts[name] = (handler, description, arguments)

    async def register(self) -> dict:
        import httpx

        url = f"{self.platform_url}/api/v1/mcp-servers"
        headers = {"Content-Type": "application/json"}
        if self.token:
            headers["Authorization"] = f"Bearer {self.token}"

        payload = {
            "name": self.config.name,
            "version": self.config.version,
            "description": self.config.description or "",
            "transport_type": self.config.transport_type.value,
            "url": self.config.url or "",
            "status": "pending",
        }

        async with httpx.AsyncClient() as client:
            try:
                response = await client.post(url, json=payload, headers=headers)
                if response.status_code == 201:
                    data = response.json()
                    self.server_id = data.get("data", {}).get("id")
                    logger.info(
                        f"Registered with Managing-Up platform: {self.server_id}"
                    )
                    return data
                else:
                    logger.warning(
                        f"Registration failed with status {response.status_code}"
                    )
                    return {}
            except Exception as e:
                logger.warning(f"Registration failed: {e}")
                return {}

    def record_request(self, success: bool, latency: float):
        self._metrics.requests_total += 1
        if success:
            self._metrics.requests_success += 1
        else:
            self._metrics.requests_error += 1
        self._metrics.latency_seconds += latency

    def record_tool_call(self, tool_name: str):
        if tool_name not in self._metrics.tool_calls:
            self._metrics.tool_calls[tool_name] = 0
        self._metrics.tool_calls[tool_name] += 1

    def get_metrics(self) -> Metrics:
        return self._metrics

    async def serve_stdio(self):
        import sys
        import asyncio

        loop = asyncio.get_event_loop()

        async def handle_request(request: dict) -> dict:
            method = request.get("method", "")
            params = request.get("params", {})
            id = request.get("id")

            if method == "tools/list":
                tools = [
                    {"name": name, "description": desc, "inputSchema": schema}
                    for name, (_, desc, schema) in self._tools.items()
                ]
                return {"jsonrpc": "2.0", "id": id, "result": {"tools": tools}}

            elif method == "resources/list":
                resources = [
                    {"uri": uri, "name": name, "mimeType": mime_type}
                    for uri, (_, name, mime_type) in self._resources.items()
                ]
                return {"jsonrpc": "2.0", "id": id, "result": {"resources": resources}}

            elif method == "resources/read":
                uri = params.get("uri", "")
                if uri in self._resources:
                    handler = self._resources[uri][0]
                    content = await handler()
                    return {
                        "jsonrpc": "2.0",
                        "id": id,
                        "result": {
                            "contents": [
                                {"uri": uri, "mimeType": self._resources[uri][2], "text": str(content)}
                            ]
                        },
                    }
                return {"jsonrpc": "2.0", "id": id, "error": {"code": -32601, "message": f"Resource not found: {uri}"}}

            elif method == "prompts/list":
                prompts = [
                    {"name": name, "description": desc, "arguments": args}
                    for name, (_, desc, args) in self._prompts.items()
                ]
                return {"jsonrpc": "2.0", "id": id, "result": {"prompts": prompts}}

            elif method == "prompts/get":
                prompt_name = params.get("name", "")
                if prompt_name in self._prompts:
                    handler = self._prompts[prompt_name][0]
                    arguments = params.get("arguments", {})
                    result = await handler(arguments)
                    return {"jsonrpc": "2.0", "id": id, "result": result}
                return {"jsonrpc": "2.0", "id": id, "error": {"code": -32601, "message": f"Prompt not found: {prompt_name}"}}

            elif method == "tools/call":
                tool_name = params.get("name")
                arguments = params.get("arguments", {})

                import time

                start = time.time()

                try:
                    if tool_name in self._tools:
                        handler = self._tools[tool_name][0]
                        result = await handler(arguments)
                        self.record_tool_call(tool_name)
                        self.record_request(True, time.time() - start)
                        return {
                            "jsonrpc": "2.0",
                            "id": id,
                            "result": {
                                "content": [{"type": "text", "text": str(result)}]
                            },
                        }
                    else:
                        self.record_request(False, time.time() - start)
                        return {
                            "jsonrpc": "2.0",
                            "id": id,
                            "error": {
                                "code": -32601,
                                "message": f"Tool not found: {tool_name}",
                            },
                        }
                except Exception as e:
                    self.record_request(False, time.time() - start)
                    return {
                        "jsonrpc": "2.0",
                        "id": id,
                        "error": {"code": -32603, "message": str(e)},
                    }

            return {
                "jsonrpc": "2.0",
                "id": id,
                "error": {"code": -32601, "message": "Method not found"},
            }

        while True:
            line = await loop.run_in_executor(None, sys.stdin.readline)
            if not line:
                break

            try:
                request = json.loads(line.strip())
                response = await handle_request(request)
                if response:
                    print(json.dumps(response), flush=True)
            except json.JSONDecodeError:
                continue


async def main():
    import argparse

    parser = argparse.ArgumentParser(description="Managing-Up MCP Server")
    parser.add_argument("--name", default=os.getenv("MANAGING_UP_NAME"))
    parser.add_argument("--version", default=os.getenv("MANAGING_UP_VERSION", "1.0.0"))
    parser.add_argument("--transport", default="http")
    args = parser.parse_args()

    server = MCPServer.from_env()

    await server.register()
    await server.serve_stdio()


if __name__ == "__main__":
    asyncio.run(main())
