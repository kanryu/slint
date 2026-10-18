# SlintCFlat Hike/Go Binding

This module provides Go bindings for the SlintCFlat API located in
`deps/slint-cflat-windows`, using Hike's `cfunc` and `.syso` generation. The
binding API does not expose cgo; the native packaging step uses Go's external
link mode when it must carry a Rust static archive and Windows system
libraries.

The public API is in the root package, while the Hike bridge for the SlintCFlat
ABI is in `internal/cflat`. The Hike source is
`internal/cflat/cflat.go.hike`.

```go
app, err := slint.New(`export component App inherits Window {}`)
if err != nil { panic(err) }
defer app.Close()
```

The `examples/hello` sample and the Makefile are in the project root. To
generate the files for the host OS, run:

Native libraries (`.a` and `.syso` files) are managed with Git LFS. Install Git
LFS and run `git lfs install` before cloning. If you have already cloned the
repository and the files are still LFS pointers, run `git lfs pull` to download
the actual files.

```text
make
```

This generates `cflat_windows_amd64.syso` on Windows;
`cflat_linux_amd64.syso`, `cflat_linux_arm64.syso`, and
`cflat_linux_arm32.syso` on Linux; and `cflat_darwin_arm64.syso` on macOS.

You can also select a target explicitly:

```text
make syso-windows-amd64
make syso-linux-arm64
make syso-linux-arm32
make syso-darwin-arm64
```

Override a target's static library with `SLINT_LIB_<target>`.

`hikepack` places external static libraries in separate `.syso` files, keeping
them separate from Hike-generated files when passing them to Go. Before
running the build, add `hikepack` to your `PATH` or specify its location, for
example:

```text
make HIKEPACK="C:/path/to/hikepack.exe" syso-windows-amd64
```

Note that Go's COFF linker cannot handle the overflow form used when a single
object contains more than 65,535 relocations. If you can rebuild the external
Slint library, split code generation into smaller units to keep the relocation
count per object below this limit. If a prebuilt archive exceeds this limit,
the `.syso` produced by `hikepack` may be readable by LLVM but unusable during
Go's final link.

The Makefile assumes an environment with POSIX commands such as `rm`. On
Windows, the executable output name automatically gets the `.exe` extension.

## Building a Hike binding with an external static library

The important build detail is that `hikec go` and external-library packaging
are separate operations. `hikec go` compiles `.go.hike` files into a Hike
`.syso` object and generates the Go declarations and assembly stubs. It does
not link a Rust or C static archive into that object. Do not append a static
archive to the output with a general-purpose `clang -r` command: Rust
`staticlib` archives contain their own runtime objects, Windows import
members, COFF directives, and relocation records. Flattening such an archive
can create a `.syso` that LLVM accepts but Go's internal COFF linker rejects.

Declare the foreign functions in Hike, and describe the native inputs in
`hike.mod` instead:

```text
package ./internal/cflat {
    target windows {
        # Keep the generated Hike object separate from the archive package.
        hike-object: "none"
        link: "deps/slint-cflat-windows/libslint_cflat.a"
        link: "-Wl,-u,slint_c_create"
        link: "-lgdi32"
        link: "-lopengl32"
        link: "-luser32"
        link: "-lkernel32"
    }
}
```

The file archive and the system-library flags have different roles. The
archive is preserved byte-for-byte and exposed through a generated native
package. The `-l...` entries are deferred to the final external link, because
absorbing Windows import libraries into the archive with `--whole-archive`
causes duplicate definitions. The `-Wl,-u,...` entry is useful to force the
public Slint entry point to remain reachable, but it is not passed through as
a cgo flag.

Build the package with `hikepack`, not directly with `hikec go`:

```text
hikepack syso -v -target windows .
```

`hikepack` performs the following steps:

1. Runs the Hike Go backend and creates the generated Go declarations and
   assembly stubs.
2. Reads the target-specific native configuration from `hike.mod`.
3. Keeps the external `.a` archive intact instead of flattening it into the
   Hike object.
4. Places the archive-backed `.syso` file in the package's native companion
   directory.
5. Generates the small platform-specific linker package that defers ordinary
   system-library flags to Go's final external link.

This separation is the key to using a Rust/Slint static library from Hike:
Hike owns the ABI bridge and Go stubs, while `hikepack` owns native archive
packaging. The Makefile invokes `hikepack syso` for each supported target and
therefore should not repeat the old `clang -r <hike-object> <library>.a`
step.

The official `slint-bundler` command, which combines Slint sources into one
file, is a Go program. Build it with:

```text
make slint-bundler
```

For example:

```text
bin/slint-bundler <slint-directory> -o appbundle.slint
```

## Rebuilding SlintCFlat for Windows GNU LLVM

The Windows static library must be built with the same GNU/LLVM object and
linker model that the Hike/Go package expects. A normal MSVC build, or a build
that reuses a different `rustc` wrapper, can produce an archive whose COFF
relocations cannot be consumed by the final Go link.

From the SlintCFlat source tree, configure the Rust toolchain explicitly before
building:

```powershell
Remove-Item Env:RUSTC_WRAPPER -ErrorAction SilentlyContinue

$env:RUSTC = "$env:USERPROFILE\.cargo\bin\rustc.exe"
$env:CARGO_TARGET_X86_64_PC_WINDOWS_GNULLVM_LINKER = `
  "C:\local\gnu\msys64\ucrt64\bin\clang.exe"
$env:RUSTFLAGS = "-C link-arg=-fuse-ld=lld"

rustup target add --toolchain stable x86_64-pc-windows-gnullvm
rustup run stable rustc -vV
rustup run stable cargo clean
rustup run stable cargo build `
  --release `
  --target x86_64-pc-windows-gnullvm
```

The important details are the `x86_64-pc-windows-gnullvm` target, the UCRT
Clang linker, and LLD selected through `RUSTFLAGS`. Run `cargo clean` after
changing these settings so that objects from a previous MSVC or GNU target are
not mixed into the new static archive. Copy the resulting target-specific
`libslint_cflat.a` and headers into `deps/slint-cflat-windows` before running
`hikepack syso` in this repository.

The MVC router sample is in `examples/mvc-router`. It does not use SQLite; it
displays a list of files in the current directory. The sample's Makefile uses
`slint-bundler` to combine the UI Slint files into `generated/ui_bundle.slint`,
which is then passed to `slint.New` from a Go string variable populated with
`go:embed`.

```text
cd examples/mvc-router
make bundle
go test .
```

## License

For the purposes of GitHub repository distribution, this project is licensed under **MIT License**.

In practice, this thin binding layer is intended to be treated functionally as an extension or part of Slint itself. You are free to use, link, and redistribute this wrapper under any licensing terms that you have agreed to with Slint (including the Slint Commercial License or alternative proprietary terms).