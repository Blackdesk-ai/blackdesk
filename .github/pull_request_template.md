## Summary

- what changed
- why it changed

## Validation

- [ ] `gofmt -w` on changed Go files
- [ ] `go test ./...`
- [ ] public documentation updated when needed
- [ ] release label added when this PR should produce a versioned release

## Risk Review

- [ ] no secrets or local machine artifacts added
- [ ] no provider-specific payloads leaked into UI or domain code
- [ ] no screener data added to AI context payloads
