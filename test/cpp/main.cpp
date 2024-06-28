#include <iostream>
#include <cassert>
#if defined(JSONIF_USE_NLOHMANN_JSON)
#else
#include <boost/json/src.hpp>
#endif

#include "empty.json.h"
#include "message.json.h"
#include "enumpb.json.h"
#include "nested.json.h"
#include "repeated.json.h"
#include "oneof.json.h"
#include "optional.json.h"
#include "importing.json.h"
#include "bytes.json.h"
#include "jsonfield.json.h"
#include "optimistic.json.h"
#include "discard_if_default.json.h"
#include "no_serializer.json.h"

template<class T>
T identify(T v) {
  auto vs = jsonif::to_json(v);
  auto r = jsonif::from_json<T>(vs);
  auto rs = jsonif::to_json(v);
  assert(r == v);
  assert(rs == vs);
  return r;
}

void test_empty() {
  empty::Test a;
  a = identify(a);
}

void test_message() {
  message::Person a;
  assert(a.name == "");
  assert(a.flag == false);
  a = identify(a);
  assert(a.name == "");
  assert(a.flag == false);

  a.name = "foo";
  a.flag = true;
  a = identify(a);
  assert(a.name == "foo");
  assert(a.flag == true);
}

void test_enumpb() {
  enumpb::Data a = enumpb::FOO;
  a = identify(a);
  assert(a == enumpb::FOO);

  a = enumpb::Data::BAR;
  a = identify(a);
  assert(a == enumpb::BAR);
}

void test_nested() {
  nested::nested::Test2 a;
  assert(a.nested_message.name == "");
  assert(a.nested_enum == nested::nested::Test::FOO);
  assert(a.test.nested_message.name == "");
  assert(a.test.nested_enum == nested::nested::Test::FOO);
  a = identify(a);
  assert(a.nested_message.name == "");
  assert(a.nested_enum == nested::nested::Test::FOO);
  assert(a.test.nested_message.name == "");
  assert(a.test.nested_enum == nested::nested::Test::FOO);

  a.nested_message.name = "foo";
  a.nested_enum = nested::nested::Test::BAR;
  a.test.nested_message.name = "bar";
  a.test.nested_enum = nested::nested::Test::HOGE;
  a = identify(a);
  assert(a.nested_message.name == "foo");
  assert(a.nested_enum == nested::nested::Test::BAR);
  assert(a.test.nested_message.name == "bar");
  assert(a.test.nested_enum == nested::nested::Test::HOGE);
}

void test_repeated() {
  repeated::Test a;
  assert(a.a.empty());
  assert(a.b.empty());
  assert(a.c.empty());
  assert(a.d.empty());
  a = identify(a);
  assert(a.a.empty());
  assert(a.b.empty());
  assert(a.c.empty());
  assert(a.d.empty());

  a.a.push_back(1);
  a.b.push_back("foo");
  a.c.push_back(repeated::BAR);
  a.d.push_back(repeated::Message{"bar"});
  a = identify(a);
  assert(a.a.size() == 1 && a.a.at(0) == 1);
  assert(a.b.size() == 1 && a.b.at(0) == "foo");
  assert(a.c.size() == 1 && a.c.at(0) == repeated::BAR);
  assert(a.d.size() == 1 && a.d.at(0).name == "bar");
}

void test_oneof() {
  oneof::Test a;
  assert(a.test_oneof_case == oneof::Test::TestOneofCase::NOT_SET);
  a = identify(a);
  assert(a.test_oneof_case == oneof::Test::TestOneofCase::NOT_SET);

  a.set_a(1);
  assert(a.test_oneof_case == oneof::Test::TestOneofCase::kA);
  assert(a.a == 1);
  a = identify(a);
  assert(a.test_oneof_case == oneof::Test::TestOneofCase::kA);
  assert(a.a == 1);

  a.set_b("foo");
  assert(a.test_oneof_case == oneof::Test::TestOneofCase::kB);
  assert(a.b == "foo");
  a = identify(a);
  assert(a.test_oneof_case == oneof::Test::TestOneofCase::kB);
  assert(a.b == "foo");

  a.set_c(oneof::BAR);
  assert(a.test_oneof_case == oneof::Test::TestOneofCase::kC);
  assert(a.c == oneof::BAR);
  a = identify(a);
  assert(a.test_oneof_case == oneof::Test::TestOneofCase::kC);
  assert(a.c == oneof::BAR);

  a.set_d(oneof::Message{"bar"});
  assert(a.test_oneof_case == oneof::Test::TestOneofCase::kD);
  assert(a.d.name == "bar");
  a = identify(a);
  assert(a.test_oneof_case == oneof::Test::TestOneofCase::kD);
  assert(a.d.name == "bar");

  a.clear_c();
  assert(a.d.name == "bar");
  assert(a.test_oneof_case == oneof::Test::TestOneofCase::kD);
  a.clear_d();
  assert(a.test_oneof_case == oneof::Test::TestOneofCase::NOT_SET);

  a.set_a(10);
  a.clear_test_oneof_case();
  assert(a.test_oneof_case == oneof::Test::TestOneofCase::NOT_SET);
}

