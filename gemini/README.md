# Gemini API トラック

Gemini API(Interactions API)を使ったテキスト生成・動画読解のサンプル。`gas/` / `go/` のOAuth(同意画面・GCPプロジェクトの紐付け等)とは違い、**APIキーだけで使える**ため認証周りのセットアップが大幅に簡単。
Dockerコンテナ上で実行することを前提にしている(ホストにGo/Python/Node.jsをインストールする必要はない)。

複数言語でサンプルを用意している: **Go**(メイン) / **Python** / **Node.js**(どちらもサンプルコードのみ、要動作確認)。

## 0. APIキーの取得

1. https://aistudio.google.com/apikey を開き、GoogleアカウントでログインしてAPIキーを作成する(`gas`/`go`トラックのようなGCPプロジェクト作成・OAuth同意画面の設定は不要)
2. リポジトリルートで `gemini/.env` を作成し、以下の1行を書く(ファイル名はgitignore対象かつ、Claudeからは読めないディレクトリ/ファイル名として扱われるよう `.claude/settings.local.json` でブロック済み。`go/credentials.json` / `go/token.json` と同じ扱い)

```
GEMINI_API_KEY=取得したAPIキー
```

各言語の `docker run` では `--env-file` オプションでこのファイルをそのままコンテナに渡す(ホストの環境変数にエクスポートする必要はない)。

## 1. Go(メイン)

公式Go SDK (`google.golang.org/genai`) が存在する。ただし `ai.google.dev` のクイックスタート/テキスト生成ガイドにはGoタブがなく、ファイル(画像・動画)を扱うガイドにのみGoの例が載っている、という珍しい状態になっている。そのため:

- [gemini/go/cmd/text/main.go](go/cmd/text/main.go) — 基本のテキスト生成のみ。SDKのGoタブが存在しないため、標準ライブラリの `net/http` でREST APIを直接呼んでいる(外部依存なし)。
- [gemini/go/cmd/video/main.go](go/cmd/video/main.go) — 動画読解の実験用(後述)。こちらは公式SDK `google.golang.org/genai` を使用。

```powershell
# Windows (PowerShell)
docker run --rm -it `
  --env-file gemini/.env `
  -v "${PWD}\gemini\go:/workspace" -w /workspace `
  golang:1.25 bash
```

```bash
# WSL / Linux
docker run --rm -it \
  --env-file gemini/.env \
  -v "$(pwd)/gemini/go:/workspace" -w /workspace \
  golang:1.25 bash
```

> **注意**: `google.golang.org/genai` が `go >= 1.24` を要求するため、`go/` / `gas/` トラックで使っている `golang:1.23` ではなく `golang:1.25` イメージを使う(`go.mod` も `go 1.24` にしてある)。古いイメージだと `go mod tidy` が `toolchain upgrade needed ... GOTOOLCHAIN=local` で失敗する。

コンテナ内のシェルで、まずテキスト生成(外部依存なしなので `go mod tidy` 不要):

```bash
go run ./cmd/text
```

`Explain how AI works in a few words` に対するGeminiの応答テキストが標準出力に表示される。

> **注意**: このリポジトリの検証時、Docker Desktopが起動していなかったためコンパイル確認ができていない(未検証)。詰まったら教えてほしい。

### 動画読解の実験 (`cmd/video`)

Gemini APIがどこまで動画の内容を読み取れるか試すサンプル。フリーライセンスの動画 **Big Buck Bunny**(2008年、Blender Foundation制作、[Creative Commons Attribution 3.0](https://peach.blender.org/about/)。再配布・商用利用可、要クレジット表記)を使う。

コンテナ内のシェルで、サンプル動画を取得(本家配布元 download.blender.org から取得。約65MBのZIPで、解凍後は320x180・約10分のmp4):

```bash
apt-get update && apt-get install -y unzip
curl -L -o bbb.zip "https://download.blender.org/peach/bigbuckbunny_movies/BigBuckBunny_320x180.mp4.zip"
unzip bbb.zip && rm bbb.zip
```

依存パッケージを取得してから実行:

```bash
go mod tidy
go run ./cmd/video
```

要約・登場キャラクター・印象的な場面のタイムスタンプ・音の種類をGeminiに答えさせる(プロンプトは [main.go](go/cmd/video/main.go) 参照)。動画は数分で処理されるため、初回はアップロード後に「処理中...」が何度か表示される。

> **注意**: `google.golang.org/genai` のドキュメントを元に実装したが、Docker Desktopが起動していない状態での作業だったためコンパイル・実行の確認ができていない(未検証)。`go mod tidy` や `go run` でエラーが出たら教えてほしい。
>
> 別の動画を試したい場合は `go run ./cmd/video <path/to/video.mp4>` のように引数でパスを渡せる(ライセンスは別途確認すること)。

## 2. Python(サンプル)

[gemini/python/main.py](python/main.py)。公式SDK `google-genai` を使用。

```bash
# WSL / Linux (Windowsは -v "${PWD}\gemini\python:/workspace" に読み替え)
docker run --rm -it \
  --env-file gemini/.env \
  -v "$(pwd)/gemini/python:/workspace" -w /workspace \
  python:3.12 bash
```

コンテナ内のシェルで:

```bash
pip install -r requirements.txt
python main.py
```

> **注意**: このサンプルは未検証(実機での動作確認はまだ行っていない)。詰まったら教えてほしい。

## 3. Node.js(サンプル)

[gemini/node/index.mjs](node/index.mjs)。公式SDK `@google/genai` を使用。`gas/` トラックと同じ `node:20` イメージを流用できる。

```bash
# WSL / Linux (Windowsは -v "${PWD}\gemini\node:/workspace" に読み替え)
docker run --rm -it \
  --env-file gemini/.env \
  -v "$(pwd)/gemini/node:/workspace" -w /workspace \
  node:20 bash
```

コンテナ内のシェルで:

```bash
npm install
node index.mjs
```

> **注意**: このサンプルは未検証(実機での動作確認はまだ行っていない)。詰まったら教えてほしい。

## ファイル構成

- `go/cmd/text/main.go` — テキスト生成。REST直叩き(SDKにGoタブがないため)。
- `go/cmd/video/main.go` — 動画読解の実験。公式SDK `google.golang.org/genai` を使用。
- `python/main.py` — `google-genai` SDKを使ったテキスト生成サンプル。
- `node/index.mjs` — `@google/genai` SDKを使ったテキスト生成サンプル。

## 次のステップ

- Go/Python/Node.jsサンプルの実機動作確認(このリポジトリでの作業時はDocker Desktopが起動しておらず未検証)
- マルチターン対話(`previous_interaction_id`)、ストリーミング、画像生成など他の機能も試す
- `gas/` / `go/` トラックで触った Sheets API の結果をGeminiに渡して要約させる、といった連携サンプル
