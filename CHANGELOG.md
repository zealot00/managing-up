# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Added
- Redis configuration support (`REDIS_ADDR`, `REDIS_PASSWORD`, `REDIS_DB`)
- Embedding service configuration (`EMBEDDING_PROVIDER`, `EMBEDDING_MODEL`, `EMBEDDING_API_KEY`, `EMBEDDING_BASE_URL`)
- `EMBEDDING_BASE_URL` environment variable for configurable embedding API endpoint

### Changed
- **Gateway**: Improved multi-provider routing with fallback failure switching
- **Gateway**: API Key authentication now uses AES-GCM encryption for secure storage
- **Gateway**: Pricing lookup optimized to O(1) using lowercase map
- **Gateway**: Redis-based distributed rate limiting with circuit breaker
- **MCP**: Two-phase registration (validation separated from network I/O) to avoid blocking other MCP operations
- **MCP**: Proper context lifecycle management to prevent goroutine leaks

### Fixed
- **Worker**: Fixed duplicate execution storm - now uses `sync.Map` for deduplication
- **Worker**: Added bounded semaphore (max 50 concurrent) to prevent overload
- **Router**: Removed unnecessary `currentIndex` state from FallbackRouter
- **Embedding**: Return error on non-200 responses instead of silent nil
- **Stdio validation**: Properly defer cancel() to release timer resources
- **Redis rate limiter**: Added expiration (PX) to `:reset` key to prevent memory leak

## [2026-04-07]

### Added
- Redis-based circuit breaker with exponential backoff
- Redis-based distributed rate limiter
- Redis-based budget checker with atomic check and decrement
- Token budget middleware for gateway endpoints
- Gateway configuration for scanner buffer sizes
- `GATEWAY_MAX_TOKEN_ESTIMATE` for configurable token limit

### Fixed
- Streaming truncation issues
- Usage tracking improvements

## [2026-04-03]

### Added
- PostgreSQL CRUD for provider keys and user budgets
- Gateway provider key management API

### Changed
- Scanner buffer size now configurable via environment variables

## [2026-04-01]

### Added
- LLM Gateway with OpenAI/Anthropic compatible endpoints
- Support for 10 LLM providers (OpenAI, Anthropic, Google, Azure, Ollama, Minimax, Zhipu AI, DeepSeek, Baidu, Alibaba)
- API Key authentication and usage tracking
- Cost tracking per provider/model

## [2026-03-31]

### Added
- MCP Server management API (CRUD + approve workflow)
- MCP Server validation (command whitelist, shell metacharacter validation, CRLF header injection prevention)
- Stdio and HTTP/SSE transport support

## [2026-03-25]

### Added
- Experiment tracking with A/B comparison
- Regression detection
- Trace replay functionality

## [2026-03-20]

### Added
- Skill registry with version control
- SOP to Skill generator
- Execution engine with state machine and checkpoints
- Approval gate for human-in-the-loop

### Changed
- PostgreSQL persistence with migrations
