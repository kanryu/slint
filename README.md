# SlintCFlat Hike/Go Binding

## Important: install the Slint module locally with Git LFS

Do not install this module with the usual `go get` workflow. Go may fetch the
repository through the module proxy without downloading the large Git LFS
objects, leaving files such as `cflat_windows_amd64.syso` as LFS pointers.
`GOPROXY=direct` alone is not sufficient because Go still fetches the Git
repository rather than the GitHub Release ZIP.

Instead, fetch the module directly with Git LFS and use it as a local module:

```bash
git lfs install
git clone --depth 1 --branch v0.1.1 --single-branch https://github.com/kanryu/slint.git third_party/slint
git -C third_party/slint lfs pull

go mod edit -replace github.com/kanryu/slint=./third_party/slint
go mod tidy
go build ./...
```

This bypasses `proxy.golang.org` and ensures that the native
`cflat_windows_amd64.syso` file is available as a real binary. Verify the
module selection with:

```text
go list -m all
```

The output should include the following line, confirming that the local module
is being used:

```text
github.com/kanryu/slint => ./third_party/slint
```

## What is Slint?

Slint is a modern, general-purpose GUI framework for building user interfaces
that run on desktop, mobile, web, and embedded platforms. It is implemented in
Rust and is designed to provide a native user experience across operating
systems and processor architectures while keeping the UI description concise
and declarative.

