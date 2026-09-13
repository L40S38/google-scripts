# Google Apps Script (GAS) トラック

Google純正のスクリプト実行環境。OAuthクライアントの設定なしで `SpreadsheetApp` / `DocumentApp` / `SlidesApp` / `GmailApp` / `DriveApp` を呼び出せる(初回実行時にはGoogleの権限承認ダイアログでの同意が必要)。
ローカルでコードを書いて `clasp` でGoogle側にpush/pullする。
Dockerコンテナ上で実行することを前提にしている(ホストにNode.js/claspをインストールする必要はない)。ホストに直接入れて動かしたい場合は末尾の「Dockerを使わない場合」を参照。

## 0. コンテナの起動

`clasp login` はブラウザからのループバックコールバックを受け取る。固定ポート(ここでは `8081`。go/トラックで使っている `8080` とは別のポート)をホストのループバックにだけ公開する。ログイン情報 (`~/.clasprc.json`) をコンテナを作り直しても失わないよう、名前付きボリュームをホームディレクトリにマウントする。

> **注意**: 以下の `docker run` はリポジトリの**ルートディレクトリ**(`google-scripts/`)で実行する。`gas/` の中で実行すると `-v "$(pwd)/gas:/workspace"` が存在しない `gas/gas` を指してしまい、Dockerがその空ディレクトリを自動生成して(中身が空の)`/workspace` にマウントしてしまう(`.clasp.json` や `appsscript.json` が見えない状態になる)。

初回のみボリュームを作成:

```bash
docker volume create clasp-home
```

```powershell
# Windows (PowerShell)
docker run --rm -it `
  -p 127.0.0.1:8081:8081 `
  -v clasp-home:/root `
  -v "${PWD}\gas:/workspace" -w /workspace `
  node:20 bash
```

```bash
# WSL / Linux
docker run --rm -it \
  -p 127.0.0.1:8081:8081 \
  -v clasp-home:/root \
  -v "$(pwd)/gas:/workspace" -w /workspace \
  node:20 bash
```

コンテナ内のシェルで:

```bash
npm install -g @google/clasp
```

> **注意**: `--rm` 付きコンテナは終了すると同時に破棄されるため、`npm install -g` でインストールした `clasp` 本体(`/usr/local/lib/node_modules` 以下)は永続化されない。`clasp-home` ボリューム(`/root`)や `gas/` のbindマウント(`/workspace`)には影響しないため、**新しいコンテナを起動するたびに毎回この `npm install -g @google/clasp` を実行し直す**必要がある(`clasp: command not found` になった場合はこれが原因)。

`gas/` をボリュームマウントしているので、生成される `.clasp.json` などはホスト側の `gas/` にもそのまま残る。`clasp-home` ボリュームに保存されたログイン情報は、同じ名前のボリュームを指定して起動する限り、コンテナを作り直しても引き継がれる(完全に消したい場合は `docker volume rm clasp-home`)。

## 1. ログイン

コンテナ内のシェルで:

```bash
clasp login --redirect-port 8081
```

表示された認証URLをホスト側のブラウザで開いて許可する。許可すると `http://127.0.0.1:8081` へリダイレクトされ、公開したポート経由でコンテナ内のclaspが認証を完了する。

### 新規プロジェクトの作成

`clasp create` はリモート側に新規作成したプロジェクトのデフォルトファイル(`appsscript.json` など)を書き込む。`/workspace`(=ホストの `gas/`)に直接向けると、チェックイン済みで `oauthScopes` を設定済みの `appsscript.json` がデフォルト内容で上書きされてしまうため、コンテナ内の別ディレクトリで作成してから `.clasp.json` だけを `/workspace` に移す。

`/tmp/gas-create` には `.clasp.json` の他に `appsscript.json`(と場合により `Code.gs`)もpullされるが、これらは移さずに `rm -rf` で捨ててよい。理由は次の2点:

- `clasp create` が生成する `appsscript.json` はGoogle側のデフォルト内容(`timeZone` / `dependencies` / `exceptionLogging` / `runtimeVersion` のみ)で、`oauthScopes` が設定されていない。これで `gas/appsscript.json` を上書きすると、スプレッドシート等へのアクセスに必要な `oauthScopes` の設定が失われてしまう。
- `.clasp.json` に必要なのは新規作成された `scriptId` のひも付けだけ。ファイルの中身自体はこの後の `clasp push` で `gas/`(リポジトリ側)の `appsscript.json` / `src/*.js` がリモートへ送られ、デフォルト内容を正しい設定で上書きするため、`/tmp/gas-create` 側のファイルは不要になる。

> **注意**: clasp 3.2.0以降はパストラバーサル対策(CVE-2026-4092)により、`--rootDir` に実行時のカレントディレクトリの外を指定すると `Security Error: Content directory "..." resolves outside the project root "..."` で拒否される。`--rootDir` オプションは使わず、先に対象ディレクトリへ `cd` してから実行する(cwd自体がproject rootとして扱われるため一致する)。

> **注意**: `clasp create` (または `clasp clone`) が `User has not enabled the Apps Script API. Enable it by visiting https://script.google.com/home/usersettings then retry.` で失敗する場合、Googleアカウント側で Apps Script API が無効になっている(プロジェクト単位ではなくアカウント単位の設定)。ホスト側のブラウザで https://script.google.com/home/usersettings を開き「Google Apps Script API」をONにし、数分待ってから再実行する。このコマンドが失敗すると `.clasp.json` は作られないため、次の `mv` は成功を確認してから実行する。

