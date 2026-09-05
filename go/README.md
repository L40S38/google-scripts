# Go トラック

`google.golang.org/api` 経由でGoogle Sheets APIをREST越しに叩くサンプル。GASと違い、OAuth2認証を自分で組む必要がある。
Dockerコンテナ上で実行することを前提にしている(ホストにGoをインストールする必要はない)。ホストに直接Goを入れて動かしたい場合は末尾の「Dockerを使わない場合」を参照。

## 0. コンテナの起動

Docker Desktop (または Docker Engine) がインストール済みで `docker` コマンドが使えること。

OAuth認証はブラウザからのループバックコールバックを受け取る必要があるため、固定ポート(ここでは `8080`)をホストのループバック (`127.0.0.1`) にだけ公開し、コンテナ内はその固定ポートで待ち受けるよう `OAUTH_CALLBACK_ADDR` 環境変数で指定する。`-p` と `-e` のポート番号は必ず一致させる。8080が使用中なら、両方を別の空きポートに変える。

```powershell
# Windows (PowerShell)
docker run --rm -it `
  -p 127.0.0.1:8080:8080 `
  -e OAUTH_CALLBACK_ADDR=0.0.0.0:8080 `
  -v "${PWD}\go:/workspace" -w /workspace `
  golang:1.23 bash
```

```bash
# WSL / Linux
docker run --rm -it \
  -p 127.0.0.1:8080:8080 \
  -e OAUTH_CALLBACK_ADDR=0.0.0.0:8080 \
  -v "$(pwd)/go:/workspace" -w /workspace \
  golang:1.23 bash
```

以降の `go mod tidy` / `go run` は、このコンテナ内のシェルで実行する想定。`go/` をボリュームマウントしているので、`go.sum` や後で置く `credentials.json` / `token.json` はホスト側の `go/` にもそのまま残る。`exit` でコンテナを抜けると `--rm` により後片付けされ、ホストにはGo本体は一切インストールされない。

## 1. Google Cloud Console 側の準備