For the official project information, visit the [Slint website](https://slint.dev)
and the [official Slint Git repository](https://github.com/slint-ui/slint).

Instead of constructing every widget from ordinary application code, you
describe windows, controls, layout, properties, and bindings in Slint's own
markup language. A `.slint` file can look like this:

```slint
export component HelloWindow inherits Window {
    width: 400px;
    height: 240px;

    Text {
        text: "Hello from Slint";
        horizontal-alignment: center;
        vertical-alignment: center;
    }
}
```

Slint keeps the UI description separate from the application's business logic.
The framework can compile Slint markup ahead of time, optimize its property
expressions, and use configurable rendering backends. It also provides an
interpreter, which is the mode used by this binding: the application supplies
the Slint source at startup and the Slint runtime creates the UI from it.

For Hike language programs, this project provides the binding needed to use
that runtime. The binding does not replace the Slint language. Instead, it
helps an application collect multiple `.slint` files,
combine them into one deterministic source file, embed that source in the
executable as a Go string, and pass it to `slint.New` when the application
starts.

The Hike compiler is a build-time tool in this architecture. It generates the
native `.syso` objects, Go declarations, and assembly stubs required by the
binding; it is not a runtime dependency that changes what Slint is or how the
application's UI is described.

The typical flow is:

```text
individual .slint files
        │
        ├─ slint-bundler
        ▼
one bundled .slint source
        │
        ├─ go:embed
        ▼
embedded string in the executable
        │
        └─ slint.New(source)
        ▼
Slint interpreter and native window
```

This makes it possible to organize a larger UI into reusable pages and
components while still shipping one self-contained Hike/Go application.

This module provides a Hike language binding for the SlintCFlat API. Its native
bridge is implemented with Hike `cfunc` declarations and generated Go
declarations/assembly stubs. The binding API does not expose cgo. Windows
amd64 packages its Rust static archive as a `.syso`; Linux amd64 links
`deps/slint-cflat-linux-amd64/libslint_cflat.a` copied into its Linux
`internal/cflat_native` package under the platform `.syso` name.

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

This generates `cflat_windows_amd64.syso` on Windows. Linux amd64 copies its
archive to
`internal/cflat_native/cflat_linux_amd64.syso`
and runs Hike's Go backend.

You can also select a target explicitly:

```text
make syso-windows-amd64
make syso-linux-amd64
make syso-linux-arm64
make syso-linux-arm32
make syso-darwin-arm64
```

Override a Hikepack-backed target's static library with `SLINT_LIB_<target>`.

On Linux amd64, build the Hello example with `make -C examples/hello`. The
Makefile copies the release archive into `internal/cflat_native` under its
platform `.syso` name, runs `hikec go` to generate the bridge, and builds the
executable. The corresponding `target linux` entry in `hike.mod` records the
release archive and Linux system libraries. A working `hikec` installation is
required.

## Building and running on Linux amd64

Install Go, Git LFS, GNU Make, Hike (`hikec`), and a C compiler/linker supported
by cgo. The build also needs the development files for the native libraries
listed in `hike.mod`, including Fontconfig. Fetch the native archive from Git
LFS and build the Hello example:

```sh
git lfs install
git lfs pull
make -C examples/hello build
make -C examples/hello run
```

The native library is packaged from
`deps/slint-cflat-linux-amd64/libslint_cflat.a` by the `syso-linux-amd64`
target. The Hello Makefile invokes this target automatically before building.

Running the application requires a graphical desktop session with a supported
X11 or Wayland display backend, as well as the shared libraries required by
Slint. In particular, `libxkbcommon-x11.so.0` is needed at runtime. If the
application fails to start because a shared library is missing, inspect its
dependencies with:

```sh
ldd examples/hello/hello
```

Install the missing runtime packages using your distribution's package manager.
For example, on Debian or Ubuntu, `libxkbcommon-x11-0` provides
`libxkbcommon-x11.so.0`; the Fontconfig runtime library is also required.
Runtime package names vary by distribution. Linux amd64 is the Linux
configuration documented and verified for this binding.

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

For Go programmers using this binding, the project includes SlintBinder, the
`slint-bundler` command. Slint UI is often easier to maintain as multiple files
under a `ui` directory: a router, a window component, reusable widgets, and
page components. SlintBinder reads that directory, resolves the imports between
the files, and writes one deterministic `.slint` source file.

The application can then embed the generated file and give its contents to the
Slint interpreter at startup. This keeps the UI organized as a normal folder
tree during development while allowing the final Go application to carry one
self-contained Slint source string.

The SlintBinder command is implemented in Go. Build it with:

```text
make slint-bundler
```

With Go 1.24 or later, register the GitHub module as a tool dependency with:

```text
go get -tool github.com/kanryu/slint/cmd/slint-bundler@latest
```

Then build and run the tool with:

```text
go tool slint-bundler
```

The `go get -tool` command records the tool in `go.mod`; `go tool` builds it
from the recorded module dependency when it runs. If you want to install a
standalone executable in Go's standard binary directory instead, use:

```text
go install github.com/kanryu/slint/cmd/slint-bundler@latest
```

For example:

```text
bin/slint-bundler <slint-directory> -o appbundle.slint
```

## Building on Windows amd64

Install Go, Git LFS, GNU Make, Hike (`hikec`), and `hikepack`. For the external
Go link, install the MSYS2 UCRT64 LLVM toolchain, including Clang and LLD. Make
sure `hikec`, `hikepack`, `make`, and the UCRT64 tools are available from the
shell's `PATH`.

After cloning or updating the repository, fetch the native libraries managed by
Git LFS. In PowerShell, set `CC` to the UCRT64 Clang executable and build the
Hello example:

```powershell
git lfs install
git lfs pull

$env:CC = "C:\msys64\ucrt64\bin\clang.exe"
$env:CGO_ENABLED = "1"

make syso-windows-amd64
make -C examples/hello build
```

The first command runs `hikepack syso -v -target windows .`, which reads the
Windows linker configuration from `hike.mod`, generates the Hike bridge, and
packages the Windows static archive. The Hello Makefile runs this native step
automatically, so the explicit `make syso-windows-amd64` is optional when
building Hello. To launch the sample, run `make -C examples/hello run`.

If Hike tools are not on `PATH`, pass their full paths to Make:

```powershell
make HIKEC="C:/path/to/hikec.exe" HIKEPACK="C:/path/to/hikepack.exe" syso-windows-amd64
```

The Makefiles use POSIX utilities such as `rm` and `mkdir`; run them from an
MSYS2/UCRT64 shell or another environment that provides GNU Make and these
utilities. The native target currently supports Windows amd64.

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
