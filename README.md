# google-scripts

Google Workspace (スプレッドシート / ドキュメント / スライド / メール / ドライブ) を操作する方法を学ぶためのリポジトリ。
以下の2トラックを並行して試す。

- [`gas/`](gas/) — Google Apps Script (JavaScript)。Google純正のスクリプト環境で、認証不要ですぐ各サービスを叩ける。
- [`go/`](go/) — Go言語 + 公式クライアントライブラリ (`google.golang.org/api`)。OAuth2認証を自分で組み、外部プログラムからREST API経由で操作する。

## 進め方

1. まずは Sheets (スプレッドシート) から。両トラックにサンプルを用意済み。
2. 慣れたら Docs → Slides → Gmail → Drive の順に同じパターンで広げていく。
3. GASは「Workspace内で完結する自動化」、Goは「外部システムとの連携」に向いている、という違いを意識しながら比較する。

## 前提ツール

| ツール | 用途 | 状態 |
|---|---|---|
| Node.js | GASのローカル開発ツール `clasp` の実行に使用 | インストール済み (v20.19.1) |
| `@google/clasp` | GASプロジェクトをローカルで編集・pushするCLI | 未導入。[gas/README.md](gas/README.md) 参照 |
| Go | Goサンプルのビルド・実行 | 未インストール。https://go.dev/dl/ からインストールするか `winget install GoLang.Go` |

詳細な認証・実行手順は各ディレクトリのREADMEを参照。
