# SlintCFlat Hike/Go binding

このモジュールは、`deps/slint-cflat-windows` に配置された SlintCFlat の
APIを、Hikeの `cfunc` と `.syso` 生成を経由してGoから利用するための
cgo-freeバインディングです。

公開APIはパッケージルート、SlintCFlat ABIのHikeブリッジは
`internal/cflat` に分離しています。Hikeソースは
`internal/cflat/cflat.go.hike` です。

```go
app, err := slint.New(`export component App inherits Window {}`)
if err != nil { panic(err) }
defer app.Close()
```

サンプルは `examples/hello`、Makefileはプロジェクト直下にあります。
ホストOSに応じた生成は次のコマンドで行います。

```text
make
```

Windowsでは `cflat_windows_amd64.syso`、Linuxでは
`cflat_linux_amd64.syso`、`cflat_linux_arm64.syso`、
`cflat_linux_arm32.syso`、macOSでは `cflat_darwin_arm64.syso` を生成します。

ターゲットを個別に指定することもできます。

```text
make syso-windows-amd64
make syso-linux-arm64
make syso-linux-arm32
make syso-darwin-arm64
```

各ターゲットの静的ライブラリは `SLINT_LIB_<target>` で上書きできます。

`hikepack` は外部静的ライブラリを `.syso` 拡張子の別ファイルとして配置し、
Hike生成物と分離したままGoへ渡します。実行前に `hikepack` を PATH に置くか、例えば
`make HIKEPACK="C:/path/to/hikepack.exe" syso-windows-amd64` のように指定してください。

なお、Go の COFF リンカーは一つのオブジェクト内のリロケーション数が 65535 を
超えるオーバーフロー形式を扱えません。Slint の外部ライブラリを再ビルドできる
場合は、コード生成単位を分割して各オブジェクトのリロケーション数を抑えてください。
配布済みアーカイブがこの制限を超える場合、`hikepack` が作る `.syso` は LLVM では
読めても Go の最終リンクでは使用できません。

MakefileはPOSIXコマンド（`rm` など）が使える環境を前提にしています。
Windowsでビルドする場合、実行ファイルの出力名には自動的に `.exe` が付きます。

Slintソースを1ファイルへ束ねる正式コマンド `slint-bundler` はGoプログラムです。
次のコマンドでビルドできます。

```text
make slint-bundler
```

実行例:

```text
bin/slint-bundler <slint-directory> -o appbundle.slint
```

MVCルーター構成のサンプルは `examples/mvc-router` です。SQLiteは使わず、
カレントディレクトリのファイル一覧を表示します。UIの各Slintファイルは
サンプル内のMakefileで `slint-bundler` により
`generated/ui_bundle.slint` へ結合し、Goの `go:embed` によるstring変数から
`slint.New` へ渡します。

```text
cd examples/mvc-router
make bundle
go test .
```
