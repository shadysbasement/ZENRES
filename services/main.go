// Vercel detects this entrypoint before running the build command in vercel.json.
// The build command compiles auth-api/cmd/server, which owns the actual HTTP server.
package main

func main() {
	panic("build the auth API with the command in vercel.json")
}
