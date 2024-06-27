# 変更履歴

- CHANGE
    - 下位互換のない変更
- UPDATE
    - 下位互換がある変更
- ADD
    - 下位互換がある追加
- FIX
    - バグ修正

## develop

- [ADD] TypeScript の生成を追加
    - @melpon

## 0.12.1 (2024-02-24)

- [FIX] C の oneof 値が extern になってなかったのを修正
    - @melpon

## 0.12.0 (2024-01-19)

- [ADD] optional フィールドに `has_<field>` 関数を追加
    - @melpon
- [ADD] oneof, optional フィールドに `clear_<field>` 関数を追加
    - @melpon

## 0.11.0 (2024-01-19)

- [CHANGE] oneof の enum 値を `<FIELD_NAME>_CASE_NOT_SET` から `NOT_SET` に変更
    - @melpon
- [ADD] proto3 の optional 仕様に対応
    - @melpon
- [FIX] １つのメッセージの中に oneof を２つ以上定義するとコンパイルエラーになるのを修正
    - @melpon

## 0.10.0 (2023-11-14)

- [CHANGE] enum class にするのをやめて enum で定義する
    - @melpon

## 0.9.1 (2023-11-13)

- [FIX] 複数のファイルから C ヘッダーをインクルードすると複数回定義のエラーが出ていたのを修正
    - @melpon

## 0.9.0 (2023-10-22)

- [ADD] 構造体のサイズを取得できる C API を追加
    - @melpon

## 0.8.2 (2023-10-19)

- [FIX] ヘッダーの include が足りてなかったのを修正
    - @melpon

## 0.8.1 (2023-10-05)

- [FIX] C のテストが全く通ってなかったのを修正
    - @melpon

## 0.8.0 (2023-10-02)

- [ADD] nlohmann::json に対応する
    - @melpon

## 0.7.2 (2023-05-21)

- [FIX] C++ 向けの関数を C リンケージにしていたのを修正
    - @melpon

## 0.7.1 (2023-05-05)

- [FIX] ちゃんとC言語のヘッダーになるように修正
    - @melpon

## 0.7.0 (2023-05-04)

- [UPDATE] 依存ライブラリを最新バージョンに更新する
    - @melpon
- [ADD] C 言語向けヘッダーの出力に対応
    - @melpon

## 0.6.0 (2022-06-08)

- [ADD] Darwin ARM64 向けバイナリを追加
    - @melpon

## 0.5.0 (2022-03-05)

- [ADD] protoc-gen-jsonif-cpp に no_serializer と no_deserializer カスタムオプションを追加
    - @melpon

## 0.4.2 (2022-03-04)

- [FIX] JSON のキー名が camelCase になってしまっていたのを修正
    - @melpon

## 0.4.1 (2022-03-04)

- [FIX] リリースパッケージの proto ディレクトリの配置が間違ってたのを修正
    - @melpon

## 0.4.0 (2022-03-04)

- [ADD] protoc-gen-jsonif-cpp に optimistic と discard_if_default カスタムオプションを追加
    - @melpon

## 0.3.0 (2022-02-28)

- [ADD] protoc-gen-jsonif-cpp に json_name フィールドオプションを追加
    - @melpon

## 0.2.0 (2022-02-28)

- [ADD] protoc-gen-jsonif-cpp を bytes 型に対応
    - @melpon

## 0.1.1 (2021-07-12)

- [ADD] zip のバイナリも追加
    - @melpon