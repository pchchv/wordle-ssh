<div align="center">

# wordle-ssh

**Wordle, now over SSH.**

</div>

## How to play

You have 6 attempts to guess the correct word. Each guess must be a valid 5 letter
word.

After submitting a guess, the letters will turn green, yellow, or gray.

- **Green:** The letter is correct, and is in the correct position.
- **Yellow:** The letter is present in the solution, but is in the wrong position.
- **Gray:** The letter is not present in the solution.

### Running the server

```
go run main.go -s
```

### Running the client

```
go run main.go
```

### SSH connection

```
ssh localhost -p 1337
```
