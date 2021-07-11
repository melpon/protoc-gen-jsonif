# protoc-gen-jsonif

proto ファイルから、JSON フォーマットでやりとりする型定義ファイルを出力する protoc プラグインです。

- proto ファイルで言語を越えて型定義が出来るのはとても良い
- しかし protobuf ライブラリを入れるのが面倒
- 今のプロジェクトには既に JSON ライブラリが入っているので JSON でやり取りしたい

という時に使うプラグインです。

## 実装状況

- [x] C++ 用コードの出力 (Boost.JSON 利用)
- [x] Unity 用コードの出力
- [x] message, enum 対応
- [x] repeated 対応
- [x] oneof 対応 
- [x] オブジェクトの等値判定対応
- [x] テスト
- [x] 自動ビルド環境

## 対応するかもしれないもの

- bytes 型の対応
- オブジェクトの大小の比較

## 対応する予定が無いもの

- proto2 シンタックス対応
- map, any 型の対応
- service 定義の対応
- 実行速度の最適化（速度が欲しいならちゃんと protobuf 入れましょう）

## 使い方

まず、[protobuf のリリース](https://github.com/protocolbuffers/protobuf/releases) から、自身のプラットフォームの最新のバイナリをダウンロードして下さい。
Windows なら `protoc-<version>-win64.zip`、macOS なら　`protoc-<version>-osx-x86_64.zip` などです。

ダウンロードが完了したら、これを展開し、`protoc/bin` ディレクトリに環境変数 `PATH` を通しておいて下さい。

次に、[protoc-gen-jsonif のリリース](https://github.com/melpon/protoc-gen-jsonif/releases) から、`protoc-gen-jsonif.tar.gz` をダウンロードして下さい。

ダウンロードが完了したら、これを展開し、自身のプラットフォームのディレクトリに環境変数 `PATH` を通しておいて下さい。
Windows なら `protoc-gen-jsonif/windows/amd64`、macOS なら `protoc-gen-jsonif/macos/amd64` などです。

次に以下の内容を `test.proto` として保存して下さい。

```proto
syntax = "proto3";

package test;

message Person {
  string name = 1;
}
```

あとは以下のように実行して `test.proto` ファイルを変換します。

```
mkdir -p out_cpp
protoc --jsonif-cpp_out=out_cpp/
```

これで `out_cpp/` ディレクトリに C++ 用のファイルが出力されます。

また、

```
mkdir -p out_unity
protoc --jsonif-unity_out=out_unity/
```

こうすると `out_unity/` ディレクトリに Unity 用のファイルが出力されます。

### PATH を通さずに実行する

環境変数 `PATH` を設定しなくても、以下のように指定すれば実行できます。

Windows 上で、ダウンロードした protoc が `./protoc` に、protoc-gen-jsonif が `./protoc-gen-jsonif` にあるという状態だとすると、

```
# C++
mkdir -p out_cpp
./protoc/bin/protoc.exe \
  --plugin=protoc-gen-jsonif-cpp=./protoc-gen-jsonif/windows/amd64/protoc-gen-jsonif-cpp.exe \
  --jsonif-cpp_out=out_cpp/ \
  test.proto

# Unity
mkdir -p out_unity
./protoc/bin/protoc.exe \
  --plugin=protoc-gen-jsonif-unity=./protoc-gen-jsonif/windows/amd64/protoc-gen-jsonif-unity.exe \
  --jsonif-unity_out=out_unity/ \
  test.proto
```

これで `PATH` を設定しなくても変換できます。

## 例

例では全て、以下のような `test.proto` ファイルがあるとしています。

```proto
syntax = "proto3";

package test;

message Person {
  string name = 1;
}
message PersonList {
  repeated Person people = 1;
}
```

### C++

C++ 用のファイルを出力するには以下のように利用します。

```bash
protoc --jsonif-cpp_out=out/ test.proto
```

こうすると C++ ファイルが自動生成されて、以下のように型定義を使って JSON 文字列をシリアライズ・デシリアライズ可能になります。

```cpp
#include "test.json.h"

int main() {
  test::PersonList p;

  p.people.push_back(test::Person{"hoge"});
  p.people.push_back(test::Person{"fuga"});

  // JSON 文字列への変換
  std::string str = jsonif::to_json(p);

  std::cout << str << std::endl;
  // → {"people":[{"name":"hoge"},{"name":"fuga"}]}

  // JSON 文字列から元に戻す
  p = jsonif::from_json<test::PersonList>(str);

  std::cout << p[0].name << std::endl;
  // → hoge

  std::cout << p[1].name << std::endl;
  // → fuga
}
```

自動生成された `test.json.h` は以下のようになっています（若干変わっている可能性もあります）。
C++ 標準には JSON ライブラリが無いため、現在は Boost.JSON を利用しています。

```cpp
#include <string>
#include <vector>
#include <stddef.h>

#include <boost/json.hpp>


namespace test {

struct Person {
  std::string name;
};

struct PersonList {
  std::vector<::test::Person> people;
};

// ::test::Person
void tag_invoke(const boost::json::value_from_tag&, boost::json::value& jv, const ::test::Person& v) {
  jv = {
    {"name", boost::json::value_from(v.name)},
  };
}

::test::Person tag_invoke(const boost::json::value_to_tag<::test::Person>&, const boost::json::value& jv) {
  ::test::Person v;
  v.name = boost::json::value_to<std::string>(jv.at("name"));
  return v;
}

// ::test::PersonList
void tag_invoke(const boost::json::value_from_tag&, boost::json::value& jv, const ::test::PersonList& v) {
  jv = {
    {"people", boost::json::value_from(v.people)},
  };
}

::test::PersonList tag_invoke(const boost::json::value_to_tag<::test::PersonList>&, const boost::json::value& jv) {
  ::test::PersonList v;
  v.people = boost::json::value_to<std::vector<::test::Person>>(jv.at("people"));
  return v;
}


}

#ifndef JSONIF_HELPER_DEFINED
#define JSONIF_HELPER_DEFINED

namespace jsonif {

template<class T>
inline T from_json(const std::string& s) {
  return boost::json::value_to<T>(boost::json::parse(s));
}

template<class T>
inline std::string to_json(const T& v) {
  return boost::json::serialize(boost::json::value_from(v));
}

}

#endif
```

### Unity

Unity 用のファイルを出力するには以下のように利用します。

```bash
protoc --jsonif-unity_out=out/ test.proto
```

こうすると Unity 用の C# ファイルが自動生成されて、以下のように JSON 文字列をシリアライズ・デシリアライズ可能になります。

```cs
    void Start()
    {
        var p = new Test.PersonList();

        p.people.Add(new Test.Person() { name = "hoge" });
        p.people.Add(new Test.Person() { name = "fuga" });

        // JSON 文字列への変換
        string str = Jsonif.Json.ToJson(p);

        Debug.Log(str);
        // → {"people":[{"name":"hoge"},{"name":"fuga"}]}

        // JSON 文字列から元に戻す
        p = Jsonif.Json.FromJson<Test.PersonList>(str);

        Debug.Log(p.people[0].name);
        // → hoge

        Debug.Log(p.people[1].name);
        // → fuga
    }
```

自動生成された `Test.cs` と `Jsonif.cs` は以下のようになっています（若干変わっている可能性もあります）。
Unity では内部的に JsonUtility を利用しています。

```cs
// Test.cs
namespace Test
{
    
    [System.Serializable]
    public class Person
    {
        public string name;
    }
    
    [System.Serializable]
    public class PersonList
    {
        public global::Test.Person[] people;
    }
    
}
```
```cs
// Jsonif.cs
using UnityEngine;

namespace Jsonif
{
    
    public static class Json
    {
        public static string ToJson<T>(T v)
        {
            return JsonUtility.ToJson(v);
        }
        public static T FromJson<T>(string s)
        {
            return JsonUtility.FromJson<T>(s);
        }
    }
    
}
```

## FAQ

### Q. jsonif って何？

A. JSON Interface の略です。

本当は protoc-gen-json という名前にしようと思ってたのですが、protoc-gen-json というのは[既に存在](https://github.com/sourcegraph/prototools/blob/master/README.json.md) していて、これは proto ファイルそのものを JSON 化するものです。
これは欲しい機能ではなかったし、単に protoc-gen-json というとこっちを指すんだとするとちょっと違うので、Interface を付けて別の名前にしました。

### Q. protoc-gen って何？

A. protoc で生成するプラグインの命名ルールです。

protoc は、デフォルトでは `--<NAME>_out=...` と指定したら `protoc-gen-<NAME>` プログラムを実行してファイル生成を呼び出す仕組みになっています（`--plugin` オプションで上書きできます）。
そのため protoc プラグインのリポジトリ名やバイナリ名に `protoc-gen-` プリフィックスを付けるのが一般的です。