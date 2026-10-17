---
sidebar_position: 5
---

# Shared AI drivers

Built-in and external Go plugins can import `github.com/jmylchreest/tinct/pkg/ai`.
Drivers receive explicit options and requests. They do not read environment variables, mutate CLI flags, choose models, cache results or write files. Plugins retain those policies and colour extraction.

| Driver | Capability | Protocol |
|---|---|---|
| `OpenAI` | Image | `/images/generations` |
| `OpenAIText` | Text | `/chat/completions` |
| `Anthropic` | Text | `/messages` |
| `OpenRouter` | Image | Chat completions with image modalities |
| `MiniMax` | Image | `/image_generation` |
| `Google` | Image | Google SDK GenerateImages or GenerateContent |

`ImageGenerator` and `TextGenerator` are separate interfaces. Anthropic's driver does not claim native image generation. Models and capabilities are caller-selected; no model catalogue or paid requests are used to probe support.

## Local OpenAI-compatible image servers

```go
driver := ai.OpenAI{Options: ai.Options{
    BaseURL: "http://127.0.0.1:8080/v1",
    // APIKey is optional for local endpoints.
}}
image, err := driver.GenerateImage(ctx, ai.ImageRequest{
    Model: "your-image-model",
    Prompt: "A forest at dawn",
    Size: "1024x1024",
})
```

The base URL includes the server's API prefix; trailing slashes are accepted. The driver appends `/images/generations`. It accepts base64 image data or a downloadable image URL. It does not force a provider-specific response format. Image downloads do not receive the API key.

This supports local hosting when the server actually implements the Images API. An OpenAI-compatible text endpoint (including a llama.cpp or Ollama deployment) is not evidence that it implements image generation. Use `OpenAIText` for its chat text protocol; use the separate `OpenRouter` adapter only for servers implementing that image extension. Unsupported endpoints return errors rather than silently switching protocols. A future generic CLI input plugin can wrap this driver with endpoint/model/key flags without changing the transport.

## Policy and limits

`Options` accepts an injected `*http.Client`, API key, timeout and response limit. Defaults are five minutes per HTTP request and 64 MiB per HTTP response. Generation and a subsequent image download each receive that timeout; set a deadline on the caller context to bound the whole operation. Earlier context deadlines take precedence, including with injected clients. HTTP generation requests are not retried and redirects are rejected to avoid replaying requests or forwarding credentials. HTTP errors expose status codes through `*APIError`, without echoing response bodies. Provider errors inside successful responses are validated separately.

`Google` accepts a configured SDK client for Gemini or Vertex credentials and applies a five-minute context deadline. Its SDK controls response decoding, transport limits and retry policy; the HTTP Options settings do not configure Google. Supply only size and negative-prompt options supported by the selected Google model/backend.

`EnhanceForWallpaper(prompt, disabled)` and `BuildNegative(extra, suppress)` expose the existing prompt policy without CLI globals. Drivers do not add prompt enhancements implicitly. MiniMax and OpenRouter callers can append negative prompts as instructions; Google Imagen uses its dedicated field where supported. OpenAI's Images API does not receive a synthetic negative-prompt field.

Existing Google, OpenRouter and MiniMax input plugins use these drivers. Their flags, cache handling, metadata and palette extraction stay in the plugins. OpenAI and Anthropic are SDK capabilities in this change, not new registered CLI inputs.
