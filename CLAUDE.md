## Relevant Files and Folders
- CLAUDE.md
- documentation/
- docs/ai-agent-rulesets
- docs/ai-agent-rulesets/specific-requests
- docs/ai-agent-rulesets/thinking
- docs/ai-agent-rulesets/general-principles.md
- docs/ai-agent-rulesets/guidelines-devops-infrastructure.md
- docs/skills-main

## General
- Always follow the project's coding standards and style guide
- Always write clear and concise documentation for your code
- Always write unit tests for new features and bug fixes
- Always use meaningful variable and function names
- Always handle errors and exceptions gracefully

## Git Workflow
Git operations (branching, commits, PRs, releases) are managed by the `claude-git-flow-manager` agent.
See `.claude/agents/claude-git-flow-manager.md` for Git Flow conventions and commit message formats.

## Rulesets (AI Agent Guidelines)

  | File                                           | Purpose                                              |
  | ---------------------------------------------- | ---------------------------------------------------- |
  | general-principles.md                          | Core development principles                          |
  | guidelines-devops-infrastructure.md            | DevOps/Infrastructure (K8s, Docker, Terraform, etc.) |
  | guidelines-golang-backend-api.md               | Go backend API development                           |
  | guidelines-java-kotlin-maven-gradle-backend.md | Java/Kotlin Spring backend                           |
  | guidelines-nixos.md                            | NixOS/Nix Flakes                                     |
  | guidelines-react-typescript.md                 | React + TypeScript frontend                          |
  | specific-requests/systematic-diagnosis.md      | Issue diagnosis workflow                             |
  | specific-requests/systematic-implementation.md | Feature implementation workflow                      |
  | thinking/structured-contemplation.md           | Problem-solving framework                            |
  | thinking/structured-reflection.md              | Learning/self-awareness framework                    |

  Available Slash Commands

  | Command                          | Description                                    |
  | -------------------------------- | ---------------------------------------------- |
  | /structured-reflection           | Expert guide for reflection techniques         |
  | /structured-contemplation        | Expert guide for contemplation/problem-solving |
  | /claude-create-architecture-docs | Generate architecture docs with diagrams       |
  | /systematic-diagnosis            | Issue diagnosis and root cause analysis        |
  | /systematic-implementation       | Validation-driven feature implementation       |

  Available Specialized Agents

  | Agent                                       | Use Case                          |
  | ------------------------------------------- | --------------------------------- |
  | mpc-nextjs-fullstack                        | Next.js 14+ App Router full-stack |
  | mpc-java-kotlin-maven-gradle-backend        | Java/Kotlin Spring backend        |
  | mpc-golang-backend-api                      | Go backend APIs                   |
  | mpc-devops-infrastructure                   | DevOps, K8s, Terraform, CI/CD     |
  | mpc-react-typescript                        | React + TypeScript frontend       |
  | mpc-nixos                                   | NixOS, Nix Flakes, Home Manager   |
  | mpc-sveltekit-frontend                      | Svelte 5 + SvelteKit frontend     |
  | mpc-sveltekit-fullstack                     | SvelteKit full-stack apps         |
  | claude-code-architecture-review             | Review code for best practices    |
  | claude-security-engineer                    | Security/compliance specialist    |
  | claude-mlops-engineer                       | ML pipelines and MLOps            |
  | claude-expert-prompt-engineering            | LLM prompt optimization           |
  | claude-expert-mcp-server                    | MCP server integration            |
  | claude-expert-error-detective               | Log analysis and debugging        |
  | claude-expert-code-review                   | Code review for quality           |
  | claude-expert-agent-creation                | Creating specialized agents       |
  | claude-git-flow-manager                     | Git Flow workflows                |
  | claude-markdown-formatter                   | Markdown formatting specialist    |

  Available Skills

  | Skill             | Purpose                                     |
  | ----------------- | ------------------------------------------- |
  | artifacts-builder | Build complex React/Tailwind HTML artifacts |
  | mcp-builder       | Create MCP servers                          |
  | skill-creator     | Create new skills                           |
  | theme-factory     | Style artifacts with themes (10 presets)    |