1. https://console.cloud.google.com/ でプロジェクトを作成(または既存プロジェクトを選択)。
2. 「APIとサービス」→「ライブラリ」から **Google Sheets API** を有効化。
3. 「APIとサービス」→「OAuth同意画面」に進む。初回は「Google Auth Platform はまだ構成されていません」という画面が出るので、そのまま「開始」を押して設定を進める。
   - **アプリ情報**: アプリ名(自由。ただし後述の制限に注意)、ユーザーサポートメール → 自分のGoogleアカウントのメールアドレス。
   - **対象(Audience)**: 個人のGoogleアカウントの場合は **External(外部)** を選ぶ(Internal はGoogle Workspace組織アカウント専用)。
   - **連絡先情報**: デベロッパーの連絡先メールアドレス → 自分のGoogleアカウントのメールアドレス。
   - **完了**: 「Google API Services User Data Policy」への同意にチェックして保存。
   - 設定が終わったら「APIとサービス」→「OAuth同意画面」→「対象」タブを開く。公開ステータスが **Testing** になっていることを確認し、同じ「対象」タブ内の**テストユーザー**セクションに自分のGoogleアカウントを追加する。
   - **アプリ名の制限**: 「Google」など既存サービス名を含む名前は弾かれる(ブランド名との混同を避けるため)。禁止例・許可例は[Googleのサポート記事](https://support.google.com/cloud/answer/15549049#app-name)を参照。例えば `google-scripts-study` は弾かれるため、`gas-go-scripts-study` のように名前を変える。
4. 「APIとサービス」→「認証情報」→「クライアント」タブを開き、上部の3点リーダー(⋮)メニュー→「クライアントを作成」→ アプリケーションの種類は **デスクトップアプリ** を選択。
5. 作成されたクライアントのJSONをダウンロードし、`go/credentials.json` として保存する (このファイルはgitignore対象)。コンテナ内からは `/workspace/credentials.json` として見える。

## 2. 依存関係の取得

コンテナ内のシェルで:

```
go mod tidy
```

`golang.org/x/oauth2` と `google.golang.org/api` を取得し、`go.sum` が生成される(ホスト側の `go/go.sum` にも反映される)。

## 3. 実行

コンテナ内のシェルで:

```
go run ./cmd/sheets <スプレッドシートID>
```

スプレッドシートIDはURLの `https://docs.google.com/spreadsheets/d/<ここ>/edit` 部分。
このコマンドは指定したスプレッドシートの `A1:B2` を固定値で無条件に上書きする。既存データが失われるため、必ず使い捨て用に新規作成したスプレッドシートのIDを渡すこと。

初回実行時はターミナルに認証URLが表示されるので、それをホスト側のブラウザで開いてアクセスを許可する。
許可すると `http://127.0.0.1:8080`(`-p` で公開した固定ポート)へリダイレクトされ、コンテナ内で待ち受けているサーバーがそれを受け取って認証を完了する(コードをコピペする必要はない)。
成功すると `go/token.json` にトークンとスコープがキャッシュされ、同じスコープでの再実行では再認証なしで動く。Docs/Gmailなど別のスコープを使うコマンドを追加した場合は、スコープが一致しないため自動的に再度ブラウザでの許可が求められる。

## ファイル構成

- `internal/auth/auth.go` — OAuth2認証の共通処理 (どのAPIのサンプルでも使い回す)。`OAUTH_CALLBACK_ADDR` 環境変数でループバックコールバックの待受アドレスを上書きできる(未指定時は `127.0.0.1:0` = ホストのランダムなポートで待ち受け)。
- `cmd/sheets/main.go` — スプレッドシートの読み書きサンプル。

## 次のステップ

`cmd/docs/` `cmd/slides/` `cmd/gmail/` `cmd/drive/` を追加し、対応するAPIパッケージ
(`google.golang.org/api/docs/v1` など) と `auth.GetClient` に渡すスコープを変えていく。

## Dockerを使わない場合(ホストに直接Goをインストール)

コンテナを使わず、Goをホスト(またはWSL)に直接インストールして動かすこともできる。この場合 `OAUTH_CALLBACK_ADDR` は設定しなくてよく、自動的にランダムなポートの `127.0.0.1` で待ち受ける。`docker run ... bash` で入るかわりに、以下のコマンドをホスト(またはWSL)のシェルで直接実行する。

### Windows

未インストールの場合、管理者権限があれば:

```powershell
winget install GoLang.Go
```

管理者権限がない環境では、zip配布版をユーザーディレクトリに展開するだけで使える(インストーラー不要):

```powershell
$sdkDir = "$env:LOCALAPPDATA\go-sdk"
mkdir $sdkDir -Force
curl.exe -L -o "$sdkDir\go.zip" https://go.dev/dl/go1.27.0.windows-amd64.zip  # 最新版は https://go.dev/dl/ で確認
Expand-Archive "$sdkDir\go.zip" -DestinationPath $sdkDir -Force
$env:GOROOT = "$sdkDir\go"
$env:PATH = "$env:GOROOT\bin;$env:PATH"
go version
```

この `$env:PATH` の変更は現在のシェルセッション限り。恒久的に使うなら `$PROFILE` に同じ2行を追記するか、ユーザー環境変数(こちらも管理者権限不要)の PATH に `$sdkDir\go\bin` を追加する。

または https://go.dev/dl/ から手動インストール。インストール後、新しいシェルで `go version` が通ることを確認する。

> **トラブルシューティング**: `go build` / `go mod tidy` で `package weak is not in std` のようなエラーが出た場合、`%LOCALAPPDATA%\go-build` のビルドキャッシュが別のGoインストール試行(失敗したwinget実行など)で壊れている可能性がある。`$env:GOCACHE` を新しい空のディレクトリ(例: `$sdkDir\gocache`)に向けて再実行する。

### WSL / Linux

apt でインストールする場合(バージョンがやや古いことがある):

```bash
sudo apt update && sudo apt install -y golang-go
```

最新版が欲しい場合、公式のtar.gz配布版をユーザーディレクトリに展開する(root不要):

```bash
sdk_dir="$HOME/go-sdk"
mkdir -p "$sdk_dir"
curl -L -o "$sdk_dir/go.tar.gz" https://go.dev/dl/go1.27.0.linux-amd64.tar.gz  # 最新版は https://go.dev/dl/ で確認
tar -C "$sdk_dir" -xzf "$sdk_dir/go.tar.gz"
export GOROOT="$sdk_dir/go"
export PATH="$GOROOT/bin:$PATH"
go version
```

`export` の2行は現在のシェルセッション限りなので、恒久的に使うなら `~/.bashrc` (または `~/.zshrc`) に追記する。

WindowsとWSLは別のファイルシステム・別のGo環境として扱われるため、`go/credentials.json` や `go/token.json` の置き場所、`go mod tidy` で生成される `go.sum` はどちらか一方の環境で作業を統一する(両方で並行して `go mod tidy` を実行しても内容は同じになるはずだが、混在させる必要はない)。
