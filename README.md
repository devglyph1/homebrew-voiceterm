VoiceTerm AI 🎤
VoiceTerm is a smart command-line interface (CLI) tool that transforms your spoken language into executable shell commands. It acts as your personal voice assistant for the terminal, designed to streamline your workflow and reduce typing.

What It Does
VoiceTerm listens for a voice command, intelligently converts it into the correct shell command or sequence of commands, and executes them for you. It's perfect for complex, multi-step operations like Git workflows or file management.

Voice-to-Command: Simply speak what you want to do, like "commit my changes with the message 'bug fix' and push to the develop branch."

AI-Powered: Uses state-of-the-art AI to understand your intent and generate accurate commands.

Interactive & Safe: It shows you the generated commands before running them and interactively asks for any missing information (like a filename or a specific branch name).

How It Works
The tool follows a simple four-step process:

🎙️ Record Audio: It uses the SoX utility to capture audio from your microphone when you press Enter.

✍️ Transcribe to Text: The recorded audio is sent to OpenAI's Whisper API, which transcribes the speech into text with high accuracy, specifically configured for English.

🧠 Generate Commands: The transcribed text is then sent to OpenAI's Chat API (gpt-4o). A specialized prompt instructs the model to convert your request into precise shell commands. If information is missing, it creates a placeholder (e.g., [COMMIT_MESSAGE]).

🚀 Execute Commands: The application parses the generated commands. If it finds any placeholders, it prompts you to provide the missing information. Finally, it executes the completed commands one by one in your shell.

Technology Stack
VoiceTerm is built with the following technologies:

Go: The programming language used for the CLI application.

Cobra: A popular Go library for creating powerful and modern CLI applications.

SoX (Sound eXchange): A cross-platform command-line utility for audio recording.

OpenAI Whisper API: For best-in-class, English-language voice-to-text transcription.

OpenAI Chat API (gpt-4o): For natural language understanding and command generation.

Prerequisites
Before installing, you must have the following set up on your system:

Homebrew (for macOS or Linux)

If you don't have it, install it from brew.sh.

SoX (Sound eXchange)

This is essential for recording audio.

Install via Homebrew:

brew install sox

On Debian/Ubuntu:

sudo apt-get update && sudo apt-get install sox libsox-fmt-all

OpenAI API Key

You need an active API key from the OpenAI Platform.

For the tool to work, you must set this key as an environment variable. Add the following line to your shell's configuration file (~/.zshrc, ~/.bash_profile, etc.):

export OPENAI_API_KEY='your-api-key-here'

Remember to run source ~/.zshrc (or your shell's equivalent) or restart your terminal for the changes to take effect.

Installation
Once you have published your Homebrew tap, voiceterm can be easily installed.

Tap the repository (you only need to do this once):

brew tap devglyph1/voiceterm

Install voiceterm:

brew install voiceterm

Usage Example
Using voiceterm is designed to be intuitive.

Open your terminal.

Run the command:

voiceterm

The tool will prompt you: 🎤 Press Enter to start recording, press Enter again to stop...

Press Enter, and you will see: 🔴 Recording... Press Enter to stop.

Speak your command clearly. For example:

"Stage all my changes, commit with the message 'refactor api module', and then push to the main branch."

Press Enter again to finish recording.

The tool will then display the transcribed text and the generated commands for your approval before executing them.

More Command Examples
"List all files in the current directory, including hidden ones, in a long format."

"Create a new git branch called 'feature/user-dashboard'."

"Find all markdown files in my documents folder that were modified in the last 24 hours."
