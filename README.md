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
- [ ] オブジェクトの同値判定対応
- [x] テスト
- [ ] 自動ビルド環境

## 対応する予定が無いもの

- proto2 シンタックス対応
- bytes, map, any 型の対応
- service 定義の対応
- 実行速度の最適化（速度が欲しいならちゃんと protobuf 入れましょう）

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

## C++

C++ 用のファイルを出力するには以下のように利用します。

```bash
protoc --jsonif-cpp_out=out/ test.proto
```

こうすると C++ ファイルが自動生成されて、以下のように型定義を使って JSON 文字列をシリアライズ・デシリアライズ可能になります。

```cpp
#include "test.json.h"

int main() {
  std::string str = R"(
{
  "people": [{
    "name": "hoge"
  }, {
    "name": "fuga"
  }]
}
  )";
  // JSON 文字列からのロード
  test::PersonList p = jsonif::from_json<test::PersonList>(str);

  std::cout << p[0].name << std::endl;
  // → hoge

  std::cout << p[1].name << std::endl;
  // → fuga

  // JSON 文字列へのシリアライズ
  std::cout << jsonif::to_json<test::PersonList>(p) << std::endl;
  // → {"people": [{"name": "hoge"}, {"name": "fuga" }]}
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

## Unity

Unity 用のファイルを出力するには以下のように利用します。

```bash
protoc --jsonif-unity_out=out/ test.proto
```

こうすると Unity 用の C# ファイルが自動生成されて、以下のように JSON 文字列をシリアライズ・デシリアライズ可能になります。

```cs
    void Start()
    {
        string json = "{\"people\":[{\"name\":\"hoge\"},{\"name\":\"fuga\"}]}";
        var p = JsonUtility.FromJson<Test.PersonList>(json);

        Debug.Log(p.people[0].name);
        // → hoge
        Debug.Log(p.people[1].name);
        // → fuga

        Debug.Log(JsonUtility.ToJson(p));
        // → {"people":[{"name":"hoge"},{"name":"fuga"}]}
    }
```

自動生成された `Test.cs` は以下のようになっています（若干変わっている可能性もあります）。
Unity では JsonUtility を利用しています。

```cs
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

## FAQ

### Q. jsonif って何？

A. JSON Interface の略です。

本当は protoc-gen-json という名前にしようと思ってたのですが、protoc-gen-json というのは[既に存在](https://github.com/sourcegraph/prototools/blob/master/README.json.md) していて、これは proto ファイルそのものを JSON 化するものです。
これは欲しい機能ではなかったし、単に protoc-gen-json というとこっちを指すんだとするとちょっと違うので、Interface を付けて別の名前にしました。

### Q. protoc-gen って何？

A. protoc で生成するプラグインの命名ルールです。

protoc は、デフォルトでは `--<NAME>_out=...` と指定したら `protoc-gen-<NAME>` プログラムを実行してファイル生成を呼び出す仕組みに[なっています](。
そのため protoc プラグインのリポジトリ名やバイナリ名に `protoc-gen-` プリフィックスを付けるのが一般的です。