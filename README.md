# VoiceTerm AI 🎤

**VoiceTerm** is a smart command-line interface (CLI) tool that transforms your spoken language into executable shell commands.  
It acts as your personal voice assistant for the terminal — designed to streamline your workflow and reduce typing.

---

## 🚀 What It Does

VoiceTerm listens for a voice command, intelligently converts it into the correct shell command or sequence of commands, and executes them for you.  
It's perfect for complex, multi-step operations like Git workflows or file management.

### ✨ Features

- **Voice-to-Command:** Simply speak what you want to do, like  
  > "commit my changes with the message 'bug fix' and push to the develop branch."
- **AI-Powered:** Uses state-of-the-art AI to understand your intent and generate accurate commands.
- **Interactive & Safe:** Shows you the generated commands before running them and asks for any missing details (e.g., filename, branch name).

---

## ⚙️ How It Works

The tool follows a simple four-step process:

1. **🎙️ Record Audio:**  
   Uses the **SoX** utility to capture audio from your microphone when you press `Enter`.

2. **✍️ Transcribe to Text:**  
   The recorded audio is sent to **OpenAI’s Whisper API**, which transcribes your speech into text with high accuracy (English).

3. **🧠 Generate Commands:**  
   The transcribed text is then processed by **OpenAI’s Chat API (gpt-4o)**.  
   A specialized prompt converts your request into shell commands.  
   Missing details are represented as placeholders (e.g., `[COMMIT_MESSAGE]`).

4. **🚀 Execute Commands:**  
   The app parses and displays the generated commands for review.  
   If placeholders exist, you’ll be prompted to fill them before execution.  
   Finally, the commands are executed sequentially in your shell.

---

## 🧩 Technology Stack

| Component | Description |
|------------|-------------|
| **Go** | Core programming language used for the CLI |
| **Cobra** | Framework for building modern CLI tools in Go |
| **SoX (Sound eXchange)** | Cross-platform utility for audio recording |
| **OpenAI Whisper API** | High-accuracy voice-to-text transcription |
| **OpenAI Chat API (gpt-4o)** | Natural language understanding and command generation |

---

## 🔧 Prerequisites

Before installing, make sure you have:

### 🧱 Homebrew (for macOS or Linux)
If not installed, get it from [brew.sh](https://brew.sh).

### 🎙️ SoX (Sound eXchange)
Essential for recording audio.

**Install via Homebrew:**
```bash
brew install sox
```

**On Debian/Ubuntu:**
```bash
sudo apt-get update && sudo apt-get install sox libsox-fmt-all
```

### 🔑 OpenAI API Key
You need an active API key from the [OpenAI Platform](https://platform.openai.com/).

Add your key to your shell config file (`~/.zshrc`, `~/.bash_profile`, etc.):

```bash
export OPENAI_API_KEY='your-api-key-here'
```

Then refresh your terminal session:

```bash
source ~/.zshrc
```

---

## 🧭 Installation

Once your Homebrew tap is published, install **voiceterm** easily:

**Step 1: Tap the repository**
```bash
# Replace YOUR_USERNAME with your GitHub username
brew tap YOUR_USERNAME/voiceterm
```

**Step 2: Install the tool**
```bash
brew install voiceterm
```

---

## 💻 Usage Example

Using **voiceterm** is simple and intuitive.

1. Open your terminal.
2. Run:
   ```bash
   voiceterm
   ```
3. The tool will prompt:
   ```
   🎤 Press Enter to start recording, press Enter again to stop...
   ```
4. Press `Enter` → You’ll see:
   ```
   🔴 Recording... Press Enter to stop.
   ```
5. Speak clearly, e.g.:
   > "Stage all my changes, commit with the message 'refactor api module', and then push to the main branch."
6. Press `Enter` again to finish.
7. The tool displays:
   - Transcribed text  
   - Generated commands  
   You can confirm before execution.

---

## 🧠 More Command Examples

- “List all files in the current directory, including hidden ones, in a long format.”
- “Create a new git branch called ‘feature/user-dashboard’.”
- “Find all markdown files in my documents folder that were modified in the last 24 hours.”

---

**VoiceTerm AI** — Turn your voice into command-line power. ⚡
