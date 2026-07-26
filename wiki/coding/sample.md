# Go Code Sample

This page shows the expected Nuubot Go layout and coding style.

The purpose is readability. The user should know where each kind of code normally lives.

Canonical example: [`cmd/nuubot-bt-bot/main.go`](../../cmd/nuubot-bt-bot/main.go).

The three section markers are mandatory, even when one section has no functions.
Other guidance is strong by default and may bend when required or directed by the user.

```go
// Section 1 - Program Flow

func Run() error {
	// load input
	var input, err = loadInput()
	if err != nil {
		return err
	}

	// execute work
	err = execute(input)
	if err != nil {
		return err
	}
	return nil
}

// Section 2 - Domain Helpers

// Section 2.1 - PnL Calculations

// Section 3 - Generic Helpers
```

- Keep code clean and organized.
- Always keep Sections 1, 2, and 3, including empty sections.
- Group code and functions by intent.
- Use short verb-first intent comments.
- Leave one blank line before each intent comment.
- Reading comments in sequence should reveal the flow.
- Keep Program Flow lean.
- Put application mechanics in Domain Helpers.
- Put formatting, files, cleanup, and profiling in Generic Helpers.
- Prefer simple names. Qualify only when ambiguity exists.
- Use readable blocks before extracting unnecessary helpers.
- Do not over-comment syntax or every error branch.
