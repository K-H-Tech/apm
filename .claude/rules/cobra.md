# Cobra CLI Best Practices

## Overview

Cobra is a library for creating powerful modern CLI applications in Go. It's used by Kubernetes, Hugo, GitHub CLI, and many other popular tools.

## Installation

```bash
go get -u github.com/spf13/cobra@latest
go install github.com/spf13/cobra-cli@latest
```

## Project Structure

### Recommended Layout

```
myapp/
├── cmd/
│   ├── root.go          # Root command
│   ├── serve.go         # serve subcommand
│   ├── migrate.go       # migrate subcommand
│   └── version.go       # version subcommand
├── internal/            # Application logic
├── main.go              # Entry point
└── go.mod
```

### Main Entry Point

```go
// main.go
package main

import "myapp/cmd"

func main() {
    cmd.Execute()
}
```

## Command Definition

### Root Command

```go
// cmd/root.go
package cmd

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var (
    cfgFile string
    verbose bool
)

var rootCmd = &cobra.Command{
    Use:   "myapp",
    Short: "A brief description of your application",
    Long: `A longer description that spans multiple lines.

Include examples and detailed explanation here.`,
    // Uncomment if you have a default action
    // Run: func(cmd *cobra.Command, args []string) { },
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}

func init() {
    cobra.OnInitialize(initConfig)

    // Persistent flags (available to all subcommands)
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.myapp.yaml)")
    rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

    // Bind to viper
    viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

func initConfig() {
    if cfgFile != "" {
        viper.SetConfigFile(cfgFile)
    } else {
        home, err := os.UserHomeDir()
        cobra.CheckErr(err)

        viper.AddConfigPath(home)
        viper.SetConfigType("yaml")
        viper.SetConfigName(".myapp")
    }

    viper.AutomaticEnv()

    if err := viper.ReadInConfig(); err == nil {
        if verbose {
            fmt.Println("Using config file:", viper.ConfigFileUsed())
        }
    }
}
```

### Subcommand

```go
// cmd/serve.go
package cmd

import (
    "fmt"

    "github.com/spf13/cobra"
)

var (
    port int
    host string
)

var serveCmd = &cobra.Command{
    Use:   "serve",
    Short: "Start the server",
    Long:  `Start the HTTP server with the specified configuration.`,

    // Aliases for convenience
    Aliases: []string{"s", "server"},

    // Example usage
    Example: `  myapp serve --port 8080
  myapp serve -p 3000 -H 0.0.0.0`,

    RunE: func(cmd *cobra.Command, args []string) error {
        fmt.Printf("Starting server on %s:%d\n", host, port)
        // Start server...
        return nil
    },
}

func init() {
    rootCmd.AddCommand(serveCmd)

    // Local flags (only for this command)
    serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "port to listen on")
    serveCmd.Flags().StringVarP(&host, "host", "H", "localhost", "host to bind to")

    // Mark required
    // serveCmd.MarkFlagRequired("port")
}
```

## Flags

### Flag Types

```go
// Boolean
cmd.Flags().BoolVarP(&debug, "debug", "d", false, "enable debug mode")

// String
cmd.Flags().StringVarP(&name, "name", "n", "", "your name")

// Integer
cmd.Flags().IntVarP(&count, "count", "c", 1, "number of items")

// String slice
cmd.Flags().StringSliceVarP(&tags, "tag", "t", []string{}, "tags to apply")

// Duration
cmd.Flags().DurationVarP(&timeout, "timeout", "T", 30*time.Second, "operation timeout")

// Count (increments each time flag is used)
cmd.Flags().CountVarP(&verbosity, "verbose", "v", "verbosity level (-v, -vv, -vvv)")
```

### Persistent vs Local Flags

```go
// Persistent: Available to this command AND all subcommands
rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")

// Local: Available only to this specific command
serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "port number")
```

### Required Flags

```go
cmd.Flags().StringVarP(&name, "name", "n", "", "your name")
cmd.MarkFlagRequired("name")

// Required together
cmd.MarkFlagsRequiredTogether("username", "password")

// Mutually exclusive
cmd.MarkFlagsMutuallyExclusive("json", "yaml")

// At least one required
cmd.MarkFlagsOneRequired("source", "stdin")
```

### Flag Groups

```go
// Flags that must be used together
cmd.MarkFlagsRequiredTogether("cert", "key")

// Flags that cannot be used together
cmd.MarkFlagsMutuallyExclusive("format-json", "format-yaml")
```

## Arguments

### Argument Validation

```go
// No arguments allowed
cmd.Args = cobra.NoArgs

// Exactly N arguments
cmd.Args = cobra.ExactArgs(2)

// Minimum N arguments
cmd.Args = cobra.MinimumNArgs(1)

// Maximum N arguments
cmd.Args = cobra.MaximumNArgs(5)

// Range of arguments
cmd.Args = cobra.RangeArgs(1, 3)

// Custom validation
cmd.Args = func(cmd *cobra.Command, args []string) error {
    if len(args) < 1 {
        return errors.New("requires at least one argument")
    }
    if !isValidName(args[0]) {
        return fmt.Errorf("invalid name: %s", args[0])
    }
    return nil
}
```

### Positional Arguments

```go
var createCmd = &cobra.Command{
    Use:   "create [name]",
    Short: "Create a new resource",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        name := args[0]
        // Use name...
        return nil
    },
}
```

## Command Lifecycle Hooks

### PreRun and PostRun

