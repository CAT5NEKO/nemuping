## NemuPing
ねむだる豆腐でPingを行う狂気のツールです。  
[SheeplaさんのPingソフト](https://github.com/sheepla/pingu)の実装方法を参考に通信の勉強を目的に作りました。

## 使い方
1.リポジトリをクローンします。  
2.`go build`でビルド  
3.`nemuPing 192.XXX.X.XXX`のように実行します。  

## オプション
```shell
nemuPing -c 4 192.XXX.X.XXX //指定回数のPingを行います(表示回数はint型で指定します)  
nemuPing -v //バージョンを表示します
nemuPing -p //管理者権限で実行します
```

### 参考資料
```
Copyright (c) 2022 sheepla
https://github.com/sheepla/pingu  閲覧日10/31/23
Copyright 2022 The Prometheus Authors
Copyright 2016 Cameron Sparr and contributors.
https://github.com/prometheus-community/pro-bing　　閲覧日10/31/23 
```

## 変更履歴

### 2023/10/31 V1.0
初版リリース  
このバージョンでは、アスキーアートは固定されていました。  
また、引数として表示数を明示しないとエラーを起こす問題を抱えていました。  

![image](https://github.com/CAT5NEKO/nemuping/assets/111590457/f7724159-3ee5-41ad-96d9-2bb1e217240b)


### 2024/04/04 V2.0
`.env`ファイルから自分のお好きなアスキーアートでPingを行うことができるようになりました。

![image](https://github.com/CAT5NEKO/nemuping/assets/111590457/d98df370-119e-4735-820c-48c2286081ff)


[設定方法]

1. `.env`ファイルを書き換え
2. `go build`で再度ビルド
3. `nemuPing`を実行
