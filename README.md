# aiguide 📖

> **Master any subject with AI-generated, structured study guides.**

`aiguide` is a CLI tool written in Go that acts as your personal curriculum developer. It leverages LLMs (OpenAI, Anthropic, Ollama, etc.) to generate comprehensive, academic-grade study manuals. It breaks complex subjects into core concepts, creates a roadmap, and generates deep-dive explanations with diagrams and code examples.

## ✨ Features

- **🚀 Parallel Generation**: Utilizes concurrent workers to generate content fast without losing order.
- **📑 Structured Output**: Produces a single, cohesive Markdown file with a working Table of Contents.
- **🧠 Deep Dives**: optimized for depth—defaults to small chunks to allow the AI to generate long, detailed explanations.
- **🎨 Visuals & Code**: Automatically requests **Mermaid.js** diagrams for workflows and code snippets for technical concepts.
- **🔧 Custom Prompts**: Bring your own system prompt or inject specific context via CLI flags.
- **🔌 Universal API Support**: Works with OpenAI or any compatible endpoint (LocalAI, Ollama, DeepSeek, etc.).

## 📦 Installation

```bash
go install github.com/yuriiter/aiguide@latest
```

## ⚙️ Configuration

Set your environment variables. Only the API Key is strictly required; others have sensible defaults.

```bash
# Required
export OPENAI_API_KEY="sk-..."

# Optional (Defaults shown)
export OPENAI_BASE_URL="https://api.openai.com/v1"
export OPENAI_MODEL="gpt-4o"
```

## 🚀 Usage

### Basic Usage
Generate a standard guide (100 concepts) for a subject:

```bash
aiguide "Quantum Physics"
```

### Advanced Usage

**1. faster generation with threads:**
Use 5 concurrent threads to speed up the process.
```bash
aiguide "System Design" -t 5
```

**2. Customizing the depth:**
Generate only 20 key concepts, but keep the chunk size small (2) for maximum detail per answer.
```bash
aiguide "Kubernetes Networking" -n 20 -c 2
```

**3. Custom Persona/Prompt:**
Point to a custom system prompt file (e.g., to generate content in a specific language or tone).
```bash
aiguide "French History" --system-prompt ./prompts/french_tutor.txt
```

**4. Ad-hoc Instructions:**
Add specific constraints without changing the file.
```bash
aiguide "React Hooks" -i "Focus heavily on performance pitfalls and rendering cycles."
```

## 🚩 Options / Flags

| Flag | Short | Default | Description |
|------|-------|:-------:|-------------|
| `--number` | `-n` | `100` | Total number of concepts/questions to generate. |
| `--chunk` | `-c` | `2` | Number of items to process per API call. Lower = more detail. |
| `--threads` | `-t` | `1` | Number of concurrent API workers. |
| `--stdout` | `-o` | `false` | Print to console instead of writing to a file. |
| `--info` | `-i` | `""` | Append extra instructions to the system prompt. |
| `--system-prompt`| `-s` | `(embedded)`| Path to a custom system prompt text file. |

## 🛠️ How it Works

1. **Curriculum Generation**: The tool asks the AI to list exactly `N` core concepts regarding your subject.
2. **Structure & ToC**: It parses this list and pre-calculates a Table of Contents with valid Markdown anchors.
3. **Parallel Processing**: 
   - The list is split into chunks (default size: 2).
   - Worker threads send these chunks to the API.
   - Results are buffered in memory to ensure the **final output remains strictly ordered**, regardless of which thread finishes first.
4. **Cleanup**: It strips Markdown artifacts (like fencing) and compiles the final `.md` file.

## 🤝 Contributing

Pull requests are welcome! Please ensure you respect the formatting logic for the Table of Contents.

## 📄 License

MIT

