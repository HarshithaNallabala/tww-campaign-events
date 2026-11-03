# AI Usage Disclosure

## 1. AI Tools Used
- **ChatGPT:** Used for understanding assignment requirements, discussing implementation approaches, debugging concepts, reviewing Go code, and drafting documentation.

## 2. How AI Was Used
- Clarified the requirements and identified important assumptions for the event ingestion API.
- Discussed handling malformed events, duplicate event IDs, validation, and campaign statistics.
- Used AI assistance to understand Go concepts such as JSON decoding, `json.RawMessage`, mutexes, HTTP handlers, and testing.
- Received assistance in identifying and fixing bugs in the debugging program.
- Used AI assistance to draft `NOTES.md`, `BUGS.md`, and this disclosure.

## 3. AI Suggestions That Were Changed
An initial approach of decoding the complete request directly into a slice of `Event` structs was reconsidered. I changed the implementation to decode the request into `[]json.RawMessage` and process each event individually.

This approach allows individual malformed or invalid events to be rejected without automatically rejecting every other valid event in the batch. I reviewed the implementation and tested the endpoint.

## 4. What I Learned
- How to implement HTTP endpoints and JSON request handling in Go.
- Why validating individual events is important for robust batch ingestion.
- How mutexes help protect shared state in concurrent programs.
- How duplicate event IDs affect idempotency.
- How to write tests using Go's `net/http/httptest` package.
- How to reason about event aggregation, debugging, and scalability trade-offs.

## 5. AI-Generated Code That I Tested or Debugged
- Reviewed and tested the event ingestion and campaign statistics implementation.
- Ran `go test ./...` successfully for the service.
- Identified and fixed four bugs in the debugging program.
- Ran the debugging program repeatedly and compared its output against `expected_output.txt`.
- Verified that the debugging output matched the expected output.

## 6. Limitations
AI assistance was used as a development aid, not as a substitute for understanding the implementation. I reviewed the code, ran the available tests, and verified the debugging output. Some additional production-level features and testing were intentionally left out, as documented in the project notes.
