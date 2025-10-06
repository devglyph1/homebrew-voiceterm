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
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// Global variable to hold the OpenAI API Key.
var apiKey string

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "voiceterm",
	Short: "A voice-powered AI terminal assistant",
	Long: `VoiceTerm listens to your voice commands, converts them to text,
and uses OpenAI to generate and execute the corresponding shell commands.

Example:
Run 'voiceterm', wait for the prompt, and say:
"Save all my changes, commit them with the message initial commit, and push to the main branch."`,
	Run: func(cmd *cobra.Command, args []string) {
		runVoiceTerm()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Add a persistent flag for the API key. It can be set via command line or environment variable.
	rootCmd.PersistentFlags().StringVarP(&apiKey, "api-key", "k", "", "OpenAI API key (or use OPENAI_API_KEY env var)")
}

// Main application logic.
func runVoiceTerm() {
	// Check for API key.
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if apiKey == "" {
		fmt.Println("Error: OpenAI API key not found.")
		fmt.Println("Please set it using the --api-key flag or the OPENAI_API_KEY environment variable.")
		return
	}

	fmt.Println("🎤 Press Enter to start recording, press Enter again to stop...")
	bufio.NewReader(os.Stdin).ReadBytes('\n') // Wait for user to press Enter to start

	// --- 1. Record Audio ---
	// We use the 'rec' command from SoX. It's a reliable cross-platform audio recorder.
	// The command records a WAV file and automatically stops on silence.
	audioFilePath := "command.wav"
	// Ensure SoX is installed.
	if _, err := exec.LookPath("rec"); err != nil {
		fmt.Println("Error: SoX is not installed. Please install it to use this tool.")
		fmt.Println("On macOS: brew install sox")
		fmt.Println("On Debian/Ubuntu: sudo apt-get install sox")
		return
	}

	fmt.Println("🔴 Recording... Press Enter to stop.")

	// Start recording in a goroutine
	recCmd := exec.Command("rec", "-c", "1", "-r", "16000", audioFilePath)
	if err := recCmd.Start(); err != nil {
		fmt.Printf("Error starting recording: %v\n", err)
		return
	}

	// Wait for user to press Enter to stop
	bufio.NewReader(os.Stdin).ReadBytes('\n')
	if err := recCmd.Process.Signal(os.Interrupt); err != nil { // Send interrupt to stop rec gracefully
		fmt.Printf("Error stopping recording: %v\n", err)
		return
	}
	recCmd.Wait() // Wait for the process to exit
	fmt.Println("✅ Recording finished.")
	defer os.Remove(audioFilePath) // Clean up the audio file afterwards.

	// --- 2. Transcribe Audio to Text (using OpenAI Whisper) ---
	fmt.Println("🤖 Transcribing audio to text...")
	transcribedText, err := transcribeAudio(audioFilePath)
	if err != nil {
		fmt.Printf("Error transcribing audio: %v\n", err)
		return
	}
	fmt.Printf("🗣️ You said: \"%s\"\n", transcribedText)

	// --- 3. Convert Text to Shell Commands (using OpenAI Chat API) ---
	fmt.Println("🧠 Generating shell commands...")
	commands, err := generateCommands(transcribedText)
	if err != nil {
		fmt.Printf("Error generating commands: %v\n", err)
		return
	}

	if len(commands) == 0 || (len(commands) == 1 && commands[0] == "") {
		fmt.Println("🤷 No commands were generated. Please try a different phrase.")
		return
	}

	fmt.Println("💻 Generated Commands:")
	for _, c := range commands {
		fmt.Printf("- %s\n", c)
	}
	fmt.Println("---------------------------------")

	// --- 4. Execute Commands (with interactive prompts for placeholders) ---
	executeCommands(commands)
}

// transcribeAudio sends the recorded audio file to OpenAI's Whisper API.
func transcribeAudio(filePath string) (string, error) {
	// Create a buffer to store our request body
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Open the audio file
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	// Create a form file field
	part, err := writer.CreateFormFile("file", filePath)
	if err != nil {
		return "", fmt.Errorf("creating form file: %w", err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return "", fmt.Errorf("copying file to form: %w", err)
	}

	// Add the model field
	writer.WriteField("model", "whisper-1")

	// Close the writer
	writer.Close()

	// Create the HTTP request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/audio/transcriptions", &requestBody)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Decode the JSON response
	var result struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding JSON response: %w", err)
	}

	return result.Text, nil
}

// generateCommands sends the transcribed text to OpenAI's Chat API to get shell commands.
func generateCommands(prompt string) ([]string, error) {
	systemPrompt := `You are an expert shell command assistant. Convert the user's natural language request into a sequence of shell commands.
- Each command must be on a new line.
- If a piece of information is missing (like a branch name, file name, or commit message), use a placeholder in the format [DESCRIPTION_OF_MISSING_INFO].
- Do not add any explanation, conversational text, or markdown formatting. Only output the raw commands.`

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
		return nil, fmt.Errorf("marshalling payload: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding JSON response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	commandString := result.Choices[0].Message.Content
	// Clean up potential markdown code blocks
	commandString = strings.TrimPrefix(commandString, "```bash\n")
	commandString = strings.TrimPrefix(commandString, "```sh\n")
	commandString = strings.TrimPrefix(commandString, "```\n")
	commandString = strings.TrimSuffix(commandString, "\n```")
	commandString = strings.TrimSpace(commandString)

	return strings.Split(commandString, "\n"), nil
}

// executeCommands runs the generated commands, prompting for placeholders if necessary.
func executeCommands(commands []string) {
	// Regex to find placeholders like [some text]
	re := regexp.MustCompile(`\[(.*?)\]`)
	reader := bufio.NewReader(os.Stdin)

	for _, command := range commands {
		finalCommand := command
		// Find all placeholders in the current command
		placeholders := re.FindAllStringSubmatch(finalCommand, -1)

		for _, placeholder := range placeholders {
			// placeholder[0] is the full match, e.g., "[BRANCH_NAME]"
			// placeholder[1] is the group, e.g., "BRANCH_NAME"
			promptText := strings.ReplaceAll(placeholder[1], "_", " ")
			fmt.Printf("❓ Please provide %s: ", promptText)

			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)

			// Replace the first occurrence of the placeholder with the user's input
			finalCommand = strings.Replace(finalCommand, placeholder[0], input, 1)
		}

		fmt.Printf("🚀 Executing: %s\n", finalCommand)

		// Execute the command using sh -c to handle pipes, etc.
		cmd := exec.Command("sh", "-c", finalCommand)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin // Pass through stdin for commands that might need it

		err := cmd.Run()
		if err != nil {
			fmt.Printf("❌ Error executing command: %v\n", err)
			// Ask the user if they want to continue
			fmt.Print("Do you want to continue with the next command? (y/N): ")
			choice, _ := reader.ReadString('\n')
			if strings.TrimSpace(strings.ToLower(choice)) != "y" {
				fmt.Println("Aborting.")
				return
			}
		}
	}
	fmt.Println("🎉 All commands executed successfully!")
}
