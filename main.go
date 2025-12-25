package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

//go:embed system_prompt.txt
var embedSystemPrompt string

type Config struct {
	BaseURL      string
	Token        string
	Model        string
	Subject      string
	TotalCount   int
	ChunkSize    int
	Stdout       bool
	Threads      int
	Info         string
	SystemPrompt string
}

var cfg Config

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type CompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
}

type CompletionResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "aiguide [subject]",
		Short: "Generate an AI-powered study guide",
		Args:  cobra.ExactArgs(1),
		Run:   run,
	}

	rootCmd.Flags().IntVarP(&cfg.TotalCount, "number", "n", 100, "Total number of questions/concepts to generate")
	rootCmd.Flags().IntVarP(&cfg.ChunkSize, "chunk", "c", 10, "Number of questions to process per API call")
	rootCmd.Flags().BoolVarP(&cfg.Stdout, "stdout", "o", false, "Output to stdout instead of file")
	rootCmd.Flags().IntVarP(&cfg.Threads, "threads", "t", 1, "Number of concurrent threads for generating answers")
	rootCmd.Flags().StringVarP(&cfg.Info, "info", "i", "", "Additional instructions or context to append to system prompt")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	cfg.Subject = args[0]
	loadEnv()

	cfg.SystemPrompt = embedSystemPrompt
	if cfg.Info != "" {
		cfg.SystemPrompt += "\n\nADDITIONAL USER INSTRUCTIONS:\n" + cfg.Info
	}

	fmt.Printf("-> Generating list of %d concepts for subject: %s...\n", cfg.TotalCount, cfg.Subject)
	concepts, err := generateConceptList()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating concepts: %v\n", err)
		os.Exit(1)
	}

	if len(concepts) == 0 {
		fmt.Println("No concepts were generated. Exiting.")
		os.Exit(1)
	}

	var writer io.Writer

	if cfg.Stdout {
		writer = os.Stdout
	} else {
		cleanSubject := regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(cfg.Subject, "_")
		filename := fmt.Sprintf("%s_%s.md", cleanSubject, time.Now().Format("20060102-150405"))
		f, err := os.Create(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		writer = f
		fmt.Printf("-> Outputting to: %s\n", filename)
	}

	writeHeaderAndToC(writer, concepts)

	processChunks(writer, concepts)

	if !cfg.Stdout {
		fmt.Println("\n-> Done! Guide generated successfully.")
	}
}
func loadEnv() {
	cfg.BaseURL = os.Getenv("OPENAI_BASE_URL")
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.openai.com/v1"
	}
	cfg.Token = os.Getenv("OPENAI_API_KEY")
	if cfg.Token == "" {
		fmt.Fprintln(os.Stderr, "Error: OPENAI_API_KEY environment variable is required.")
		os.Exit(1)
	}
	cfg.Model = os.Getenv("OPENAI_MODEL")
	if cfg.Model == "" {
		cfg.Model = "gpt-4o"
	}
}

func generateConceptList() ([]string, error) {
	prompt := fmt.Sprintf(
		"Generate a numbered list of exactly %d core questions or concepts regarding the subject: '%s'. "+
			"Output ONLY the numbered list. Do not add introductions or conclusions.",
		cfg.TotalCount, cfg.Subject,
	)

	resp, err := callAI(prompt, "You are a helpful assistant that lists concepts concisely.")
	if err != nil {
		return nil, err
	}

	var cleanList []string
	scanner := bufio.NewScanner(strings.NewReader(resp))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && (unicodeIsDigit(line[0]) || strings.HasPrefix(line, "-")) {
			cleanList = append(cleanList, line)
		}
	}
	return cleanList, nil
}

func unicodeIsDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func writeHeaderAndToC(w io.Writer, concepts []string) {
	title := fmt.Sprintf("# Comprehensive Guide: %s\n\n", strings.ToUpper(cfg.Subject))
	toc := "## Table of Contents\n\n"

	for _, c := range concepts {
		parts := strings.SplitN(c, " ", 2)
		if len(parts) < 2 {
			continue
		}
		linkRef := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(c, " ", "-"), ".", ""))
		reg := regexp.MustCompile("[^a-z0-9-]+")
		linkRef = reg.ReplaceAllString(linkRef, "")

		toc += fmt.Sprintf("- [%s](#%s)\n", c, linkRef)
	}
	toc += "\n---\n\n"

	fmt.Fprint(w, title)
	fmt.Fprint(w, toc)
}

func processChunks(w io.Writer, concepts []string) {
	total := len(concepts)
	var wg sync.WaitGroup

	type job struct {
		index int
		items []string
	}
	jobs := make(chan job, total/cfg.ChunkSize+1)

	var writeMutex sync.Mutex

	for i := 0; i < cfg.Threads; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := range jobs {
				if !cfg.Stdout {
					fmt.Printf("   [Worker %d] Processing chunk %d-%d...\n", workerID, j.index+1, j.index+len(j.items))
				}

				chunkText := strings.Join(j.items, "\n")
				prompt := fmt.Sprintf(
					"Here is a list of concepts/questions:\n%s\n\n"+
						"Provide a detailed, numbered explanation for EACH one based on the system prompt instructions. "+
						"Maintain the original numbering.",
					chunkText,
				)

				content, err := callAI(prompt, cfg.SystemPrompt)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error processing chunk starting at %d: %v\n", j.index, err)
					continue
				}

				writeMutex.Lock()
				fmt.Fprintln(w, content)
				fmt.Fprintln(w, "\n---")
				writeMutex.Unlock()
			}
		}(i)
	}

	for i := 0; i < total; i += cfg.ChunkSize {
		end := i + cfg.ChunkSize
		if end > total {
			end = total
		}
		jobs <- job{index: i, items: concepts[i:end]}
	}
	close(jobs)

	wg.Wait()
}

func callAI(userPrompt, sysPrompt string) (string, error) {
	reqBody := CompletionRequest{
		Model: cfg.Model,
		Messages: []Message{
			{Role: "system", Content: sysPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.7,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: 120 * time.Second}
	req, err := http.NewRequest("POST", cfg.BaseURL+"chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.Token)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error: %s - %s", resp.Status, string(bodyBytes))
	}

	var completion CompletionResponse
	if err := json.Unmarshal(bodyBytes, &completion); err != nil {
		return "", err
	}

	if completion.Error != nil {
		return "", fmt.Errorf("API returned error: %s", completion.Error.Message)
	}

	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("no choices returned")
	}

	return completion.Choices[0].Message.Content, nil
}
