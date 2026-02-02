# 🔧 Tasker - Terminal CLI Todo Manager

A C++ command-line application to parse and organize your todos from a text file into an interactive terminal interface.

## Features

- 📖 **Read todos** from a text file (location stored in environment variables)
- 🎨 **Color-coded display** for different task statuses (Done, Postponed, Pending)
- 🔍 **Filter tasks** by date or status
- 🔎 **Search tasks** by keyword
- 📊 **View statistics** including completion rate and tasks by date
- 💻 **Simple CLI** interface with intuitive commands

## Requirements

- C++17 compiler (GCC, Clang, or MSVC)
- CMake 3.10 or higher

## Installation

### 1. Clone/Navigate to the project directory

```bash
cd c:\Users\fabni\Work\personal\tasker
```

### 2. Set up the environment variable

You need to set the `TODO_FILE_PATH` environment variable to point to your todo.txt file.

**On Windows (PowerShell):**
```powershell
# Temporary (current session only)
$env:TODO_FILE_PATH = "C:\path\to\your\todo.txt"

# Permanent (user-level)
[System.Environment]::SetEnvironmentVariable('TODO_FILE_PATH', 'C:\path\to\your\todo.txt', 'User')
```

**On Windows (Command Prompt):**
```cmd
# Temporary
set TODO_FILE_PATH=C:\path\to\your\todo.txt

# Permanent
setx TODO_FILE_PATH "C:\path\to\your\todo.txt"
```

**On Linux/macOS:**
```bash
# Temporary
export TODO_FILE_PATH="/path/to/your/todo.txt"

# Permanent (add to ~/.bashrc or ~/.zshrc)
echo 'export TODO_FILE_PATH="/path/to/your/todo.txt"' >> ~/.bashrc
source ~/.bashrc
```

### 3. Build the project

```bash
cmake -B build
cmake --build build
```

The executable will be created in the `build` directory.

## Todo File Format

Your `todo.txt` file should follow this format:

```
22/1/2025
1. Onway redis - Done
2. Polymorphism c++ - Postponed
3. Nicomencian ethics - Done


23/1/2025

1. Polymorphism c++ - Postponed
2. Plausly - Done
3. Tenex - Done
```

**Format rules:**
- Date headers: `DD/M/YYYY` or `DD/MM/YYYY`
- Tasks: `<number>. <description> - <status>`
- Status: `Done`, `Postponed`, `Pending`, or anything else
- Empty lines are ignored

## Usage

### Show all tasks
```bash
tasker list
```

### View statistics
```bash
tasker stats
```

### Filter by date
```bash
tasker filter date 23/1/2025
```

### Filter by status
```bash
tasker filter status done
tasker filter status postponed
tasker filter status pending
```

### Search by keyword
```bash
tasker search polymorphism
tasker search redis
```

### Show help
```bash
tasker help
```

## Project Structure

```
tasker/
├── include/          # Header files
│   ├── Task.h
│   ├── TodoParser.h
│   ├── TodoManager.h
│   └── Display.h
├── src/              # Source files
│   ├── main.cpp
│   ├── Task.cpp
│   ├── TodoParser.cpp
│   ├── TodoManager.cpp
│   └── Display.cpp
├── tests/            # Test files
│   └── sample_todo.txt
├── CMakeLists.txt    # Build configuration
└── README.md         # This file
```

## Learning Notes

This project demonstrates several C++ concepts:

- **File I/O**: Reading from files and environment variables
- **String Parsing**: Extracting structured data from text
- **STL Containers**: Using `vector`, `map` for data storage
- **STL Algorithms**: Using `copy_if`, `count_if` for filtering
- **Object-Oriented Design**: Classes with clear responsibilities
- **Enum Classes**: Type-safe status representation
- **Lambda Functions**: For filtering and searching predicates

## Troubleshooting

### "TODO_FILE_PATH environment variable not set"
- Make sure you've set the environment variable (see Installation step 2)
- Restart your terminal/PowerShell after setting the variable
- Verify with: `echo $env:TODO_FILE_PATH` (PowerShell) or `echo %TODO_FILE_PATH%` (CMD)

### "Could not open file"
- Check that the path in `TODO_FILE_PATH` is correct
- Ensure the file exists at that location
- Check file permissions

### Colors not showing on Windows
- Use Windows Terminal or PowerShell 7+ for ANSI color support
- Older CMD may not display colors correctly

## License

This is a learning project. Feel free to use and modify as needed.