```go
var cmd = &cobra.Command{
    Use: "mycommand",

    // Runs before Run for this command only
    PreRun: func(cmd *cobra.Command, args []string) {
        fmt.Println("PreRun: Setting up...")
    },

    // Main execution
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("Running main command...")
    },

    // Runs after Run for this command only
    PostRun: func(cmd *cobra.Command, args []string) {
        fmt.Println("PostRun: Cleaning up...")
    },
}

// Persistent versions (inherited by subcommands)
var rootCmd = &cobra.Command{
    PersistentPreRun: func(cmd *cobra.Command, args []string) {
        // Runs before every command
        initLogging()
    },
    PersistentPostRun: func(cmd *cobra.Command, args []string) {
        // Runs after every command
        flushLogs()
    },
}
```

### Error-Returning Hooks

```go
var cmd = &cobra.Command{
    PreRunE: func(cmd *cobra.Command, args []string) error {
        if err := validateConfig(); err != nil {
            return err
        }
        return nil
    },
    RunE: func(cmd *cobra.Command, args []string) error {
        // Main logic with error handling
        return doWork()
    },
}
```

## Shell Completion

### Built-in Completion

```go
// Automatically available:
// myapp completion bash
// myapp completion zsh
// myapp completion fish
// myapp completion powershell
```

### Custom Completions

```go
var createCmd = &cobra.Command{
    Use: "create [type]",
    ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
        if len(args) != 0 {
            return nil, cobra.ShellCompDirectiveNoFileComp
        }

        completions := []cobra.Completion{
            {Value: "user", Description: "Create a new user"},
            {Value: "group", Description: "Create a new group"},
            {Value: "project", Description: "Create a new project"},
        }

        return completions, cobra.ShellCompDirectiveNoFileComp
    },
}

// Flag completion
cmd.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    return []string{"json", "yaml", "table"}, cobra.ShellCompDirectiveNoFileComp
})
```

## Output and Formatting

### Structured Output

```go
type outputFormat string

const (
    formatJSON  outputFormat = "json"
    formatYAML  outputFormat = "yaml"
    formatTable outputFormat = "table"
)

var format outputFormat

func init() {
    cmd.Flags().StringVarP((*string)(&format), "output", "o", "table", "output format (json, yaml, table)")
}

func output(data interface{}) error {
    switch format {
    case formatJSON:
        enc := json.NewEncoder(os.Stdout)
        enc.SetIndent("", "  ")
        return enc.Encode(data)
    case formatYAML:
        out, err := yaml.Marshal(data)
        if err != nil {
            return err
        }
        fmt.Println(string(out))
    case formatTable:
        // Print as table
        printTable(data)
    }
    return nil
}
```

### Error Output

```go
func RunE(cmd *cobra.Command, args []string) error {
    if err := doSomething(); err != nil {
        // Use cmd.ErrOrStderr() for error output
        fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
        return err
    }

    // Use cmd.OutOrStdout() for normal output
    fmt.Fprintln(cmd.OutOrStdout(), "Success!")
    return nil
}
```

## Documentation Generation

### Markdown Docs

```go
import "github.com/spf13/cobra/doc"

func generateDocs() error {
    return doc.GenMarkdownTree(rootCmd, "./docs")
}
```

### Man Pages

```go
import "github.com/spf13/cobra/doc"

func generateManPages() error {
    header := &doc.GenManHeader{
        Title:   "MYAPP",
        Section: "1",
    }
    return doc.GenManTree(rootCmd, header, "./man")
}
```

## Testing

### Command Testing

```go
func TestServeCommand(t *testing.T) {
    // Reset command for testing
    cmd := NewServeCmd()

    // Set arguments
    cmd.SetArgs([]string{"--port", "3000"})

    // Capture output
    buf := new(bytes.Buffer)
    cmd.SetOut(buf)
    cmd.SetErr(buf)

    // Execute
    err := cmd.Execute()

    assert.NoError(t, err)
    assert.Contains(t, buf.String(), "3000")
}

func TestRootCommand(t *testing.T) {
    tests := []struct {
        name    string
        args    []string
        wantErr bool
    }{
        {"no args", []string{}, false},
        {"help", []string{"--help"}, false},
        {"invalid flag", []string{"--invalid"}, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cmd := NewRootCmd()
            cmd.SetArgs(tt.args)
            err := cmd.Execute()

            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## Best Practices

### 1. Use RunE for Error Handling

```go
// Good: Return errors
RunE: func(cmd *cobra.Command, args []string) error {
    if err := process(); err != nil {
        return fmt.Errorf("processing failed: %w", err)
    }
    return nil
}

// Avoid: Using os.Exit directly
Run: func(cmd *cobra.Command, args []string) {
    if err := process(); err != nil {
        fmt.Println(err)
        os.Exit(1)  // Makes testing difficult
    }
}
```

### 2. Keep Commands Thin

```go
// Good: Command just orchestrates
RunE: func(cmd *cobra.Command, args []string) error {
    cfg := config.Load()
    svc := service.New(cfg)
    return svc.Execute(args[0])
}

// Avoid: Business logic in command
Run: func(cmd *cobra.Command, args []string) {
    // 200 lines of business logic...
}
```

### 3. Consistent Flag Naming

```go
// Use kebab-case for multi-word flags
--output-format  // Good
--outputFormat   // Avoid
--output_format  // Avoid

// Use single letter shortcuts for common flags
-v (verbose)
-o (output)
-f (file)
-c (config)
```

### 4. Helpful Error Messages

```go
if len(args) == 0 {
    return errors.New("missing required argument: name\n\nUsage: myapp create <name>")
}
```

## Anti-Patterns to Avoid

1. **Don't use global state** - Pass dependencies through closures or constructors
2. **Don't call os.Exit in commands** - Return errors instead
3. **Don't ignore SilenceUsage** - Set it to prevent usage on errors
4. **Don't hardcode paths** - Use flags or config for paths
5. **Don't skip argument validation** - Use Args validators

```go
// Set this to avoid printing usage on every error
cmd.SilenceUsage = true
```