```bash
mkdir -p /tmp/gas-create
cd /tmp/gas-create
clasp create --type standalone --title "gas-go-scripts-study"
mv .clasp.json /workspace/.clasp.json
cd /workspace
rm -rf /tmp/gas-create
```

`.clasp.json` (scriptIdを含む、gitignore対象) が `gas/`(ホスト側)に置かれ、以降の `clasp push` / `clasp open-script` はこのプロジェクトに対して行われる。
既存のApps Scriptプロジェクトに紐付けたい場合は同様に `cd /tmp/gas-create && clasp clone <scriptId>` を実行し、`.clasp.json` だけを `/workspace` に移す。

> **補足**: このワークフローは実際のGoogleアカウントでのログイン完了までは検証済みだが、`clasp create` 以降の細部(フラグの挙動など)は実機での動作確認がまだなので、詰まったら教えてほしい。

### コードの反映と実行

```bash
clasp push
clasp open-script
```

コンテナ内にはブラウザがないため、`open-script` はブラウザを起動しようとした後に `Open https://script.google.com/d/<scriptId>/edit in your browser to continue.` のようにURLを表示するだけになる(動作確認済み)。そのURLをホスト側のブラウザで開く。
エディタ上で関数を選んで実行する(推奨)。`clasp run <関数名>` はCLIから直接実行できて便利そうに見えるが、実際には次のすべてが必要になり手間が大きいため、通常はエディタ実行で十分。

- スクリプトに紐づくGCPプロジェクトを、デフォルトではなく明示的な「標準GCPプロジェクト」に切り替え、その `projectId` を `.clasp.json` に追加する
- `appsscript.json` に `"executionApi": { "access": "MYSELF" }` を追加する
- エディタから `Deploy > New deployment` で種類 `API Executable` としてデプロイする
- その標準GCPプロジェクトで自分用のOAuthクライアント(`client_secret.json`)を作成し、`clasp login --use-project-scopes --creds client_secret.json` で再ログインする

参考: [clasp run のドキュメント](https://github.com/google/clasp/blob/master/docs/run.md)

エディタ上で関数を実行すると、初回は「このアプリは Google で確認されていません」という権限確認モーダルが出る。これは審査(verification)未申請のアプリでは想定内の挙動で、開発者本人のアカウントで実行する限りは問題ない(第三者に配布する場合のみ審査が必要)。モーダル内の「詳細」→「(プロジェクト名)に移動(安全ではないページ)」の順にクリックし、続く権限確認画面で許可する。

> **注意**: エディタの「実行」ボタンは選択した関数を引数なしで呼び出す。`readExistingSheet(spreadsheetId)` のように引数を取る関数を直接実行すると `Exception: Invalid argument: id` のようなエラーになる(`spreadsheetId` が `undefined` のまま呼ばれるため)。動作確認するには次のいずれかを行う。
>
> - ラッパー関数を作って固定値を渡す。例: `function testReadExistingSheet() { readExistingSheet('createAndReadSheetのログに出たスプレッドシートID'); }` を定義し、これを実行する。
> - `clasp run readExistingSheet --params '["<spreadsheetId>"]'` を使う(上記の追加セットアップが必要になるため、今の段階ではラッパー関数の方が手軽)。

## ファイル構成

- `appsscript.json` — マニフェスト。使用するAPIスコープを宣言する。
- `src/sheets.js` — スプレッドシートの読み書きサンプル。

## 次のステップ

`src/` に `docs.js` / `slides.js` / `gmail.js` / `drive.js` を追加し、`appsscript.json` の `oauthScopes` に対応するスコープを足していく。

## Dockerを使わない場合(ホストに直接Node.js/claspをインストール)

コンテナを使わず、Node.js/claspをホスト(またはWSL)に直接インストールして動かすこともできる。この場合ポートを固定する必要はなく、`clasp login`(`--redirect-port` なし)は自動的にランダムなポートの `127.0.0.1` で待ち受ける。

```powershell
npm install -g @google/clasp
clasp login
```

ブラウザでGoogleアカウントにログインし、clasp用のOAuth許可を行う。

### 新規プロジェクトの作成

以下はリポジトリのルートディレクトリから実行する想定のコマンド。

```powershell
cd gas
mkdir ../gas-create-tmp
cd ../gas-create-tmp
clasp create --type standalone --title "gas-go-scripts-study"
cd ..
Move-Item gas-create-tmp/.clasp.json gas/.clasp.json
Remove-Item -Recurse -Force gas-create-tmp
cd gas
```

`.clasp.json` (scriptIdを含む、gitignore対象) が `gas/` に置かれ、以降の `clasp push` / `clasp open-script` はこのプロジェクトに対して行われる。
既存のApps Scriptプロジェクトに紐付けたい場合は同様に一時ディレクトリで `clasp clone <scriptId>` を実行し、`.clasp.json` だけを `gas/` に移す。

### コードの反映と実行

```powershell
clasp push
clasp open-script
```

エディタ上で関数を選んで実行する(推奨)。`clasp run <関数名>` を使う場合の追加セットアップはDocker版の「コードの反映と実行」節を参照。
