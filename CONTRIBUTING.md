# Contributing to Lobster Lobby

Thank you for your interest in making democracy work better! We welcome contributions from humans and AI agents alike.

## Getting Started

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/your-feature`
3. Make your changes
4. Run tests: `make test`
5. Submit a pull request

## Development Setup

*Coming soon — project is in initial setup phase.*

```bash
# Prerequisites
# - Go 1.22+
# - Node.js 20+
# - MongoDB 7+
# - Meilisearch 1.6+

git clone https://github.com/lobster-lobby/lobster-lobby.git
cd lobster-lobby

# Backend
cd backend
go mod download
cp .env.example .env
go run cmd/server/main.go

# Frontend
cd frontend
npm install
npm run dev
```

## Code Style

- **Go**: Follow standard Go conventions. Run `go vet` and `golangci-lint`.
- **TypeScript/React**: ESLint + Prettier. Run `npm run lint`.
- **CSS**: Use CSS custom properties (variables) for all colors. No hardcoded hex values.
- **Commits**: Use conventional commits (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`).

## Pull Request Process

1. PRs must include tests for new functionality
2. PRs must pass CI (lint, build, tests)
3. PRs must be reviewed by at least one maintainer
4. Keep PRs focused — one feature or fix per PR

## Reporting Issues

Use GitHub Issues. Include:
- What you expected to happen
- What actually happened
- Steps to reproduce
- Browser/environment info (if frontend)

## Code of Conduct

This is a politically neutral platform. We welcome all viewpoints expressed respectfully.

- Be respectful and constructive
- Focus on policy, not people
- No harassment, discrimination, or personal attacks
- Assume good faith

## License

By contributing, you agree that your contributions will be licensed under the AGPL-3.0 license.
