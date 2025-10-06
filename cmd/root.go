package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "voiceterm",
	Short: "A voice-controlled AI terminal assistant",
	Long: `VoiceTerm AI listens to your voice commands, converts them to text,
generates the appropriate shell command using OpenAI, and executes it for you.`,
	Run: func(cmd *cobra.Command, args []string) {
		// --- Color and Style setup ---
		infoColor := color.New(color.FgCyan).SprintFunc()
		promptColor := color.New(color.FgYellow).SprintFunc()
		errorColor := color.New(color.FgRed).SprintFunc()
		commandColor := color.New(color.FgGreen).Add(color.Bold).SprintFunc()
		recordingColor := color.New(color.FgRed).SprintFunc()
		executingColor := color.New(color.FgMagenta).SprintFunc()
		// --- End of setup ---

		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			fmt.Println(errorColor("Error: OPENAI_API_KEY environment variable not set."))
			os.Exit(1)
		}

		tempFile := "voice_command.wav"
		defer os.Remove(tempFile) // Clean up the audio file

		// --- Recording with Spinner and Colors ---
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = recordingColor(" 🎤 Recording audio... (Press Ctrl+C to stop)")
		s.Start()

		err := recordAudio(tempFile)
		s.Stop()
		fmt.Println("\r" + infoColor("✅ Audio recording stopped.")) // Use \r to overwrite spinner line

		if err != nil {
			fmt.Println(errorColor("⚠️ Error recording audio: "), err)
			return
		}

		// --- Transcribing with Spinner ---
		s.Suffix = promptColor(" 🧠 Transcribing audio with Whisper...")
		s.Start()
		transcribedText, err := transcribeAudio(apiKey, tempFile)
		s.Stop()
		if err != nil {
			fmt.Println(errorColor("\n⚠️ Error transcribing audio: "), err)
			return
		}
		fmt.Println("\r"+infoColor("🗣️ You said:"), transcribedText)

		// --- Generating Command with Spinner ---
		s.Suffix = promptColor(" 🤖 Generating command with GPT-4o...")
		s.Start()
		shellCommand, err := generateCommand(apiKey, transcribedText)
		s.Stop()
		if err != nil {
			fmt.Println(errorColor("\n⚠️ Error generating command: "), err)
			return
		}

		// --- Executing Logic with Interactivity and Colors ---
		for {
			fmt.Println("\n" + infoColor("✨ Generated Command:"))
			fmt.Println(commandColor(shellCommand))

			// Check for placeholders like [BRANCH_NAME]
			re := regexp.MustCompile(`\[(.*?)\]`)
			placeholders := re.FindAllStringSubmatch(shellCommand, -1)

			if len(placeholders) > 0 {
				fmt.Println(promptColor("\nThis command requires more information:"))
				for _, p := range placeholders {
					placeholder := p[0] // e.g., [BRANCH_NAME]
					promptText := p[1]  // e.g., BRANCH_NAME
					fmt.Printf("%s: ", strings.ReplaceAll(promptText, "_", " "))
					reader := bufio.NewReader(os.Stdin)
					input, _ := reader.ReadString('\n')
					shellCommand = strings.Replace(shellCommand, placeholder, strings.TrimSpace(input), 1)
				}
				// Loop back to show the filled-in command
				continue
			}

			// Confirmation
			fmt.Print(promptColor("\nExecute this command? (y/n): "))
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.ToLower(strings.TrimSpace(response))

			if response == "y" || response == "yes" {
				fmt.Println(executingColor("\n🚀 Executing command..."))
				executeCommand(shellCommand)
			} else {
				fmt.Println(infoColor("Execution cancelled."))
			}
			break // Exit loop after execution or cancellation
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is the main entry point for the command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// recordAudio uses SoX to record audio from the microphone and stops on Ctrl+C.
func recordAudio(filePath string) error {
	if _, err := exec.LookPath("rec"); err != nil {
		return fmt.Errorf("SoX is not installed. Please install it to use this tool (e.g., on macOS: 'brew install sox')")
	}

	// Command to record a WAV file: 'rec -c 1 -r 16000 -V1 voice_command.wav silence 1 0.1 3% 1 3.0 3%'
	// This will start recording and stop after 3 seconds of silence.
	// For manual stop, we handle signals.
	cmd := exec.Command("rec", "-c", "1", "-r", "16000", "-V1", filePath)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("could not start recording: %w", err)
	}

	// This goroutine will wait for the command to finish, which it will if SoX detects silence.
	// Or it will wait for the interrupt signal.
	go func() {
		<-sigs
		cmd.Process.Signal(os.Interrupt)
	}()

	return cmd.Wait()
}

// transcribeAudio sends the recorded audio file to OpenAI's Whisper API.
func transcribeAudio(apiKey, filePath string) (string, error) {
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile("file", filePath)
	if err != nil {
		return "", fmt.Errorf("creating form file: %w", err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return "", fmt.Errorf("copying file to form: %w", err)
	}

	writer.WriteField("model", "whisper-1")
	writer.WriteField("language", "en")
	writer.Close()

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/audio/transcriptions", &requestBody)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding JSON response: %w", err)
	}

	return result.Text, nil
}

// generateCommand sends the transcribed text to OpenAI's Chat API to get a shell command.
func generateCommand(apiKey, prompt string) (string, error) {
	systemPrompt := `You are an expert shell command assistant. Convert the user's natural language request into a single, executable shell command line.
- If multiple steps are required, chain them together with '&&' or pipes '|'.
- If a piece of information is missing (like a branch name, file name, or commit message), use a placeholder in the format [DESCRIPTION_OF_MISSING_INFO].
- Do not add any explanation, conversational text, or markdown formatting. Only output the raw command.`

	payload := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.0,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshalling payload: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding JSON response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	commandString := result.Choices[0].Message.Content
	commandString = strings.TrimSpace(commandString)
	commandString = strings.TrimPrefix(commandString, "```bash")
	commandString = strings.TrimPrefix(commandString, "```sh")
	commandString = strings.TrimPrefix(commandString, "```")
	commandString = strings.TrimSuffix(commandString, "```")
	commandString = strings.TrimSpace(commandString)

	return commandString, nil
}

// executeCommand runs the final command string.
func executeCommand(command string) {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err != nil {
		errorColor := color.New(color.FgRed).SprintFunc()
		fmt.Println(errorColor("\n⚠️ Error executing command: "), err)
	}
}
