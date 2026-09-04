# Google Apps Script (GAS) トラック

Google純正のスクリプト実行環境。認証設定なしで `SpreadsheetApp` / `DocumentApp` / `SlidesApp` / `GmailApp` / `DriveApp` が使える。
ローカルでコードを書いて `clasp` でGoogle側にpush/pullする。

## セットアップ

```powershell
npm install -g @google/clasp
clasp login
```

ブラウザでGoogleアカウントにログインし、clasp用のOAuth許可を行う。

### 新規プロジェクトの作成

```powershell
cd gas
clasp create --type standalone --title "google-scripts-study"
```

実行すると `.clasp.json` (scriptIdを含む、gitignore対象) が生成される。
既存のApps Scriptプロジェクトに紐付けたい場合は `clasp clone <scriptId>` を使う。

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
