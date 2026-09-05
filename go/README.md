# Go トラック

`google.golang.org/api` 経由でGoogle Sheets APIをREST越しに叩くサンプル。GASと違い、OAuth2認証を自分で組む必要がある。

## 0. Goのインストール

未インストールの場合:

```powershell
winget install GoLang.Go
```

または https://go.dev/dl/ から手動インストール。インストール後、新しいシェルで `go version` が通ることを確認する。

## 1. Google Cloud Console 側の準備

1. https://console.cloud.google.com/ でプロジェクトを作成(または既存プロジェクトを選択)。
2. 「APIとサービス」→「ライブラリ」から **Google Sheets API** を有効化。
3. 「APIとサービス」→「OAuth同意画面」を設定 (External / Testing でよい。自分のGoogleアカウントをテストユーザーに追加)。
4. 「APIとサービス」→「認証情報」→「認証情報を作成」→「OAuthクライアントID」→ アプリケーションの種類は **デスクトップアプリ** を選択。
5. 作成されたクライアントのJSONをダウンロードし、`go/credentials.json` として保存する (このファイルはgitignore対象)。

## 2. 依存関係の取得

```powershell
cd go
go mod tidy
```

`golang.org/x/oauth2` と `google.golang.org/api` を取得し、`go.sum` が生成される。

## 3. 実行

```powershell
go run ./cmd/sheets <スプレッドシートID>
```

スプレッドシートIDはURLの `https://docs.google.com/spreadsheets/d/<ここ>/edit` 部分。

初回実行時はターミナルに認証URLが表示されるので、それをブラウザで開いてアクセスを許可する。
許可すると `http://127.0.0.1:<ランダムなポート>` へリダイレクトされ、サンプルが起動しているローカルサーバーがそれを受け取って認証を完了する(コードをコピペする必要はない)。
成功すると `go/token.json` にトークンとスコープがキャッシュされ、同じスコープでの再実行では再認証なしで動く。Docs/Gmailなど別のスコープを使うコマンドを追加した場合は、スコープが一致しないため自動的に再度ブラウザでの許可が求められる。

## ファイル構成

- `internal/auth/auth.go` — OAuth2認証の共通処理 (どのAPIのサンプルでも使い回す)。
- `cmd/sheets/main.go` — スプレッドシートの読み書きサンプル。

## 次のステップ

`cmd/docs/` `cmd/slides/` `cmd/gmail/` `cmd/drive/` を追加し、対応するAPIパッケージ
(`google.golang.org/api/docs/v1` など) と `auth.GetClient` に渡すスコープを変えていく。
