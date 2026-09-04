/**
 * スプレッドシートを新規作成し、データを書き込んでから読み戻すサンプル。
 * スクリプトエディタで createAndReadSheet を実行する。
 */
function createAndReadSheet() {
  const ss = SpreadsheetApp.create('google-scripts-study');
  const sheet = ss.getActiveSheet();

  const rows = [
    ['name', 'score'],
    ['Alice', 90],
    ['Bob', 85],
  ];
  sheet.getRange(1, 1, rows.length, rows[0].length).setValues(rows);

  const values = sheet.getDataRange().getValues();
  Logger.log(values);
  Logger.log('Spreadsheet URL: %s', ss.getUrl());
}

/**
 * 既存のスプレッドシートを開いて読み書きするサンプル。
 * spreadsheetId はスプレッドシートのURLに含まれるIDを指定する。
 */
function readExistingSheet(spreadsheetId) {
  const ss = SpreadsheetApp.openById(spreadsheetId);
  const sheet = ss.getSheets()[0];
  const values = sheet.getDataRange().getValues();
  Logger.log(values);
  return values;
}
