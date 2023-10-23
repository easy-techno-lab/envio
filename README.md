# envio

![https://img.shields.io/github/v/tag/easy-techno-lab/envio](https://img.shields.io/github/v/tag/easy-techno-lab/envio)
![https://img.shields.io/github/license/easy-techno-lab/envio](https://img.shields.io/github/license/easy-techno-lab/envio)

`envio` is a library designed to get/set environment variables to/from go structures.

## Installation

`envio` can be installed like any other Go library through `go get`:

```console
go get github.com/easy-techno-lab/envio
```

Or, if you are already using
[Go Modules](https://github.com/golang/go/wiki/Modules), you may specify a version number as well:

```console
go get github.com/easy-techno-lab/envio@latest
```

## Getting Started

```go
package main

import (
	"fmt"
	"os"

	"github.com/easy-techno-lab/envio"
)

type Type struct {
	// will be got/set by name 'A'
	A string
	// `env:"ENV_B"` - will be got/set by name 'ENV_B'
	B string `env:"ENV_B"`
	// `env:"-"` - will be skipped
	C string `env:"-"`
	// `env:"ENV_D,m"` - will be got/set by name 'ENV_D',
	// mandatory field - if the environment doesn't contain a variable with the specified name,
	// it returns an error: the required variable $name is missing
	D string `env:"ENV_D,m"`
}

func main() {
	in := new(Type)
	in.A = "fa"
	in.B = "fb"
	in.C = "fc"
	in.D = "fd"

	err := envio.Set(in)
	if err != nil {
		panic(err)
	}

	fmt.Printf("A: %s; ENV_B: %s; ENV_D: %s\n", os.Getenv("A"), os.Getenv("ENV_B"), os.Getenv("ENV_D"))
	// A: fa; ENV_B: fb; ENV_D: fd

	out := new(Type)

	err = envio.Get(out)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", out)
	// &{A:fa B:fb C: D:fd}
}

```
