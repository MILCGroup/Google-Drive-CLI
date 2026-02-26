# Agent Guidelines

This file provides guidelines for AI agents (Claude Code, etc.) when working with this codebase.

## Testing Guidelines

### Don't Test What the Type System Already Checks

**Rule:** Do NOT add tests that simply verify what Go's type system and compiler already ensure.

**Examples of tests to AVOID:**

1. **Constant string value tests**
   ```go
   // BAD: Testing that a constant equals its literal value
   func TestActionType(t *testing.T) {
       if ActionUpload != "upload" {  // Compiler already ensures this
           t.Errorf("ActionUpload = %q, want %q", ActionUpload, "upload")
       }
   }
   ```

2. **Simple constructor tests**
   ```go
   // BAD: Testing that NewManager returns non-nil with nil inputs
   func TestNewManager(t *testing.T) {
       m := NewManager(nil, nil)
       if m == nil {
           t.Fatal("expected non-nil manager")
       }
   }
   ```

3. **Struct field assignment tests**
   ```go
   // BAD: Testing that assigned fields have the assigned value
   func TestActionStruct(t *testing.T) {
       action := Action{Path: "file.txt"}
       if action.Path != "file.txt" {  // This is just testing Go syntax
           t.Errorf("Path = %q, want %q", action.Path, "file.txt")
       }
   }
   ```

4. **Empty/nil slice tests without logic**
   ```go
   // BAD: Testing that nil slice is empty
   func TestConvertUsers_Nil(t *testing.T) {
       got := convertUsers(nil)
       if len(got) != 0 {  // This is a property of Go, not your code
           t.Errorf("expected empty slice")
       }
   }
   ```

**What TO test instead:**

- Error handling paths and edge cases
- Business logic and algorithms
- API interactions and response handling
- State mutations and side effects
- Complex conditional logic
- Integration between components
- Actual behavior that could break

**Rationale:**
- These "type system tests" don't catch real bugs
- They inflate coverage numbers without adding value
- They create maintenance burden when refactoring
- Go's compiler and type system are the correct tools for these checks

## Code Style

Follow existing patterns in the codebase. When in doubt, look at neighboring files.

## Pull Requests

- Ensure tests pass: `go test ./...`
- Check coverage meaningfully improved (not just inflated)
- Focus on testing behavior, not syntax
