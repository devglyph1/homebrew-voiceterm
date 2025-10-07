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

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var language = "en"

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
		logoColor := color.New(color.FgHiMagenta, color.Bold).SprintFunc()
		// --- End of setup ---

		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			fmt.Println(errorColor("Error: OPENAI_API_KEY environment variable not set."))
			os.Exit(1)
		}

		// --- Welcome Message ---
		fmt.Println(logoColor(`
 _  _   __  __  ___  ____  ____  ____  ____  _  _ 
/ )( \ /  \(  )/ __)(  __)(_  _)(  __)(  _ \( \/ )
\ \/ /(  O ))(( (__  ) _)   )(   ) _)  )   // \/ \
 \__/  \__/(__)\___)(____) (__) (____)(__\_)\_)(_/
                                                                        
    `))
		fmt.Println(infoColor("Welcome to VoiceTerm, your AI smart voice terminal assistant!"))
		fmt.Println("-----------------------------------------------------------------")

		// --- Main Application Loop ---
		for {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print(promptColor("\nPress Enter to start recording (or Ctrl+C to exit)..."))
			_, err := reader.ReadString('\n')
			if err != nil { // Handles Ctrl+C (EOF)
				fmt.Println(infoColor("\nGoodbye!"))
				return
			}

			tempFile := "voice_command.wav"
			defer os.Remove(tempFile) //Ensure cleanup

			// --- Recording with Spinner and Colors ---
			s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
			s.Suffix = recordingColor(" 🎤 Recording... (Press Enter to stop)")
			s.Start()

			// Check if SoX is installed
			if _, err := exec.LookPath("rec"); err != nil {
				s.Stop()
				fmt.Println(errorColor("Error: SoX is not installed. Please install it to use this tool (e.g., on macOS: 'brew install sox')"))
				return
			}

			recCmd := exec.Command("rec", "-c", "1", "-r", "16000", "-V1", tempFile)
			if err := recCmd.Start(); err != nil {
				s.Stop()
				fmt.Println(errorColor("⚠️ Error starting recording: "), err)
				os.Remove(tempFile) // Clean up even on failure
				continue
			}

			// Goroutine to wait for 'Enter' to stop recording
			go func() {
				reader.ReadString('\n')
				if recCmd.Process != nil {
					recCmd.Process.Signal(os.Interrupt)
				}
			}()

			recCmd.Wait() // Wait for the recording process to be interrupted and finish
			s.Stop()
			fmt.Println("\r" + infoColor("✅ Audio recording stopped."))

			// --- Transcribing with Spinner ---
			s.Suffix = promptColor(" 🧠 Transcribing audio with Whisper...")
			s.Start()
			transcribedText, err := transcribeAudio(apiKey, tempFile, language)
			s.Stop()
			if err != nil {
				fmt.Println(errorColor("\n⚠️ Error transcribing audio: "), err)
				os.Remove(tempFile) // Clean up even on failure
				continue
			}
			fmt.Println("\r"+infoColor("🗣️ You said:"), transcribedText)

			// --- Generating Command with Spinner ---
			s.Suffix = promptColor(" 🤖 Generating command with GPT-4o...")
			s.Start()
			shellCommand, err := generateCommand(apiKey, transcribedText)
			s.Stop()
			if err != nil {
				fmt.Println(errorColor("\n⚠️ Error generating command: "), err)
				os.Remove(tempFile) // Clean up even on failure
				continue
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
						input, _ := reader.ReadString('\n')
						shellCommand = strings.Replace(shellCommand, placeholder, strings.TrimSpace(input), 1)
					}
					// Loop back to show the filled-in command
					continue
				}

				// Directly execute the command
				fmt.Println(executingColor("\n🚀 Executing command..."))
				executeCommand(shellCommand)
				break // Exit inner loop and wait for new recording
			}
			// Clean up the audio file at the end of the loop iteration
			os.Remove(tempFile)
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

func init() {
	rootCmd.Flags().StringVarP(&language, "language", "l", "", "Optional: Language for transcription (e.g., 'en', 'hi'). Defaults to auto-detection.")
}

// transcribeAudio sends the recorded audio file to OpenAI's Whisper API.
func transcribeAudio(apiKey, filePath, lang string) (string, error) {
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	file, err := os.Open(filePath)
	if err != nil {
		// If file doesn't exist, it's likely recording was stopped too fast
		if os.IsNotExist(err) {
			return "", fmt.Errorf("audio file not found. Recording may have been too short")
		}
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
	// Only set the language if the user provides the flag.
	// Otherwise, let Whisper auto-detect it.
	writer.WriteField("language", lang)
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
	if strings.TrimSpace(prompt) == "" {
		return "", fmt.Errorf("no text was transcribed. Cannot generate command")
	}

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