void test_optional() {
  optional::Test a;
  assert(!a.has_a());
  assert(!a.has_b());
  assert(!a.has_c());
  assert(!a.has_d());
  a = identify(a);
  assert(!a.has_a());
  assert(!a.has_b());
  assert(!a.has_c());
  assert(!a.has_d());

  a.set_a(1);
  assert(a.has_a());
  assert(a.a == 1);
  a = identify(a);
  assert(a.has_a());
  assert(a.a == 1);

  a.set_b("foo");
  assert(a.has_b());
  assert(a.b == "foo");
  a = identify(a);
  assert(a.has_b());
  assert(a.b == "foo");

  a.set_c(optional::BAR);
  assert(a.has_c());
  assert(a.c == optional::BAR);
  a = identify(a);
  assert(a.has_c());
  assert(a.c == optional::BAR);

  a.set_d(optional::Message{"bar"});
  assert(a.has_d());
  assert(a.d.name == "bar");
  a = identify(a);
  assert(a.has_d());
  assert(a.d.name == "bar");

  a.clear_a();
  assert(!a.has_a());
  a.clear_b();
  assert(!a.has_b());
  a.clear_c();
  assert(!a.has_c());
  a.clear_d();
  assert(!a.has_d());
}

void test_importing() {
  importing::Test a;
  assert(a.t.nanos == 0);
  a = identify(a);
  assert(a.t.nanos == 0);
}

void test_bytes() {
  std::string v("\x00\x01\x02\x03", 4);
  std::string v2 = u8"あいうえお";
  bytes::Test a;
  a.data = v;
  a.rp_data.push_back(v);
  a.rp_data.push_back(v2);
  a = identify(a);
  assert(a.data == v);
  assert(a.rp_data.at(0) == v);
  assert(a.rp_data.at(1) == v2);
}

void test_jsonfield() {
  jsonfield::Test a;
  a.field = 10;
  auto str = jsonif::to_json(a);
  assert(str == R"({"test":10,"hoge_field":0})" || str == R"({"hoge_field":0,"test":10})");
  a = identify(a);
  assert(a.field == 10);
}

void test_optimistic() {
  auto str = R"({"b":"hoge"})";
  auto r = jsonif::from_json<optimistic::Test>(str);
  assert(r.a == "");
  assert(r.b == "hoge");
  r = identify(r);
  assert(r.a == "");
  assert(r.b == "hoge");
}

void test_discard_if_default() {
  discard_if_default::Test a;
  auto str = jsonif::to_json(a);
  assert(str == R"({"a":""})");

  a.b = "hoge";
  str = jsonif::to_json(a);
  assert(str == R"({"a":"","b":"hoge"})" || str == R"({"b":"hoge","a":""})");

  a.b = "";
  a.c.a = 10;
  str = jsonif::to_json(a);
  assert(str == R"({"a":"","c":{"a":10}})" || str == R"({"c":{"a":10},"a":""})");
}

namespace no_serializer {

#if defined(JSONIF_USE_NLOHMANN_JSON)

static void to_json(nlohmann::json& jv, const ::no_serializer::Test& v) {
  using nlohmann::to_json;
  to_json(jv["b"], v.a);
}

static void from_json(const nlohmann::json& jv, ::no_serializer::Test& v) {
  using nlohmann::from_json;
  from_json(jv.at("b"), v.a);
}

#else

static void tag_invoke(const boost::json::value_from_tag&, boost::json::value& jv, const ::no_serializer::Test& v) {
  boost::json::object obj;
  obj["b"] = boost::json::value_from(v.a);
  jv = std::move(obj);
}

static ::no_serializer::Test tag_invoke(const boost::json::value_to_tag<::no_serializer::Test>&, const boost::json::value& jv) {
  ::no_serializer::Test v;
  v.a = boost::json::value_to<std::string>(jv.at("b"));
  return v;
}

#endif

}

void test_no_serializer() {
  no_serializer::Test a;
  a.a = "hoge";
  auto str = jsonif::to_json(a);
  assert(str == R"({"b":"hoge"})");
  a = identify(a);
  assert(a.a == "hoge");
}

int main() {
  test_empty();
  test_message();
  test_enumpb();
  test_nested();
  test_repeated();
  test_oneof();
  test_optional();
  test_importing();
  test_bytes();
  test_jsonfield();
  test_optimistic();
  test_discard_if_default();
  test_no_serializer();

  std::cout << "C++ Test passed" << std::endl;
}