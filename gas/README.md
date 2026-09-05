# Google Apps Script (GAS) トラック

Google純正のスクリプト実行環境。OAuthクライアントの設定なしで `SpreadsheetApp` / `DocumentApp` / `SlidesApp` / `GmailApp` / `DriveApp` を呼び出せる(初回実行時にはGoogleの権限承認ダイアログでの同意が必要)。
ローカルでコードを書いて `clasp` でGoogle側にpush/pullする。

## セットアップ

```powershell
npm install -g @google/clasp
clasp login
```

ブラウザでGoogleアカウントにログインし、clasp用のOAuth許可を行う。

### 新規プロジェクトの作成

`clasp create` はリモート側に新規作成したプロジェクトのデフォルトファイル(`appsscript.json` など)をカレントディレクトリに書き込む。`gas/` で直接実行すると、チェックイン済みで `oauthScopes` を設定済みの `gas/appsscript.json` がデフォルト内容で上書きされてしまうため、空の一時ディレクトリで作成してから `.clasp.json` だけを `gas/` に移す。

```powershell
mkdir ../gas-create-tmp
cd ../gas-create-tmp
clasp create --type standalone --title "google-scripts-study"
cd ..
Move-Item gas-create-tmp/.clasp.json gas/.clasp.json
Remove-Item -Recurse -Force gas-create-tmp
cd gas
```

`.clasp.json` (scriptIdを含む、gitignore対象) が `gas/` に置かれ、以降の `clasp push` / `clasp open` はこのプロジェクトに対して行われる。
既存のApps Scriptプロジェクトに紐付けたい場合は同様に一時ディレクトリで `clasp clone <scriptId>` を実行し、`.clasp.json` だけを `gas/` に移す。

### コードの反映と実行

```powershell
clasp push
clasp open   # スクリプトエディタをブラウザで開く
```

エディタ上で関数を選んで実行するか、`clasp run <関数名>` (要API有効化) で実行する。

## ファイル構成

- `appsscript.json` — マニフェスト。使用するAPIスコープを宣言する。
- `src/sheets.js` — スプレッドシートの読み書きサンプル。

## 次のステップ

`src/` に `docs.js` / `slides.js` / `gmail.js` / `drive.js` を追加し、`appsscript.json` の `oauthScopes` に対応するスコープを足していく。
