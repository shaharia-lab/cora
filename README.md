# Cora

<p align="center">
  <img src="https://github.com/shaharia-lab/cora/assets/1095008/9ce070b3-6377-42e8-bac5-516f3a8387fd" alt="Cora Logo">
</p>

**CORA**: **CO**ncatenate and **R**ead **A**ll

**Cora** is a command-line tool for concatenating files in a directory into a single output file. The name CORA stands for **"COncatenate and Read All"** reflecting its primary function of combining multiple files while providing options to include or exclude specific files based on patterns.

## Features

- Recursively walk through directories
- Include or exclude files based on glob patterns
- Customize separators between concatenated files
- Add prefixes to file paths in the output
- Debug mode for detailed logging

## Use Cases

Cora can be incredibly useful in various scenarios, particularly when dealing with large codebases or multiple files that need to be combined. Here are some key use cases:

1. **LLM Context Preparation**: When working with Large Language Models (LLMs) like GPT-3 or GPT-4, providing comprehensive context is crucial for accurate responses. Cora can concatenate an entire codebase into a single file, making it easy to input as context for LLMs. This is especially useful for:
   - Code review and analysis
   - Generating documentation
   - Answering questions about complex codebases

2. **Documentation Generation**: Combine multiple markdown files or source code files to create comprehensive documentation for your project.

3. **Code Auditing**: Merge multiple source files into a single document for easier review and analysis, especially when working with security auditing tools.

4. **Project Submissions**: Combine all relevant files for project submissions in academic or professional settings.

5. **Backup and Archiving**: Create a single file containing all important documents from a directory structure, making it easier to backup or share entire projects.

6. **Log Analysis**: Concatenate multiple log files for comprehensive analysis, while using exclude patterns to filter out irrelevant files.

7. **Content Management**: Combine multiple content pieces (e.g., blog posts, articles) into a single file for bulk editing or publishing.

8. **Data Preprocessing**: Merge multiple data files into a single file for easier processing in data analysis pipelines.

By using Cora, you can streamline these processes, saving time and reducing the complexity of managing multiple files in various scenarios.

## Difference between `cat` and `Cora`

While the `cat` command is indeed useful for simple file concatenation, Cora offers several advanced features that make it more powerful and flexible for complex scenarios.

| Feature | `cat` | `cora` |
|---------|-------|--------|
| Basic file concatenation | ✅ | ✅ |
| Recursive directory traversal | ❌ | ✅ |
| Flexible file selection (glob patterns) | ❌ | ✅ |
| Exclude patterns | ❌ | ✅ |
| Custom separators between files | ❌ | ✅ |
| File path prefixes in output | ❌ | ✅ |
| Built-in debugging mode | ❌ | ✅ |
| Cross-platform consistency | ❌ (behavior may vary) | ✅ |
| Large file handling | ✅ (but may require additional tools) | ✅ (optimized) |
| Speed for simple concatenations | ✅ (generally faster) | ✅ (may have slight overhead) |
| Requires external tools for complex tasks | ✅ (often used with find, xargs, etc.) | ❌ (all-in-one solution) |
| Customizable output file | ❌ (requires output redirection) | ✅ (direct specification) |
| Part of standard Unix toolset | ✅ | ❌ (requires installation) |

## Installation

To install Cora, make sure you have Go installed on your system, then run:

```bash
go install github.com/shaharia-lab/cora@latest
```

This command will download the source code, compile it, and install the `cora` binary in your `$GOPATH/bin` directory. Make sure your `$GOPATH/bin` is added to your system's `PATH` to run `cora` from any location.

## Usage

After installation, you can run Cora from anywhere in your terminal:

```bash
cora [flags]
```

### Flags

- `-s, --source`: Source directory to concatenate files from (required)
- `-o, --output`: Output file to write concatenated files to (required)
- `-e, --exclude`: Glob patterns to exclude (can be specified multiple times)
- `-i, --include`: Glob patterns to include (can be specified multiple times)
- `-d, --debug`: Enable debugging mode
- `-p, --separator`: Separator to use between concatenated files (default: "\n---\n")
- `-x, --path-prefix`: Prefix to add before the path of included files (default: "## ")

### Example

```bash
cora -s /path/to/source -o output.md -e "*.log" -i "*.md" -i "*.txt" -d
```

This command will concatenate all `.md` and `.txt` files from `/path/to/source`, excluding any `.log` files, and save the result to `output.md` with debug logging enabled.

## Development

### Prerequisites

- Go 1.16 or higher

### Building

To build the project locally, run:

```bash
go build
```

### Running Tests

To run the tests, use:

```bash
go test ./...
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Cobra](https://github.com/spf13/cobra) - A Commander for modern Go CLI interactions
- [testify](https://github.com/stretchr/testify) - A toolkit with common assertions and mocks that plays nicely with the standard library

## Contact

Shaharia Lab OÜ - [shaharialab.com](https://shaharialab.com) - hello@shaharialab.com

Project Link: [https://github.com/shaharia-lab/cora](https://github.com/shaharia-lab/cora)