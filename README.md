# viassh

`viassh` is a simple demo library for creating remote dialer using a chain of one or more SSH servers.  Authentication is done via the `ssh-agent`.

## Example usage

```go
import "github.com/borud/viassh"

// ...

dialer, err := viassh.Create(viassh.Config{
    Hosts:  []string{"user@example.com:22", "user@foo.com:22"},
})

// ...

conn, err := dialer.Dial("tcp","amazon.com:443")
```

See also the examples directory.
