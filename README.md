# google-scripts

Google Workspace (スプレッドシート / ドキュメント / スライド / メール / ドライブ) やGemini APIを操作する方法を学ぶためのリポジトリ。
以下のトラックを並行して試す。

- [`gas/`](gas/) — Google Apps Script (JavaScript)。Google純正のスクリプト環境で、OAuthクライアントの設定なしで各サービスを叩ける(初回実行時の権限承認は必要)。
- [`go/`](go/) — Go言語 + 公式クライアントライブラリ (`google.golang.org/api`)。OAuth2認証を自分で組み、外部プログラムからREST API経由で操作する。
- [`gemini/`](gemini/) — Gemini API (Interactions API)。APIキーのみで使え、OAuth/GCPプロジェクトの紐付けは不要。Go(メイン) / Python / Node.js のサンプルを用意。

## 進め方

1. まずは Sheets (スプレッドシート) から。両トラックにサンプルを用意済み。
2. 慣れたら Docs → Slides → Gmail → Drive の順に同じパターンで広げていく。
3. GASは「Workspace内で完結する自動化」、Goは「外部システムとの連携」に向いている、という違いを意識しながら比較する。

## 前提ツール

| ツール | 用途 | インストール方法 |
|---|---|---|
| Docker | 各トラックの実行(コンテナ上で動かすためNode.js/clasp/Go/Python自体のホストへのインストールは不要)。詳細は [gas/README.md](gas/README.md) / [go/README.md](go/README.md) / [gemini/README.md](gemini/README.md) 参照 | https://www.docker.com/products/docker-desktop/ から導入 |

詳細な認証・実行手順は各ディレクトリのREADMEを参照。
