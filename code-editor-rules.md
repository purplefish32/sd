# [SD ]

Every time you choose to apply a rule(s), explicitly state the rule(s) in the output. You can abbreviate the rule description to a single word or phrase.
Please use emojis to make the rules more readable.

## Project Context

[This project is a replacement for the Elgato Stream Deck software. ]

- [It will be primarily written in Golang and HTMX]
- [It will run primarily in Linux, but eventually will be cross-platform]
- [The primary Linux distribution will be Archlinux]
- [The primary window manager will be i3wm]
- [The primary terminal will be Ghostty]
- [The server will communicate via USB with the Stream Decks]
- [The First Stream Decks to be supported will be the Elgato Stream Deck XL, Elgato the Stream Deck Plus and the Elgato Stream Deck Pedal. Other models will be supported later.]

## Code Style and Structure

- Write concise, technical Golang code with accurate examples
- Use functional and declarative programming patterns
- Prefer iteration and modularization over code duplication
- Apply the principles of SOLID design principles
- Apply Golang best practices

## Tech Stack

- Golang
- Templ
- HTMX
- NATS.io
- Tailwind CSS

## Naming Conventions

- Follow best practices for Golang
- Follow best practices for HTMX
- Follow best practices for Tailwind CSS

## Golang Usage

- You MUST use the `templ` package for templating
- You MUST use github.com/rs/zerolog/log for logging
- You MUST use "github.com/karalabe/hid" for HID devices

## NATS.io Usage

- kv.Keys() is depricated use kv.ListKeys() instead

## HTMX Usage

- When using HTMX ssr refer to this page https://v1.htmx.org/extensions/server-sent-events/
- The ssr endpoint MUST send HTML and not JSON
- The ssr endpoint MUST send the event name in the format of "event: event-name"
- The ssr endpoint MUST send the data in the format of "data: data"
- The ssr endpoint MUST send a blank line after the data

## State Management

- Application state will be managed using NATS.io Key Value Store

## Syntax and Formatting

## UI and Styling

- Implement Tailwind CSS for styling
- Keep the UI simple and minimalistic
- Limit the use of colors
- Use of colors should be linked to semantic meaning unless it is a brand color
- Limit the number of the colors
- Have consistent colors throughout the UI
- Anything dangerous or destructive to data should be red
- Anything positive should be green
- Anything neutral should be grey
- Anything informative should be blue
- Anything warning should be orange
- Prefer using a variant of a color if needed (Lighter, Darker, etc.)

## Error Handling

- Implement proper error handling
- Handle network failures gracefully
- Provide user-friendly error messages

## Logging

- Log errors appropriately for debugging

## Testing

- Testing is not the priority of this project
- Testing will be done manually

## Security

- Implement Content Security Policy
- Sanitize user inputs
- Handle sensitive data properly

## Git Usage

Commit Message Prefixes:

- "fix:" for bug fixes
- "feat:" for new features
- "perf:" for performance improvements
- "docs:" for documentation changes
- "style:" for formatting changes
- "refactor:" for code refactoring
- "test:" for adding missing tests
- "chore:" for maintenance tasks

Rules:

- Use lowercase for commit messages
- Keep the summary line concise
- Include description for non-obvious changes
- Reference issue numbers when applicable

## Documentation

- Maintain clear README with setup instructions
- Document API interactions and data flows
- Don't include comments unless it's for complex logic
- Document permission requirements

## Development Workflow

- Use proper version control
- Implement proper code review process
- Test in multiple environments
- Follow semantic versioning for releases
- Maintain changelog

## Rule suggestions

- If you find that a subject is not covered by the rules but should be, propose to add it to the rules. But don't add it yourself.
