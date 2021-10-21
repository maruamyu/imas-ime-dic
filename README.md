maruamyu/imas-ime-dic
=====================

[THE IDOLM@STER (アイドルマスター)](http://idolmaster.jp/)にかかわる単語を記述した日本語IME向けの辞書ファイルです。

[アイマスDB](https://imas-db.jp/)が管理を行っています。

## 想定環境

- Windows
	- Microsoft IME
	- メモ帳で編集可能
- Android
	- Google日本語入力
	- 主要なテキストエディタで編集可能
- 以下の文字が利用可能
	- 﨑 … 「赤﨑千夏」など
	- ♥ … 「私はアイドル♥」など
	- ♡ … 「仲良しでいようね♡」など
	- Ø … 「ØωØver!!」
	- ➚ … 「JOKER➚オールマイティ」
	- è … 「Cafè Parade」
	- 俠 … 「俠気乱舞」
	- ✿ … 「花ざかりWeekend✿」

## 取り決め

上記想定環境から、以下のように決めます。

- 文字コードは UTF-16 LE
	- Windowsのメモ帳で扱えるようにするためBOM付きにする
- 改行コードは CRLF

## 他形式へのコンバート

リポジトリ内にある *convert_dic.go* を[Go言語](https://golang.org/)でコンパイルして実行すると
*dist/* ディレクトリ以下にファイルが生成されます。

- *gboard.zip* : Gboard(Android版)の単語リストにインポートするためのファイル
- *macosx.plist* : Mac OS Xの「キーボード」→「ユーザー辞書」にドラッグ＆ドロップで登録するためのファイル
- *skk-jisyo.imas.utf8* : SKK辞書ファイル (AquaSKKで動作確認済)

```bash
go get
go build convert_dic.go
./convert_dic
```

2021-10-22 から生成したファイルをコミットするようにしました。

## License

リポジトリ内のテキストファイルは、MITライセンス下で配布されます。

This repository is released under the MIT License, see [LICENSE](LICENSE)

各社の商標または登録商標が含まれる場合がありますが、営利利用を意図したものではありません。

## コントリビューション

歓迎します。
forkして、新規branchを作成して、pullリクエストしてください。
