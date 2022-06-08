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