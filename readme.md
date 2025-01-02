## Axon

A go implementation of a client/server that takes in an Open API spec and can register it with claude desktop and allow you to talk to it

## Main Components:

1. An MCP server that handles the MCP communication (this is forked from the `mcp-go` anthropic server)
2. Handling for reading and parsing Open API/Swagger files
3. A main entry point for running the server

## Running it

Build the project using:

`go build -o ./bin/ ./cmd/`

Make sure that your `claude_desktop_config.json` file is correctly configured. Here is what mine looks like:

```json
{
  "mcpServers": {
    "apis": {
      "command": "/absolute/path/to/project/executable/axon/bin/axon",
      "args": [
        "/absolute/path/to/project/axon/example/specs/open_api/test-spec.json"
      ]
    }
  }
}
```

Then restart claude desktop and you should see the tools icon in the bottom right corner.

## Testing

I've included a test file and test server to make testing the MCP server easy. The test file is `test-spec.json`, this is the classic pet store Open API spec.

In the `/example` directory there is a sample go server that matches the pet store Open API spec and is seeded with a few rows of sample data. It just stores data in memory.

You can run this server by going to into the `/example` directory and running `go run main.go`. This will start the server at `http://localhost:3001`.

There are also other sample open_api and swagger specs in the `/example/specs` directory.
