# go_common_module

go言語開発用の共通module

## import方法

- importする側のgo.modのrequireに下記を追記
  - `github.com/tokane888/go_common_module v1.0.0`
    - `v1.0.0`の部分はgit tagの最新のバージョンを記載
- `go mod tidy`
  - 404エラーが出て当該versionが見つからないと言われた場合
    - 15分程待機した後で再実行
      - ダウンロード元のGo proxy(proxy.golang.org)のキャッシュが更新されていない可能性が高いため
