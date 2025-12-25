# aiguide 📖

`aiguide` is a Go-based command-line tool designed to help you master any subject. By leveraging LLMs (OpenAI, Anthropic, or local via Ollama), it automates the process of creating a structured study manual. 

It works in two stages:
1. **Concept Mapping**: It generates a numbered list of core concepts or questions for your chosen subject.
2. **Knowledge Deep-Dive**: It processes those concepts in chunks to generate detailed explanations, outputting a fully formatted Markdown file with a Table of Contents.

## Features

- ⚡ **Concurrent Processing**: Uses Go routines to generate answers in parallel.
- 📂 **Markdown Output**: Generates clean, readable `.md` files with a clickable Table of Contents.
- 🔧 **Highly Configurable**: Control the number of questions, chunk sizes, and thread counts.
- 🧠 **Custom Instructions**: Inject additional context or specific learning styles via the `--info` flag.
- 🔗 **API Compatible**: Works with OpenAI or any OpenAI-compatible API (like DeepSeek, Groq, or Ollama).

## Installation

```bash
go install github.com/yuriiter/aiguide@latest
```

## Setup

Set your environment variables:

```bash
export OPENAI_API_KEY="your-api-key"
# Optional overrides:
export OPENAI_BASE_URL="https://api.openai.com/v1"
export OPENAI_MODEL="gpt-4o"
```

## Usage

Generate a guide by simply providing a subject:

```bash
aiguide "Quantum Computing"
```

### Advanced Examples

**Generate a deep-dive guide with 50 concepts using 5 parallel threads:**
```bash
aiguide "Rust Programming" -n 50 -t 5
```

**Tailor the guide for a specific audience (e.g., a beginner):**
```bash
aiguide "Organic Chemistry" --info "Explain things as if I am a high school student with no prior knowledge."
```

**Output directly to the terminal:**
```bash
aiguide "French Revolution" --stdout
```

## Options

| Flag | Shorthand | Description | Default |
|------|-----------|-------------|---------|
| `--number` | `-n` | Total number of concepts to generate | 100 |
| `--chunk` | `-c` | Concepts per AI request | 10 |
| `--threads` | `-t` | Number of concurrent workers | 1 |
| `--info` | `-i` | Additional context for the AI | "" |
| `--stdout` | `-o` | Output to console instead of file | false |

## How it works

1. **Phase 1**: The tool asks the AI to create a curriculum/list based on your subject.
2. **Phase 2**: It generates a Table of Contents with internal anchors.
3. **Phase 3**: It splits the list into "chunks" and sends them to the AI to write detailed explanations.
4. **Phase 4**: Everything is compiled into a single timestamped Markdown file.
