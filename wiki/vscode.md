# VS Code

## Go Error Highlighting

The global owner is:

```text
C:\Users\PC\AppData\Roaming\Code\User\settings.json
```

Add this entry inside the top-level `"highlight.regexes"` object:

```json
"(if (?:err|[A-Za-z][A-Za-z0-9]*Err) != nil[^\\r\\n\\{]*\\{[^\\}]*\\})": [
    {
        "color": "rgba(144, 144, 144, 0.40)",
        "fontWeight": "normal"
    }
]
```

The rule covers canonical `err`, necessary `*Err` values, and same-line
compound conditions.
